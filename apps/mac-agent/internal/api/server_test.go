package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"deskpanel-agent/internal/actions"
	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/pairing"
	"deskpanel-agent/internal/protocol"
	"deskpanel-agent/internal/ratelimit"
)

func newTestServer(t *testing.T) (*Server, *devices.Store, *pairing.Manager) {
	t.Helper()
	store, err := devices.LoadStore(filepath.Join(t.TempDir(), "devices.json"))
	if err != nil {
		t.Fatalf("LoadStore() erro: %v", err)
	}
	mgr := pairing.NewManager()

	s := &Server{
		Actions: []actions.Action{
			{ID: "app.chrome", Label: "Chrome", Icon: "chrome", Kind: actions.KindOpenApp, Parameters: map[string]any{"application": "Google Chrome"}},
			{ID: "screen.lock", Label: "Bloquear", Kind: actions.KindScreenLock},
		},
		Devices:      store,
		Pairing:      mgr,
		Limiter:      ratelimit.NewLimiter(0),
		Failures:     ratelimit.NewFailureTracker(5, time.Minute),
		AgentVersion: "0.1.0-test",
		MacName:      "Mac de Teste",
	}
	return s, store, mgr
}

// loopbackRequest configura RemoteAddr para um endereço de rede local, já
// que httptest não define isso por padrão (fica vazio -> reprovaria
// requirePrivateNetwork).
func loopbackRequest(method, target string, body []byte) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	r.RemoteAddr = "127.0.0.1:54321"
	return r
}

func TestHealth_NoAuthRequired(t *testing.T) {
	s, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/health", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestVersion(t *testing.T) {
	s, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/version", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body versionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.ProtocolVersion != protocol.Version {
		t.Errorf("ProtocolVersion = %d, want %d", body.ProtocolVersion, protocol.Version)
	}
}

func TestPair_Success(t *testing.T) {
	s, store, mgr := newTestServer(t)
	code, _ := mgr.Open(5 * time.Minute)

	reqBody, _ := json.Marshal(pairRequest{
		Code: code, DeviceID: "dev-1", DeviceName: "Moto G60",
		AppVersion: "0.1.0", ProtocolVersion: protocol.Version,
	})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("POST", "/api/v1/pair", reqBody))

	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp pairResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.AccessToken == "" || resp.DeviceID != "dev-1" {
		t.Fatalf("resposta inesperada: %+v", resp)
	}

	dev, ok := store.Find("dev-1")
	if !ok {
		t.Fatal("dispositivo deveria ter sido salvo")
	}
	if dev.TokenHash != auth.HashToken(resp.AccessToken) {
		t.Error("hash salvo não bate com o token retornado")
	}
}

func TestPair_WrongCodeRejected(t *testing.T) {
	s, _, mgr := newTestServer(t)
	_, _ = mgr.Open(5 * time.Minute)

	reqBody, _ := json.Marshal(pairRequest{
		Code: "000000", DeviceID: "dev-1", DeviceName: "X", ProtocolVersion: protocol.Version,
	})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("POST", "/api/v1/pair", reqBody))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	var body errorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error != protocol.ErrPairingCodeInvalid {
		t.Errorf("Error = %q, want %q", body.Error, protocol.ErrPairingCodeInvalid)
	}
}

func TestPair_UnsupportedProtocolVersion(t *testing.T) {
	s, _, mgr := newTestServer(t)
	code, _ := mgr.Open(5 * time.Minute)

	reqBody, _ := json.Marshal(pairRequest{Code: code, DeviceID: "dev-1", DeviceName: "X", ProtocolVersion: 999})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("POST", "/api/v1/pair", reqBody))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestActions_RequiresToken(t *testing.T) {
	s, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/actions", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 sem token", rec.Code)
	}
}

func TestActions_ValidTokenSeesActionsWithoutParameters(t *testing.T) {
	s, store, _ := newTestServer(t)
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/actions", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("Google Chrome")) {
		t.Error("resposta de /actions não deveria expor Parameters (ex.: nome do app)")
	}
	var list []actionSummary
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
}

func TestActions_RevokedDeviceLosesAccessImmediately(t *testing.T) {
	s, store, _ := newTestServer(t)
	token := "token-de-teste"
	_ = store.Put(devices.Device{DeviceID: "dev-1", TokenHash: auth.HashToken(token)})

	r := loopbackRequest("GET", "/api/v1/actions", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	rec1 := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec1, r)
	if rec1.Code != 200 {
		t.Fatalf("antes de revogar: status = %d", rec1.Code)
	}

	if err := store.Revoke("dev-1"); err != nil {
		t.Fatalf("Revoke() erro: %v", err)
	}

	r2 := loopbackRequest("GET", "/api/v1/actions", nil)
	r2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec2, r2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("depois de revogar: status = %d, want 401", rec2.Code)
	}
}

func TestState_RequiresToken(t *testing.T) {
	s, _, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, loopbackRequest("GET", "/api/v1/state", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestPublicAddressIsRejected(t *testing.T) {
	s, _, _ := newTestServer(t)
	r := httptest.NewRequest("GET", "/api/v1/health", nil)
	r.RemoteAddr = "8.8.8.8:12345" // endereço público
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 para IP público", rec.Code)
	}
}

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	s, _, _ := newTestServer(t)
	s.Limiter = ratelimit.NewLimiter(1)

	rec1 := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec1, loopbackRequest("GET", "/api/v1/health", nil))
	if rec1.Code != 200 {
		t.Fatalf("1ª requisição: status = %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec2, loopbackRequest("GET", "/api/v1/health", nil))
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("2ª requisição: status = %d, want 429", rec2.Code)
	}
}
