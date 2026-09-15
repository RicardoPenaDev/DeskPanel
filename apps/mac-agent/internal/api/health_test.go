package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("Status = %q, want %q", body.Status, "ok")
	}
	if body.ProtocolVersion != 1 {
		t.Errorf("ProtocolVersion = %d, want 1", body.ProtocolVersion)
	}

	// PROJECT.md §8.2: nunca vazar informação interna nesta resposta.
	raw := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"token", "password", "home", "/users/"} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("resposta de /health contém %q, não deveria vazar dado interno", forbidden)
		}
	}
}
