package main

import (
	"flag"
	"fmt"

	"deskpanel-agent/internal/config"
	"deskpanel-agent/internal/protocol"
)

// notImplemented reports the phase (from PROJECT.md §17) a piece of
// behavior is planned for, instead of pretending to do something it can't
// do yet.
func notImplemented(command, phase string) error {
	return fmt.Errorf("%s: ainda não implementado — ver %s em PROJECT.md", command, phase)
}

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

	fmt.Printf("config carregada de %s\n", *configPath)
	fmt.Printf("  porta configurada: %d\n", cfg.Server.Port)
	fmt.Printf("  ações registradas: %d\n", len(cfg.Actions))
	fmt.Println("\nserve: API HTTP e WebSocket ainda não implementados — ver Fase 2 em PROJECT.md")
	return nil
}

func cmdPair(args []string) error {
	return notImplemented("pair", "Fase 2 (API, WebSocket e segurança)")
}

func cmdDevices(args []string) error {
	return notImplemented("devices", "Fase 2 (API, WebSocket e segurança)")
}

func cmdRevoke(args []string) error {
	return notImplemented("revoke", "Fase 2 (API, WebSocket e segurança)")
}

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfigPath(), "caminho para config.json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("status: %w (rode 'deskpanel-agent doctor' para diagnóstico)", err)
	}

	fmt.Printf("config:   %s\n", *configPath)
	fmt.Printf("porta:    %d\n", cfg.Server.Port)
	fmt.Printf("ações:    %d\n", len(cfg.Actions))
	fmt.Println("servidor: não está rodando (serve real é Fase 2)")
	fmt.Println("conexões: n/d (Fase 2)")
	return nil
}

func cmdVersion() {
	fmt.Printf("deskpanel-agent %s (protocolo v%d)\n", version, protocol.Version)
}
