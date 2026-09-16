package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"deskpanel-agent/internal/adminsocket"
	"deskpanel-agent/internal/api"
	"deskpanel-agent/internal/appscan"
	"deskpanel-agent/internal/config"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/executor"
	"deskpanel-agent/internal/logging"
	"deskpanel-agent/internal/pairing"
	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/ratelimit"
	"deskpanel-agent/internal/weather"
	"deskpanel-agent/internal/websocket"
)

func cmdVersion() {
	fmt.Printf("deskpanel-agent %s (protocolo v%d)\n", version, protocol.Version)
}

// cmdServe é o processo de longa duração: carrega a config, sobe a API
// HTTP+WebSocket e o socket administrativo, e fica rodando até receber
// SIGINT/SIGTERM (PROJECT.md §17 Fase 2).
func cmdServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfigPath(), "caminho para config.json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	store, err := devices.LoadStore(defaultDevicesPath())
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	logger, logCloser, logErr := openServeLogger()
	if logCloser != nil {
		defer logCloser.Close()
	}
	if logErr != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", logErr)
	}
	pairingMgr := pairing.NewManager()
	limiter := ratelimit.NewLimiter(cfg.Security.RequestsPerMinute)
	failures := ratelimit.NewFailureTracker(cfg.Security.FailedAuthLimit, 5*time.Minute)

	macName, err := os.Hostname()
	if err != nil || macName == "" {
		macName = "Mac"
	}

	appScanner := appscan.New()
	weatherProvider := weather.New()

	wsHandler := &websocket.Handler{
		Devices:      store,
		Actions:      cfg.Actions,
		Apps:         appScanner,
		Executor:     &executor.MacOSExecutor{},
		MacName:      macName,
		AgentVersion: version,
		Logger:       logger,
	}

	apiServer := &api.Server{
		Actions:      cfg.Actions,
		Apps:         appScanner,
		Weather:      weatherProvider,
		Devices:      store,
		Pairing:      pairingMgr,
		Limiter:      limiter,
		Failures:     failures,
		AgentVersion: version,
		MacName:      macName,
		WSHandler:    wsHandler,
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.ListenAddress, cfg.Server.Port),
		Handler: apiServer.Handler(),
	}

	adminSrv := &adminsocket.Server{
		SocketPath: defaultAdminSocketPath(),
		Handlers:   adminHandlers(cfg, store, pairingMgr, wsHandler),
	}
	adminLn, err := adminSrv.Listen()
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	defer os.Remove(defaultAdminSocketPath())
	go adminSrv.Serve(adminLn)

	fmt.Printf("DeskPanel Agent ouvindo em %s (protocolo v%d)\n", httpServer.Addr, protocol.Version)
	fmt.Printf("%d ações registradas, %d dispositivos pareados\n", len(cfg.Actions), len(store.List()))
	fmt.Printf("socket administrativo: %s\n", defaultAdminSocketPath())

	errCh := make(chan error, 1)
	go func() { errCh <- httpServer.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve: %w", err)
		}
	case <-sigCh:
		fmt.Println("\nencerrando...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
		_ = adminLn.Close()
	}
	return nil
}

func cmdPair(args []string) error {
	result, err := adminsocket.Call(defaultAdminSocketPath(), "pair", nil)
	if err != nil {
		return fmt.Errorf("pair: %w", err)
	}
	var res pairResult
	if err := json.Unmarshal(result, &res); err != nil {
		return fmt.Errorf("pair: resposta inesperada do agente: %w", err)
	}
	fmt.Printf("DeskPanel pairing enabled for %d minutes.\n", res.ExpiresInSeconds/60)
	fmt.Printf("Code: %s\n", res.Code)
	return nil
}

func cmdDevices(args []string) error {
	result, err := adminsocket.Call(defaultAdminSocketPath(), "devices", nil)
	if err != nil {
		return fmt.Errorf("devices: %w", err)
	}
	var list []deviceSummary
	if err := json.Unmarshal(result, &list); err != nil {
		return fmt.Errorf("devices: resposta inesperada do agente: %w", err)
	}
	if len(list) == 0 {
		fmt.Println("nenhum dispositivo pareado")
		return nil
	}
	for _, d := range list {
		status := "ativo"
		if d.Revoked {
			status = "revogado"
		}
		fmt.Printf("%-36s  %-20s  %-8s  pareado em %s\n", d.DeviceID, d.DeviceName, status, d.PairedAt.Format(time.RFC3339))
	}
	return nil
}

func cmdRevoke(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("revoke: uso: deskpanel-agent revoke <device-id>")
	}
	if _, err := adminsocket.Call(defaultAdminSocketPath(), "revoke", map[string]string{"deviceId": args[0]}); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	fmt.Printf("dispositivo %s revogado\n", args[0])
	return nil
}

func cmdStatus(args []string) error {
	result, err := adminsocket.Call(defaultAdminSocketPath(), "status", nil)
	if err != nil {
		return fmt.Errorf("status: agente não parece estar rodando — inicie com 'deskpanel-agent serve' (%w)", err)
	}
	var res statusResult
	if err := json.Unmarshal(result, &res); err != nil {
		return fmt.Errorf("status: resposta inesperada do agente: %w", err)
	}
	fmt.Printf("porta:                  %d\n", res.Port)
	fmt.Printf("ações registradas:      %d\n", res.ActionsCount)
	fmt.Printf("dispositivos pareados:  %d\n", res.Devices)
	fmt.Printf("conexões ativas:        %d\n", res.Connections)
	return nil
}

// openServeLogger abre o logger estruturado do agente em arquivo rotativo
// (PROJECT.md §13). Se o diretório de logs não puder ser criado ou o
// arquivo não puder ser aberto, cai para stdout — uma falha de log nunca
// deve impedir o agente de subir.
func openServeLogger() (*slog.Logger, io.Closer, error) {
	path := defaultLogPath()
	if path == "" {
		return logging.New(), nil, fmt.Errorf("logging: não foi possível determinar o diretório home; usando stdout")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return logging.New(), nil, fmt.Errorf("logging: não foi possível criar %s: %w", filepath.Dir(path), err)
	}
	logger, closer, err := logging.NewFile(path)
	if err != nil {
		return logger, closer, err
	}
	return logger, closer, nil
}
