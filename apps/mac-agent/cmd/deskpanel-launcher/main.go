package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	agent := filepath.Join(home, "Library", "Application Support", "DeskPanel", "bin", "deskpanel-agent")
	if _, err := os.Stat(agent); err != nil {
		return
	}

	cmd := exec.Command(agent, "setup")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Assistente local aberto em ") {
			url := strings.TrimPrefix(line, "Assistente local aberto em ")
			_ = exec.Command("open", url).Run()
			break
		}
	}
	_ = cmd.Wait()
}
