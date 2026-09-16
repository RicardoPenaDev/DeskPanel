package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/weather"
)

// fakeWeatherDoer responde com corpos fixos por URL, sem tocar a
// internet de verdade — mesmo padrão de internal/weather/weather_test.go.
type fakeWeatherDoer struct {
	geoBody      string
	forecastBody string
}

func (f *fakeWeatherDoer) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if strings.Contains(url, "ip-api.com") {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(f.geoBody))}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(f.forecastBody))}, nil
}

func TestWeather_RequiresToken(t *testing.T) {
	s, _, _ := newTestServer(t)
	s.Weather = &weather.Provider{}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/weather", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 sem token", rec.Code)
	}
}

func TestWeather_ValidTokenReturnsSnapshot(t *testing.T) {
	s, store, _ := newTestServer(t)
	s.Weather = &weather.Provider{Client: &fakeWeatherDoer{
		geoBody:      `{"status":"success","city":"São Paulo","lat":-23.55,"lon":-46.63}`,
		forecastBody: `{"current":{"temperature_2m":24.4,"weather_code":0}}`,
	}}
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/weather", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var snap weather.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if snap.City != "São Paulo" || snap.Description != "céu limpo" {
		t.Fatalf("snapshot inesperado: %+v", snap)
	}
}

func TestWeather_ProviderErrorReturns503(t *testing.T) {
	s, store, _ := newTestServer(t)
	s.Weather = &weather.Provider{Client: &fakeWeatherDoer{
		geoBody:      `{"status":"fail"}`,
		forecastBody: `{}`,
	}}
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/weather", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
