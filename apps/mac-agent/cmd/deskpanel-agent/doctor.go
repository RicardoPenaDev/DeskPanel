package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// cmdDoctor checks the parts of the environment that don't depend on
// config.json yet (PROJECT.md §13): required macOS commands are on PATH.
// Config/permission/port checks are added once those pieces exist
// (Fase 1+).
func cmdDoctor(args []string) error {
	fmt.Printf("SO: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != "darwin" {
		fmt.Println("aviso: deskpanel-agent deve rodar no macOS; este binário está sendo executado em", runtime.GOOS)
	}

	requiredBins := []string{"open", "shortcuts", "osascript", "pmset"}
	allOK := true
	for _, bin := range requiredBins {
		if path, err := exec.LookPath(bin); err == nil {
			fmt.Printf("  [ok]     %-10s %s\n", bin, path)
		} else {
			allOK = false
			fmt.Printf("  [faltando] %-10s não encontrado no PATH\n", bin)
		}
	}

	fmt.Println("\nchecagens de config.json, permissões e porta: ainda não implementadas (Fase 1).")

	if !allOK && runtime.GOOS == "darwin" {
		return fmt.Errorf("um ou mais comandos macOS obrigatórios não foram encontrados")
	}
	return nil
}
