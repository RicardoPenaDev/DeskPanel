// Package protocol implements the DeskPanel wire protocol described in
// docs/PROTOCOL.md and PROJECT.md §8-9: the HTTP/WebSocket envelope, error
// codes and message size limits. It does not itself open sockets — that is
// internal/websocket and internal/api.
package protocol

import (
	"encoding/json"
	"time"
)

// Version is the DeskPanel wire protocol version implemented by this build.
// Incompatible changes must bump this and the agent must reject mismatched
// clients with ErrProtocolUnsupported.
const Version = 1

// Message size limits (PROJECT.md §12.4-12.5).
const (
	MaxWebSocketMessageBytes = 16 * 1024
	MaxHTTPRequestBytes      = 64 * 1024
)

// Envelope is the standard wrapper for every WebSocket message
// (docs/PROTOCOL.md "Envelope padrão").
type Envelope struct {
	Type            string          `json:"type"`
	ProtocolVersion int             `json:"protocolVersion"`
	RequestID       string          `json:"requestId"`
	Timestamp       time.Time       `json:"timestamp"`
	Payload         json.RawMessage `json:"payload"`
}

// Message types.
const (
	TypeAuthHello     = "auth.hello"
	TypeAuthAccepted  = "auth.accepted"
	TypeActionExecute = "action.execute"
	TypeActionStarted = "action.started"
	TypeActionResult  = "action.result"
	TypePing          = "ping"
	TypePong          = "pong"
	TypeStateSnapshot = "state.snapshot"
	TypeStateChanged  = "state.changed"
	TypeError         = "error"
)

// ErrorCode is one of the fixed error codes from docs/PROTOCOL.md.
type ErrorCode string

const (
	ErrAuthRequired        ErrorCode = "AUTH_REQUIRED"
	ErrAuthInvalid         ErrorCode = "AUTH_INVALID"
	ErrAuthRevoked         ErrorCode = "AUTH_REVOKED"
	ErrPairingClosed       ErrorCode = "PAIRING_CLOSED"
	ErrPairingCodeInvalid  ErrorCode = "PAIRING_CODE_INVALID"
	ErrPairingCodeExpired  ErrorCode = "PAIRING_CODE_EXPIRED"
	ErrProtocolUnsupported ErrorCode = "PROTOCOL_UNSUPPORTED"
	ErrInvalidMessage      ErrorCode = "INVALID_MESSAGE"
	ErrActionNotFound      ErrorCode = "ACTION_NOT_FOUND"
	ErrActionNotAllowed    ErrorCode = "ACTION_NOT_ALLOWED"
	ErrActionTimeout       ErrorCode = "ACTION_TIMEOUT"
	ErrActionFailed        ErrorCode = "ACTION_FAILED"
	ErrRateLimited         ErrorCode = "RATE_LIMITED"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
)
