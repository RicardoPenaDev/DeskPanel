// Package weather busca a previsão do tempo atual para a tela ambiente
// do Android (relógio + clima). É a única chamada de saída do próprio
// Mac para a internet geral no agente — diferente do resto (PROJECT.md
// §11/12: cliente Android <-> agente só na rede local). O Mac já tem
// acesso normal à internet como qualquer computador; aqui só é leitura
// (geolocalização pelo IP público do Mac + previsão), nada é enviado
// além do próprio IP de origem da requisição HTTP, e o resultado fica
// em cache para não bater nas APIs gratuitas a cada pedido do Android.
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	geoIPURL       = "http://ip-api.com/json/?fields=status,message,lat,lon,city"
	weatherURLFmt  = "https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,weather_code&timezone=auto"
	cacheTTL       = 15 * time.Minute
	requestTimeout = 5 * time.Second
)

// Snapshot é o que o Android recebe em GET /api/v1/weather.
type Snapshot struct {
	City        string    `json:"city"`
	TempC       float64   `json:"tempC"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// httpDoer permite injetar um transporte falso nos testes — nenhum
// teste deve bater na internet de verdade (mesmo padrão do
// runCommandFunc injetável em internal/executor).
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Provider busca e cacheia o clima atual.
type Provider struct {
	// Client é o transporte HTTP; nil usa http.DefaultClient.
	Client httpDoer

	mu       sync.Mutex
	cached   Snapshot
	cachedAt time.Time
}

// New cria um Provider pronto para uso com o cliente HTTP padrão.
func New() *Provider {
	return &Provider{}
}

func (p *Provider) client() httpDoer {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

// Get devolve o clima em cache se ainda estiver fresco (< 15 min);
// senão busca de novo (geolocalização por IP + previsão do Open-Meteo).
func (p *Provider) Get(ctx context.Context) (Snapshot, error) {
	p.mu.Lock()
	if !p.cachedAt.IsZero() && time.Since(p.cachedAt) < cacheTTL {
		cached := p.cached
		p.mu.Unlock()
		return cached, nil
	}
	p.mu.Unlock()

	snap, err := p.fetch(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	p.mu.Lock()
	p.cached = snap
	p.cachedAt = time.Now()
	p.mu.Unlock()
	return snap, nil
}

type geoResponse struct {
	Status string  `json:"status"`
	City   string  `json:"city"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

type forecastResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

func (p *Provider) fetch(ctx context.Context) (Snapshot, error) {
	geo, err := p.geolocate(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	forecast, err := p.forecast(ctx, geo.Lat, geo.Lon)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		City:        geo.City,
		TempC:       forecast.Current.Temperature,
		Description: describe(forecast.Current.WeatherCode),
		UpdatedAt:   time.Now(),
	}, nil
}

func (p *Provider) geolocate(ctx context.Context) (geoResponse, error) {
	var geo geoResponse
	if err := p.getJSON(ctx, geoIPURL, &geo); err != nil {
		return geoResponse{}, fmt.Errorf("weather: geolocalização: %w", err)
	}
	if geo.Status != "success" {
		return geoResponse{}, fmt.Errorf("weather: geolocalização recusada (status %q)", geo.Status)
	}
	return geo, nil
}

func (p *Provider) forecast(ctx context.Context, lat, lon float64) (forecastResponse, error) {
	var f forecastResponse
	url := fmt.Sprintf(weatherURLFmt, lat, lon)
	if err := p.getJSON(ctx, url, &f); err != nil {
		return forecastResponse{}, fmt.Errorf("weather: previsão: %w", err)
	}
	return f, nil
}

func (p *Provider) getJSON(ctx context.Context, url string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resposta HTTP %d de %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// describe traduz o weather_code (código WMO) do Open-Meteo numa
// descrição curta em português. Um código não mapeado cai num genérico
// em vez de quebrar a exibição.
func describe(code int) string {
	switch {
	case code == 0:
		return "céu limpo"
	case code == 1 || code == 2:
		return "poucas nuvens"
	case code == 3:
		return "nublado"
	case code == 45 || code == 48:
		return "névoa"
	case code >= 51 && code <= 57:
		return "garoa"
	case code >= 61 && code <= 67:
		return "chuva"
	case code >= 71 && code <= 77:
		return "neve"
	case code >= 80 && code <= 82:
		return "pancadas de chuva"
	case code >= 85 && code <= 86:
		return "pancadas de neve"
	case code >= 95 && code <= 99:
		return "tempestade"
	default:
		return "tempo variável"
	}
}
