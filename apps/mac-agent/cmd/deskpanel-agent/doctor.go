package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"deskpanel-agent/internal/config"
)

// cmdDoctor verifica o que dá pra verificar sem depender de um servidor
// rodando (PROJECT.md §13): comandos do macOS no PATH, e agora também a
// existência/validade de config.json e a permissão do diretório de dados.
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

	fmt.Println("\nchecagens de porta e conexões: ainda não implementadas (Fase 2 — servidor real).")

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
