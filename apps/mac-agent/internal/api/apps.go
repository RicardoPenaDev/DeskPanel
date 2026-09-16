package api

import (
	"encoding/json"
	"net/http"

	"deskpanel-agent/internal/appscan"
)

// appSummary é o que o Android recebe por app instalado — sem o caminho
// real em disco, só o suficiente para listar e buscar o ícone
// (mesmo princípio de actionSummary em actions.go).
type appSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	HasIcon bool   `json:"hasIcon"`
}

// AppsHandler implementa GET /api/v1/apps (Bearer token obrigatório): a
// lista ao vivo de tudo instalado em /Applications e afins, para o
// editor Android deixar escolher qualquer app como atalho.
func AppsHandler(scanner *appscan.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apps := scanner.Scan()
		summaries := make([]appSummary, 0, len(apps))
		for _, a := range apps {
			summaries = append(summaries, appSummary{ID: a.ID, Name: a.Name, HasIcon: a.HasIcon()})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summaries)
	}
}

// AppIconHandler implementa GET /api/v1/apps/{id}/icon (Bearer token
// obrigatório): o PNG do ícone real do app, convertido do .icns do
// bundle. 404 se o app sumiu do disco ou nunca teve ícone.
func AppIconHandler(scanner *appscan.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		png, ok := scanner.Icon(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(png)
	}
}
