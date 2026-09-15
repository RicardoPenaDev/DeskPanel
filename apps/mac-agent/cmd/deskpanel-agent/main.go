// Command deskpanel-agent is the DeskPanel Agent for macOS (PROJECT.md §7).
// It offers subcommands for serving the API/WebSocket, pairing devices and
// diagnosing the local setup. There is no subcommand or flag that accepts
// an arbitrary shell command from the caller.
package main

import (
	"fmt"
	"os"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "serve":
		err = cmdServe(args)
	case "pair":
		err = cmdPair(args)
	case "devices":
		err = cmdDevices(args)
	case "revoke":
		err = cmdRevoke(args)
	case "status":
		err = cmdStatus(args)
	case "doctor":
		err = cmdDoctor(args)
	case "version":
		cmdVersion()
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "comando desconhecido: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`deskpanel-agent <comando>

Comandos:
  serve     inicia API, WebSocket e execução de ações
  pair      abre uma janela de pareamento por 5 minutos
  devices   lista dispositivos pareados (sem mostrar tokens)
  revoke    remove a autorização de um dispositivo
  status    informa porta, estado e conexões
  doctor    verifica configuração, permissões, arquivos, porta e comandos do macOS
  version   mostra a versão do agente e do protocolo`)
}
