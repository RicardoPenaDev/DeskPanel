package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"deskpanel-agent/internal/adminsocket"
	"deskpanel-agent/internal/config"
)

// cmdDoctor verifica comandos do macOS no PATH, a existência/validade de
// config.json, a permissão do diretório de dados, e se um `serve` já está
// rodando (via o socket administrativo) — PROJECT.md §13.
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
	configOK := checkConfig()

	fmt.Println()
	checkRunningAgent()

	if runtime.GOOS == "darwin" && (!binsOK || !configOK) {
		return fmt.Errorf("uma ou mais checagens falharam — veja acima")
	}
	return nil
}

func checkConfig() bool {
	path := defaultConfigPath()
	if path == "" {
		fmt.Println("  [erro]     não foi possível determinar o diretório home do usuário")
		return false
	}

	dir := defaultAppDir()
	if info, err := os.Stat(dir); err != nil {
		fmt.Printf("  [faltando] diretório de dados não existe ainda: %s\n", dir)
		fmt.Printf("             (copie configs/config.example.json para %s para começar)\n", path)
		return false
	} else if info.Mode().Perm()&0o077 != 0 {
		fmt.Printf("  [aviso]    diretório de dados com permissão %o — esperado 0700: %s\n", info.Mode().Perm(), dir)
	} else {
		fmt.Printf("  [ok]       diretório de dados: %s (permissão %o)\n", dir, info.Mode().Perm())
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Printf("  [erro]     config.json inválido ou ausente: %v\n", err)
		return false
	}
	fmt.Printf("  [ok]       config.json válido — %d ações registradas\n", len(cfg.Actions))
	return true
}

// checkRunningAgent tenta falar com um `serve` já em execução pelo socket
// administrativo. Não é um erro o agente estar parado — doctor só informa.
func checkRunningAgent() {
	result, err := adminsocket.Call(defaultAdminSocketPath(), "status", nil)
	if err != nil {
		fmt.Println("  [info]     nenhum 'deskpanel-agent serve' rodando no momento")
		return
	}
	var res statusResult
	if err := json.Unmarshal(result, &res); err != nil {
		fmt.Println("  [aviso]    agente rodando, mas resposta de status inesperada")
		return
	}
	fmt.Printf("  [ok]       agente rodando — porta %d, %d ações, %d conexões ativas\n", res.Port, res.ActionsCount, res.Connections)
}
