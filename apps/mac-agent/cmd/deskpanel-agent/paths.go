package main

import (
	"os"
	"path/filepath"
)

// defaultConfigPath returns ~/Library/Application Support/DeskPanel/config.json
// (PROJECT.md §7.2). Retorna "" se o diretório home não puder ser resolvido.
func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "DeskPanel", "config.json")
}

// defaultAppDir retorna o diretório de dados do agente (mesmo pai de
// defaultConfigPath), usado pelo doctor para checar permissões (§12.1).
func defaultAppDir() string {
	path := defaultConfigPath()
	if path == "" {
		return ""
	}
	return filepath.Dir(path)
}
