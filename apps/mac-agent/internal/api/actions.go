package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/actions"
)

// actionSummary é o que o Android recebe: o suficiente para desenhar o
// botão. Parameters nunca é exposto — é detalhe interno de execução no Mac.
type actionSummary struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	Icon             string `json:"icon"`
	Kind             string `json:"kind"`
	RequireLongPress bool   `json:"requireLongPress"`
}

// ActionsHandler implementa GET /api/v1/actions (Bearer token obrigatório).
func ActionsHandler(catalog []actions.Action) http.HandlerFunc {
	summaries := make([]actionSummary, 0, len(catalog))
	for _, a := range catalog {
		summaries = append(summaries, actionSummary{
			ID:               a.ID,
			Label:            a.Label,
			Icon:             a.Icon,
			Kind:             string(a.Kind),
			RequireLongPress: a.RequiresLongPress(),
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summaries)
	}
}
