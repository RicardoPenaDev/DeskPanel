// Package api implementa os endpoints HTTP do DeskPanel Agent
// (docs/PROTOCOL.md "HTTP"). Só /health, /version e /pair existem fora de
// autenticação; o resto exige Bearer token. Não existe endpoint genérico
// de execução — isso é só via WebSocket (internal/websocket).
package api

import (
	"net/http"

	"deskpanel-agent/internal/actions"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/pairing"
	"deskpanel-agent/internal/ratelimit"
)

// Server monta o roteador HTTP completo do agente.
type Server struct {
	Actions      []actions.Action
	Devices      *devices.Store
	Pairing      *pairing.Manager
	Limiter      *ratelimit.Limiter
	Failures     *ratelimit.FailureTracker
	AgentVersion string
	MacName      string

	// WSHandler atende GET /api/v1/ws. Injetado de fora (internal/websocket)
	// para não criar ciclo de import; se nil, /ws não é registrada.
	WSHandler http.Handler
}

// Handler monta o http.Handler completo, com middlewares aplicados.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", HealthHandler)
	mux.HandleFunc("GET /api/v1/version", VersionHandler(s.AgentVersion))
	mux.HandleFunc("POST /api/v1/pair", PairHandler(s.Pairing, s.Devices, s.Failures))
	mux.Handle("GET /api/v1/actions", requireBearerToken(s.Devices)(ActionsHandler(s.Actions)))
	mux.Handle("GET /api/v1/state", requireBearerToken(s.Devices)(StateHandler(s.MacName, s.AgentVersion)))
	if s.WSHandler != nil {
		mux.Handle("GET /api/v1/ws", s.WSHandler)
	}

	var handler http.Handler = mux
	handler = requirePrivateNetwork(handler)
	handler = rateLimit(s.Limiter)(handler)
	handler = maxBody(handler)
	return handler
}
