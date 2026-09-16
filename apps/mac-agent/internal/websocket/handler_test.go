package websocket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"deskpanel-agent/internal/actions"
	"deskpanel-agent/internal/appscan"
	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/executor"
	"deskpanel-agent/internal/protocol"
)

const testToken = "token-de-teste-para-websocket"

func newTestHandler(t *testing.T, exec executor.Executor) (*Handler, *devices.Store) {
	t.Helper()
	store, err := devices.LoadStore(filepath.Join(t.TempDir(), "devices.json"))
	if err != nil {
		t.Fatalf("LoadStore() erro: %v", err)
	}
	_ = store.Put(devices.Device{DeviceID: "dev-1", DeviceName: "Moto G60", TokenHash: auth.HashToken(testToken)})

	h := &Handler{
		Devices:      store,
		Actions:      []actions.Action{{ID: "app.chrome", Label: "Chrome", Kind: actions.KindOpenApp}},
		Executor:     exec,
		MacName:      "Mac de Teste",
		AgentVersion: "0.1.0-test",
		AuthTimeout:  time.Second,
	}
	return h, store
}

func wsURL(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func dial(t *testing.T, server *httptest.Server) *coderws.Conn {
	t.Helper()
	conn, _, err := coderws.Dial(context.Background(), wsURL(server), nil)
	if err != nil {
		t.Fatalf("Dial() erro: %v", err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn
}

func sendEnvelope(t *testing.T, ctx context.Context, conn *coderws.Conn, msgType, requestID string, payload any) {
	t.Helper()
	if err := wsjson.Write(ctx, conn, newEnvelope(msgType, requestID, payload)); err != nil {
		t.Fatalf("write(%s) erro: %v", msgType, err)
	}
}

func readEnvelope(t *testing.T, ctx context.Context, conn *coderws.Conn) protocol.Envelope {
	t.Helper()
	var env protocol.Envelope
	if err := wsjson.Read(ctx, conn, &env); err != nil {
		t.Fatalf("read() erro: %v", err)
	}
	return env
}

// readEnvelopeSkipping lê mensagens até achar um tipo que não esteja em skip
// — útil para ignorar o state.snapshot automático quando o teste só quer a
// resposta de auth.
func readEnvelopeUntil(t *testing.T, ctx context.Context, conn *coderws.Conn, wantType string) protocol.Envelope {
	t.Helper()
	for i := 0; i < 5; i++ {
		env := readEnvelope(t, ctx, conn)
		if env.Type == wantType {
			return env
		}
	}
	t.Fatalf("não recebeu mensagem do tipo %q depois de várias tentativas", wantType)
	return protocol.Envelope{}
}

func authenticate(t *testing.T, ctx context.Context, conn *coderws.Conn, deviceID, token string) protocol.Envelope {
	t.Helper()
	sendEnvelope(t, ctx, conn, protocol.TypeAuthHello, "req-auth", authHelloPayload{DeviceID: deviceID, AccessToken: token, AppVersion: "0.1.0"})
	return readEnvelope(t, ctx, conn)
}

func TestWebSocket_AuthSuccess(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()

	env := authenticate(t, ctx, conn, "dev-1", testToken)
	if env.Type != protocol.TypeAuthAccepted {
		t.Fatalf("Type = %q, want %q", env.Type, protocol.TypeAuthAccepted)
	}

	// docs/PROTOCOL.md: logo após autenticar, o servidor manda o snapshot.
	snapshot := readEnvelope(t, ctx, conn)
	if snapshot.Type != protocol.TypeStateSnapshot {
		t.Fatalf("Type = %q, want %q", snapshot.Type, protocol.TypeStateSnapshot)
	}
	var payload stateSnapshotPayload
	_ = json.Unmarshal(snapshot.Payload, &payload)
	if payload.MacName != "Mac de Teste" {
		t.Errorf("MacName = %q, want %q", payload.MacName, "Mac de Teste")
	}
}

func TestWebSocket_AuthInvalidToken(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()

	env := authenticate(t, ctx, conn, "dev-1", "token-errado")
	if env.Type != protocol.TypeError {
		t.Fatalf("Type = %q, want %q", env.Type, protocol.TypeError)
	}
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrAuthInvalid {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrAuthInvalid)
	}
}

func TestWebSocket_AuthUnknownDevice(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	env := authenticate(t, context.Background(), conn, "dev-fantasma", testToken)
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrAuthInvalid {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrAuthInvalid)
	}
}

func TestWebSocket_AuthRevokedDeviceRejectedImmediately(t *testing.T) {
	h, store := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	if err := store.Revoke("dev-1"); err != nil {
		t.Fatalf("Revoke() erro: %v", err)
	}

	conn := dial(t, server)
	env := authenticate(t, context.Background(), conn, "dev-1", testToken)
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrAuthRevoked {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrAuthRevoked)
	}
}

