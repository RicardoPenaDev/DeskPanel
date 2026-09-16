package weather

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeDoer responde com um corpo fixo por URL (contém), sem nunca tocar
// a rede de verdade — contadores permitem provar o cache.
type fakeDoer struct {
	geoBody      string
	geoStatus    int
	forecastBody string
	forecastCode int

	geoCalls      int32
	forecastCalls int32
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	switch {
	case strings.Contains(url, "ip-api.com"):
		atomic.AddInt32(&f.geoCalls, 1)
		status := f.geoStatus
		if status == 0 {
			status = http.StatusOK
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(f.geoBody))}, nil
	case strings.Contains(url, "open-meteo.com"):
		atomic.AddInt32(&f.forecastCalls, 1)
		code := f.forecastCode
		if code == 0 {
			code = http.StatusOK
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(f.forecastBody))}, nil
	default:
		return nil, context.DeadlineExceeded
	}
}

const validGeoBody = `{"status":"success","city":"São Paulo","lat":-23.55,"lon":-46.63}`
const validForecastBody = `{"current":{"temperature_2m":24.4,"weather_code":1}}`

func TestGet_Success(t *testing.T) {
	doer := &fakeDoer{geoBody: validGeoBody, forecastBody: validForecastBody}
	p := &Provider{Client: doer}

	snap, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() erro: %v", err)
	}
	if snap.City != "São Paulo" {
		t.Errorf("City = %q, want São Paulo", snap.City)
	}
	if snap.TempC != 24.4 {
		t.Errorf("TempC = %v, want 24.4", snap.TempC)
	}
	if snap.Description != "poucas nuvens" {
		t.Errorf("Description = %q, want %q", snap.Description, "poucas nuvens")
	}
}

func TestGet_CachesWithinTTL(t *testing.T) {
	doer := &fakeDoer{geoBody: validGeoBody, forecastBody: validForecastBody}
	p := &Provider{Client: doer}

	if _, err := p.Get(context.Background()); err != nil {
		t.Fatalf("1ª Get() erro: %v", err)
	}
	if _, err := p.Get(context.Background()); err != nil {
		t.Fatalf("2ª Get() erro: %v", err)
	}

	if doer.geoCalls != 1 || doer.forecastCalls != 1 {
		t.Errorf("chamadas = geo:%d forecast:%d, want 1 e 1 (deveria vir do cache na 2ª)", doer.geoCalls, doer.forecastCalls)
	}
}

func TestGet_GeolocationRejected(t *testing.T) {
	doer := &fakeDoer{geoBody: `{"status":"fail","message":"private range"}`, forecastBody: validForecastBody}
	p := &Provider{Client: doer}

	if _, err := p.Get(context.Background()); err == nil {
		t.Fatal("Get() deveria falhar quando a geolocalização recusa")
	}
}

func TestGet_HTTPErrorPropagates(t *testing.T) {
	doer := &fakeDoer{geoBody: validGeoBody, forecastBody: "", forecastCode: http.StatusServiceUnavailable}
	p := &Provider{Client: doer}

	if _, err := p.Get(context.Background()); err == nil {
		t.Fatal("Get() deveria falhar quando a previsão retorna erro HTTP")
	}
}

func TestDescribe_UnknownCodeFallsBackToGeneric(t *testing.T) {
	if got := describe(9999); got != "tempo variável" {
		t.Errorf("describe(9999) = %q, want %q", got, "tempo variável")
	}
}

func TestDescribe_KnownCodes(t *testing.T) {
	cases := map[int]string{
		0:  "céu limpo",
		3:  "nublado",
		61: "chuva",
		95: "tempestade",
	}
	for code, want := range cases {
		if got := describe(code); got != want {
			t.Errorf("describe(%d) = %q, want %q", code, got, want)
		}
	}
}
