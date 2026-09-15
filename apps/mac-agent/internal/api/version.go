package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/protocol"
)

type versionResponse struct {
	AgentVersion    string `json:"agentVersion"`
	ProtocolVersion int    `json:"protocolVersion"`
}

// VersionHandler implementa GET /api/v1/version — não requer autenticação
// (PROJECT.md §8.1).
func VersionHandler(agentVersion string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(versionResponse{
			AgentVersion:    agentVersion,
			ProtocolVersion: protocol.Version,
		})
	}
}
