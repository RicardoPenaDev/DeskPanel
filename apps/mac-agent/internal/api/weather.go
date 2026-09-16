package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/weather"
)

// WeatherHandler implementa GET /api/v1/weather (Bearer token
// obrigatório): o clima atual para a tela ambiente do Android. Uma
// falha (sem internet, geolocalização recusada, etc.) vira 503 em vez
// de travar o app — o relógio ambiente simplesmente não mostra clima.
func WeatherHandler(provider *weather.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap, err := provider.Get(r.Context())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, protocol.ErrWeatherUnavailable, "não foi possível obter o clima agora")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snap)
	}
}
