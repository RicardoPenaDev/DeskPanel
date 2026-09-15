// Package api implements the DeskPanel HTTP endpoints (docs/PROTOCOL.md
// "HTTP"). Only /health, /version and /pair exist outside authentication;
// everything else requires a Bearer token. There is intentionally no
// generic command-execution endpoint.
package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/protocol"
)

type healthResponse struct {
	Status          string `json:"status"`
	Service         string `json:"service"`
	ProtocolVersion int    `json:"protocolVersion"`
}

// HealthHandler implements GET /api/v1/health. It must never leak hostname,
// user, paths, internal IPs, tokens or the action list (PROJECT.md §8.2).
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:          "ok",
		Service:         "deskpanel-agent",
		ProtocolVersion: protocol.Version,
	})
}
