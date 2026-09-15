package pairing

import "errors"

// Erros de Manager.Attempt — mapeiam 1:1 para os códigos de erro do
// protocolo (docs/PROTOCOL.md): PAIRING_CLOSED, PAIRING_CODE_EXPIRED e
// PAIRING_CODE_INVALID.
var (
	ErrClosed      = errors.New("pairing: nenhuma janela de pareamento ativa")
	ErrExpired     = errors.New("pairing: código expirado")
	ErrInvalidCode = errors.New("pairing: código inválido")
)
