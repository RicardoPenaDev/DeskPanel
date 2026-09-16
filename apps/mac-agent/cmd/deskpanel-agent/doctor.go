package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"

	"deskpanel-agent/internal/adminsocket"
	"deskpanel-agent/internal/config"
	"deskpanel-agent/internal/netguard"
	"deskpanel-agent/internal/protocol"
)

const launchAgentLabel = "dev.ricardopena.deskpanel.agent"

// cmdDoctor verifica comandos do macOS no PATH, a existência/validade de
// config.json, a permissão do diretório de dados, a porta do servidor, o
// status do LaunchAgent, conectividade local e a versão do protocolo
// (PROJECT.md §13).
func cmdDoctor(args []string) error {
	fmt.Printf("SO: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != "darwin" {
		fmt.Println("aviso: deskpanel-agent deve rodar no macOS; este binário está sendo executado em", runtime.GOOS)
	}

	requiredBins := []string{"open", "shortcuts", "osascript", "pmset"}
	binsOK := true
	for _, bin := range requiredBins {
		if path, err := exec.LookPath(bin); err == nil {
			fmt.Printf("  [ok]       %-10s %s\n", bin, path)
		} else {
			binsOK = false
			fmt.Printf("  [faltando] %-10s não encontrado no PATH\n", bin)
		}
	}

	fmt.Println()
	cfg, configOK := checkConfig()

	fmt.Println()
	agentRunning := checkRunningAgent()

	fmt.Println()
	portOK := true
	if cfg != nil {
		portOK = checkPort(cfg, agentRunning)
	} else {
		fmt.Println("  [info]     checagem de porta pulada — config.json inválido")
	}

	fmt.Println()
	checkLaunchAgent()

	fmt.Println()
	checkLocalConnectivity()

	fmt.Println()
	fmt.Printf("  [info]     versão do protocolo: v%d\n", protocol.Version)
	fmt.Printf("  [info]     versão do agente:    %s\n", version)

	if runtime.GOOS == "darwin" && (!binsOK || !configOK || !portOK) {
		return fmt.Errorf("uma ou mais checagens falharam — veja acima")
	}
	return nil
}

// checkConfig valida config.json e retorna a config carregada (nil se
// inválida) para que outras checagens (ex.: porta) possam reaproveitá-la.
func checkConfig() (*config.Config, bool) {
	path := defaultConfigPath()
	if path == "" {
		fmt.Println("  [erro]     não foi possível determinar o diretório home do usuário")
		return nil, false
	}

	dir := defaultAppDir()
	if info, err := os.Stat(dir); err != nil {
		fmt.Printf("  [faltando] diretório de dados não existe ainda: %s\n", dir)
		fmt.Printf("             (copie configs/config.example.json para %s para começar)\n", path)
		return nil, false
	} else if info.Mode().Perm()&0o077 != 0 {
		fmt.Printf("  [aviso]    diretório de dados com permissão %o — esperado 0700: %s\n", info.Mode().Perm(), dir)
	} else {
		fmt.Printf("  [ok]       diretório de dados: %s (permissão %o)\n", dir, info.Mode().Perm())
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Printf("  [erro]     config.json inválido ou ausente: %v\n", err)
		return nil, false
	}
	fmt.Printf("  [ok]       config.json válido — %d ações registradas (sem IDs duplicados ou inválidos)\n", len(cfg.Actions))
	return cfg, true
}

// checkRunningAgent tenta falar com um `serve` já em execução pelo socket
// administrativo. Não é um erro o agente estar parado — doctor só informa.
// Retorna true se um agente respondeu.
func checkRunningAgent() bool {
	result, err := adminsocket.Call(defaultAdminSocketPath(), "status", nil)
	if err != nil {
		fmt.Println("  [info]     nenhum 'deskpanel-agent serve' rodando no momento")
		return false
	}
	var res statusResult
	if err := json.Unmarshal(result, &res); err != nil {
		fmt.Println("  [aviso]    agente rodando, mas resposta de status inesperada")
		return true
	}
	fmt.Printf("  [ok]       agente rodando — porta %d, %d ações, %d conexões ativas\n", res.Port, res.ActionsCount, res.Connections)
	return true
}

// checkPort tenta abrir a porta configurada. Se já estiver em uso, distingue
// entre "é o próprio agente" (agentRunning, via checkRunningAgent) e
// "outro processo está usando essa porta".
func checkPort(cfg *config.Config, agentRunning bool) bool {
	addr := fmt.Sprintf("%s:%d", cfg.Server.ListenAddress, cfg.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		_ = ln.Close()
		fmt.Printf("  [ok]       porta %d livre\n", cfg.Server.Port)
		return true
	}
	if agentRunning {
		fmt.Printf("  [ok]       porta %d em uso pelo próprio deskpanel-agent\n", cfg.Server.Port)
		return true
	}
	fmt.Printf("  [erro]     porta %d ocupada por outro processo: %v\n", cfg.Server.Port, err)
	return false
}

// checkLaunchAgent reporta se o LaunchAgent está carregado no domínio do
// usuário atual (só faz sentido no macOS; em outros sistemas só informa).
func checkLaunchAgent() {
	if runtime.GOOS != "darwin" {
		fmt.Println("  [info]     checagem de LaunchAgent pulada (não é macOS)")
		return
	}
	uid := os.Getuid()
	target := fmt.Sprintf("gui/%d/%s", uid, launchAgentLabel)
	cmd := exec.Command("launchctl", "print", target)
	if err := cmd.Run(); err != nil {
		fmt.Printf("  [faltando] LaunchAgent %s não está carregado (rode scripts/install-macos.sh)\n", launchAgentLabel)
		return
	}
	fmt.Printf("  [ok]       LaunchAgent %s carregado (gui/%d)\n", launchAgentLabel, uid)
}

// checkLocalConnectivity lista os endereços IPv4 privados das interfaces de
// rede da máquina — o que o Android precisa digitar na tela de pareamento.
func checkLocalConnectivity() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("  [erro]     não foi possível listar interfaces de rede: %v\n", err)
		return
	}

	found := false
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip4 := ipNet.IP.To4()
		if ip4 == nil || ip4.IsLoopback() {
			continue
		}
		if !netguard.IsPrivateOrLoopback(ip4) {
			continue
		}
		fmt.Printf("  [ok]       endereço local alcançável: %s\n", ip4.String())
		found = true
	}
	if !found {
		fmt.Println("  [aviso]    nenhum endereço IPv4 de rede local encontrado — confira o Wi-Fi/Ethernet")
	}
}
