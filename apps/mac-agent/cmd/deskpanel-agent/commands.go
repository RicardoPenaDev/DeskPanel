package main

import (
	"fmt"

	"deskpanel-agent/internal/protocol"
)

// notImplemented reports the phase (from PROJECT.md §17) a command is
// planned for, instead of pretending to do something it can't do yet.
func notImplemented(command, phase string) error {
	return fmt.Errorf("%s: ainda não implementado — ver %s em PROJECT.md", command, phase)
}

func cmdServe(args []string) error {
	return notImplemented("serve", "Fase 2 (API, WebSocket e segurança)")
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
	return notImplemented("status", "Fase 2 (API, WebSocket e segurança)")
}

func cmdVersion() {
	fmt.Printf("deskpanel-agent %s (protocolo v%d)\n", version, protocol.Version)
}