func TestWebSocket_FirstMessageMustBeAuthHello(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-1", actionExecutePayload{ActionID: "app.chrome"})

	env := readEnvelope(t, ctx, conn)
	if env.Type != protocol.TypeError {
		t.Fatalf("Type = %q, want %q", env.Type, protocol.TypeError)
	}
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrAuthRequired {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrAuthRequired)
	}
}

func TestWebSocket_ExecuteAction_Success(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-1", actionExecutePayload{ActionID: "app.chrome"})

	started := readEnvelope(t, ctx, conn)
	if started.Type != protocol.TypeActionStarted {
		t.Fatalf("Type = %q, want %q", started.Type, protocol.TypeActionStarted)
	}

	result := readEnvelope(t, ctx, conn)
	if result.Type != protocol.TypeActionResult {
		t.Fatalf("Type = %q, want %q", result.Type, protocol.TypeActionResult)
	}
	var payload actionResultPayload
	_ = json.Unmarshal(result.Payload, &payload)
	if payload.Status != "success" {
		t.Errorf("Status = %q, want success", payload.Status)
	}
}

func TestWebSocket_ExecuteAction_UnknownAction(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-1", actionExecutePayload{ActionID: "não-existe"})

	env := readEnvelope(t, ctx, conn)
	if env.Type != protocol.TypeError {
		t.Fatalf("Type = %q, want %q", env.Type, protocol.TypeError)
	}
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrActionNotFound {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrActionNotFound)
	}
}

// recordingExecutor guarda a última ação executada, para inspecionar os
// parâmetros que o handler efetivamente montou.
type recordingExecutor struct {
	last actions.Action
}

func (r *recordingExecutor) Execute(ctx context.Context, action actions.Action) executor.Result {
	r.last = action
	return executor.Result{Status: "success"}
}

