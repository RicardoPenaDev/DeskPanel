package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/state"
)

// StateHandler implementa GET /api/v1/state (Bearer token obrigatório).
func StateHandler(macName, agentVersion string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state.Snapshot{
			MacName:      macName,
			AgentVersion: agentVersion,
		})
	}
}
