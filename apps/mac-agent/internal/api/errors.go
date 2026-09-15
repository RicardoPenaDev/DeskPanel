package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/protocol"
)

type errorResponse struct {
	Error   protocol.ErrorCode `json:"error"`
	Message string             `json:"message,omitempty"`
}

// writeError escreve uma resposta de erro em JSON. message deve ser segura
// para o cliente ver — nunca detalhes internos (PROJECT.md §13).
func writeError(w http.ResponseWriter, status int, code protocol.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: code, Message: message})
}