// writeFakeApp cria um .app mínimo com Info.plist real, o suficiente
// para appscan.Scanner reconhecer via plutil.
func writeFakeApp(t *testing.T, dir, appName, bundleName string) {
	t.Helper()
	appPath := filepath.Join(dir, appName+".app")
	if err := os.MkdirAll(filepath.Join(appPath, "Contents"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>` + bundleName + `</string>
</dict>
</plist>
`
	if err := os.WriteFile(filepath.Join(appPath, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatalf("WriteFile Info.plist: %v", err)
	}
}

// TestWebSocket_ExecuteAction_DynamicAppFromScan prova que um app que
// nunca esteve no config.json (só apareceu na varredura ao vivo de
// /Applications) consegue ser executado como open_app — e que o
// executor recebe o caminho do bundle, nunca um nome/arg vindo do
// Android (PROJECT.md §7.4).
func TestWebSocket_ExecuteAction_DynamicAppFromScan(t *testing.T) {
	dir := t.TempDir()
	writeFakeApp(t, dir, "MeuApp", "Meu App")

	rec := &recordingExecutor{}
	h, _ := newTestHandler(t, rec)
	scanner := &appscan.Scanner{Dirs: []string{dir}}
	h.Apps = scanner

	apps := scanner.Scan()
	if len(apps) != 1 {
		t.Fatalf("setup: Scan() encontrou %d apps", len(apps))
	}
	appID := apps[0].ID

	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-1", actionExecutePayload{ActionID: appID})

	readEnvelope(t, ctx, conn) // action.started
	result := readEnvelope(t, ctx, conn)
	if result.Type != protocol.TypeActionResult {
		t.Fatalf("Type = %q, want %q", result.Type, protocol.TypeActionResult)
	}
	var payload actionResultPayload
	_ = json.Unmarshal(result.Payload, &payload)
	if payload.Status != "success" {
		t.Fatalf("Status = %q, want success", payload.Status)
	}

	if rec.last.Kind != actions.KindOpenApp {
		t.Fatalf("Kind = %q, want %q", rec.last.Kind, actions.KindOpenApp)
	}
	if rec.last.Parameters["application"] != apps[0].Path {
		t.Errorf("application = %v, want %q", rec.last.Parameters["application"], apps[0].Path)
	}
}

// TestWebSocket_ExecuteAction_UninstalledAppRejected prova que um ID de
// app que sumiu do disco entre a listagem e a execução é rejeitado como
// ação desconhecida, nunca executado com dados obsoletos.
func TestWebSocket_ExecuteAction_UninstalledAppRejected(t *testing.T) {
	dir := t.TempDir()
	writeFakeApp(t, dir, "Efemero", "App Efêmero")

	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	scanner := &appscan.Scanner{Dirs: []string{dir}}
	h.Apps = scanner

	apps := scanner.Scan()
	appID := apps[0].ID
	if err := os.RemoveAll(filepath.Join(dir, "Efemero.app")); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-1", actionExecutePayload{ActionID: appID})

	env := readEnvelope(t, ctx, conn)
	if env.Type != protocol.TypeError {
		t.Fatalf("Type = %q, want %q", env.Type, protocol.TypeError)
	}
	var payload errorPayload
	_ = json.Unmarshal(env.Payload, &payload)
	if payload.Code != protocol.ErrActionNotFound {
		t.Errorf("Code = %q, want %q", payload.Code, protocol.ErrActionNotFound)
	}
}

// countingExecutor conta quantas vezes Execute roda de verdade — usado
// para provar a deduplicação por requestId (PROJECT.md §16).
type countingExecutor struct {
	calls int32
}

func (c *countingExecutor) Execute(ctx context.Context, action actions.Action) executor.Result {
	atomic.AddInt32(&c.calls, 1)
	return executor.Result{Status: "success"}
}

func TestWebSocket_ExecuteAction_DedupSameRequestID(t *testing.T) {
	exec := &countingExecutor{}
	h, _ := newTestHandler(t, exec)
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	// 1ª vez: executa de verdade — vem started + result.
	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-repetido", actionExecutePayload{ActionID: "app.chrome"})
	if env := readEnvelope(t, ctx, conn); env.Type != protocol.TypeActionStarted {
		t.Fatalf("1ª execução: Type = %q, want %q", env.Type, protocol.TypeActionStarted)
	}
	if env := readEnvelope(t, ctx, conn); env.Type != protocol.TypeActionResult {
		t.Fatalf("1ª execução: Type = %q, want %q", env.Type, protocol.TypeActionResult)
	}

	// 2ª vez, mesmo requestId: dedup — só o result (cacheado) volta, sem
	// started novo e sem chamar o executor de novo.
	sendEnvelope(t, ctx, conn, protocol.TypeActionExecute, "req-repetido", actionExecutePayload{ActionID: "app.chrome"})
	if env := readEnvelope(t, ctx, conn); env.Type != protocol.TypeActionResult {
		t.Fatalf("2ª execução (dedup): Type = %q, want %q", env.Type, protocol.TypeActionResult)
	}

	if got := atomic.LoadInt32(&exec.calls); got != 1 {
		t.Errorf("Executor.Execute foi chamado %d vezes, want 1 (dedup por requestId)", got)
	}
}

func TestWebSocket_PingPong_KeepsConnectionAlive(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	h.PingInterval = 30 * time.Millisecond
	h.PongTimeout = 30 * time.Millisecond
	h.MaxMissedPings = 2
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	// Responde a vários ciclos de heartbeat seguidos — a conexão deve
	// continuar de pé enquanto o cliente responder a tempo.
	for i := 0; i < 3; i++ {
		env := readEnvelope(t, ctx, conn)
		if env.Type != protocol.TypePing {
			t.Fatalf("ciclo %d: Type = %q, want %q", i, env.Type, protocol.TypePing)
		}
		sendEnvelope(t, ctx, conn, protocol.TypePong, env.RequestID, struct{}{})
	}
}

func TestWebSocket_MissedPongs_ClosesConnection(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	h.PingInterval = 20 * time.Millisecond
	h.PongTimeout = 20 * time.Millisecond
	h.MaxMissedPings = 2
	server := httptest.NewServer(h)
	defer server.Close()

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	// nunca responde pong — a conexão deve cair sozinha.
	readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var lastErr error
	for i := 0; i < 20; i++ {
		if _, _, err := conn.Read(readCtx); err != nil {
			lastErr = err
			break
		}
	}
	if lastErr == nil {
		t.Fatal("conexão deveria ter sido encerrada depois de perder os pongs")
	}
}

func TestHandler_ActiveConnections_TracksLifecycle(t *testing.T) {
	h, _ := newTestHandler(t, &executor.FakeExecutor{})
	server := httptest.NewServer(h)
	defer server.Close()

	if got := h.ActiveConnections(); got != 0 {
		t.Fatalf("ActiveConnections() antes de conectar = %d, want 0", got)
	}

	conn := dial(t, server)
	ctx := context.Background()
	authenticate(t, ctx, conn, "dev-1", testToken)
	readEnvelope(t, ctx, conn) // state.snapshot

	if got := h.ActiveConnections(); got != 1 {
		t.Fatalf("ActiveConnections() com 1 conexão = %d, want 1", got)
	}

	_ = conn.Close(coderws.StatusNormalClosure, "fim do teste")

	deadline := time.Now().Add(2 * time.Second)
	for h.ActiveConnections() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := h.ActiveConnections(); got != 0 {
		t.Fatalf("ActiveConnections() depois de fechar = %d, want 0", got)
	}
}
