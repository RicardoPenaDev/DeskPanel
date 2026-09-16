// Package websocket implementa o canal em tempo real do DeskPanel
// (docs/PROTOCOL.md "WebSocket"): autenticação, execução de ações,
// heartbeat e deduplicação por requestId. Usa github.com/coder/websocket
// (ver docs/DECISIONS.md ADR-0003) porque a stdlib não tem WebSocket.
package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"deskpanel-agent/internal/actions"
	"deskpanel-agent/internal/appscan"
	"deskpanel-agent/internal/auth"
	"deskpanel-agent/internal/corsorigins"
	"deskpanel-agent/internal/devices"
	"deskpanel-agent/internal/executor"
	"deskpanel-agent/internal/protocol"
)

const (
	defaultAuthTimeout    = 5 * time.Second
	defaultPingInterval   = 20 * time.Second
	defaultPongTimeout    = 10 * time.Second
	defaultMaxMissedPings = 2
	defaultActionTimeout  = 10 * time.Second
	defaultDedupTTL       = 2 * time.Minute
)

// Handler atende GET /api/v1/ws. Uma instância é compartilhada por todas
// as conexões; os campos exportados são só leitura depois de construída.
type Handler struct {
	Devices      *devices.Store
	Actions      []actions.Action
	Apps         *appscan.Scanner
	Executor     executor.Executor
	MacName      string
	AgentVersion string
	Logger       *slog.Logger

	// Timeouts — zero usa o default de produção; testes ajustam para
	// valores pequenos.
	AuthTimeout    time.Duration
	PingInterval   time.Duration
	PongTimeout    time.Duration
	MaxMissedPings int
	ActionTimeout  time.Duration
	DedupTTL       time.Duration

	activeConns int64
}

func (h *Handler) authTimeout() time.Duration {
	if h.AuthTimeout > 0 {
		return h.AuthTimeout
	}
	return defaultAuthTimeout
}

func (h *Handler) pingInterval() time.Duration {
	if h.PingInterval > 0 {
		return h.PingInterval
	}
	return defaultPingInterval
}

func (h *Handler) pongTimeout() time.Duration {
	if h.PongTimeout > 0 {
		return h.PongTimeout
	}
	return defaultPongTimeout
}

func (h *Handler) maxMissedPings() int {
	if h.MaxMissedPings > 0 {
		return h.MaxMissedPings
	}
	return defaultMaxMissedPings
}

func (h *Handler) actionTimeout() time.Duration {
	if h.ActionTimeout > 0 {
		return h.ActionTimeout
	}
	return defaultActionTimeout
}

func (h *Handler) dedupTTL() time.Duration {
	if h.DedupTTL > 0 {
		return h.DedupTTL
	}
	return defaultDedupTTL
}

func (h *Handler) logf(format string, args ...any) {
	if h.Logger != nil {
		h.Logger.Info(fmt.Sprintf(format, args...))
	}
}

// findAction procura primeiro no catálogo curado do config.json e, se não
// achar, tenta resolver como um app vindo da varredura ao vivo de
// /Applications (id no formato "app:..." — PROJECT.md §7.4: mesmo aqui,
// nenhum caminho é aceito do cliente, só o ID que o próprio Mac gerou na
// última varredura). Resolve refaz a varredura, então um app desinstalado
// nunca chega a montar uma ação.
func (h *Handler) findAction(id string) (actions.Action, bool) {
	for _, a := range h.Actions {
		if a.ID == id {
			return a, true
		}
	}
	if h.Apps != nil {
		if app, ok := h.Apps.Resolve(id); ok {
			return actions.Action{
				ID:         app.ID,
				Label:      app.Name,
				Icon:       "app",
				Kind:       actions.KindOpenApp,
				Parameters: map[string]any{"application": app.Path},
			}, true
		}
	}
	return actions.Action{}, false
}

// ServeHTTP faz o upgrade e delega para handleConnection. Erros depois do
// upgrade só são logados — a resposta HTTP já foi enviada (é agora um
// WebSocket).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: corsorigins.Allowed})
	if err != nil {
		return
	}
	conn.SetReadLimit(int64(protocol.MaxWebSocketMessageBytes))
	defer conn.CloseNow()

	atomic.AddInt64(&h.activeConns, 1)
	defer atomic.AddInt64(&h.activeConns, -1)

	if err := h.handleConnection(r.Context(), conn); err != nil {
		h.logf("websocket: conexão encerrada: %v", err)
	}
}

// ActiveConnections retorna quantas conexões WebSocket estão abertas agora
// — usado pelo comando administrativo "status".
func (h *Handler) ActiveConnections() int {
	return int(atomic.LoadInt64(&h.activeConns))
}

func (h *Handler) handleConnection(ctx context.Context, conn *websocket.Conn) error {
	dev, err := h.authenticate(ctx, conn)
	if err != nil {
		return err
	}

	// Reconexão: manda o snapshot de estado logo depois da autenticação,
	// sem o cliente precisar pedir explicitamente (docs/PROTOCOL.md).
	if err := h.send(ctx, conn, protocol.TypeStateSnapshot, "", stateSnapshotPayload{
		MacName:      h.MacName,
		AgentVersion: h.AgentVersion,
	}); err != nil {
		return err
	}

	dedup := newDedupCache(h.dedupTTL())
	lastPong := newAtomicTime(time.Now())

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go h.heartbeatLoop(connCtx, conn, cancel, lastPong)

	for {
		var env protocol.Envelope
		if err := wsjson.Read(connCtx, conn, &env); err != nil {
			return fmt.Errorf("websocket[%s]: leitura encerrada: %w", dev.DeviceID, err)
		}
		h.dispatch(connCtx, conn, env, dedup, lastPong)
	}
}

func (h *Handler) authenticate(ctx context.Context, conn *websocket.Conn) (devices.Device, error) {
	authCtx, cancel := context.WithTimeout(ctx, h.authTimeout())
	defer cancel()

	var env protocol.Envelope
	if err := wsjson.Read(authCtx, conn, &env); err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "auth.hello não recebido a tempo")
		return devices.Device{}, fmt.Errorf("websocket: auth.hello não recebido: %w", err)
	}

	if env.Type != protocol.TypeAuthHello {
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrAuthRequired, "primeira mensagem deve ser auth.hello")
		_ = conn.Close(websocket.StatusPolicyViolation, "primeira mensagem deve ser auth.hello")
		return devices.Device{}, errors.New("websocket: primeira mensagem não foi auth.hello")
	}

	var hello authHelloPayload
	if err := json.Unmarshal(env.Payload, &hello); err != nil {
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrInvalidMessage, "payload de auth.hello inválido")
		_ = conn.Close(websocket.StatusPolicyViolation, "payload inválido")
		return devices.Device{}, fmt.Errorf("websocket: payload de auth.hello inválido: %w", err)
	}

	dev, ok := h.Devices.Find(hello.DeviceID)
	switch {
	case !ok:
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrAuthInvalid, "dispositivo desconhecido")
		_ = conn.Close(websocket.StatusPolicyViolation, "auth inválida")
		return devices.Device{}, errors.New("websocket: dispositivo desconhecido")
	case dev.Revoked:
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrAuthRevoked, "dispositivo revogado")
		_ = conn.Close(websocket.StatusPolicyViolation, "dispositivo revogado")
		return devices.Device{}, errors.New("websocket: dispositivo revogado")
	case !auth.TokensMatch(hello.AccessToken, dev.TokenHash):
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrAuthInvalid, "token inválido")
		_ = conn.Close(websocket.StatusPolicyViolation, "auth inválida")
		return devices.Device{}, errors.New("websocket: token inválido")
	}

	if err := h.send(ctx, conn, protocol.TypeAuthAccepted, env.RequestID, authAcceptedPayload{
		AgentVersion: h.AgentVersion,
		MacName:      h.MacName,
	}); err != nil {
		return devices.Device{}, err
	}
	return dev, nil
}

func (h *Handler) dispatch(ctx context.Context, conn *websocket.Conn, env protocol.Envelope, dedup *dedupCache, lastPong *atomicTime) {
	switch env.Type {
	case protocol.TypePong:
		lastPong.Set(time.Now())
	case protocol.TypeActionExecute:
		h.handleActionExecute(ctx, conn, env, dedup)
	default:
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrInvalidMessage, fmt.Sprintf("tipo de mensagem desconhecido: %s", env.Type))
	}
}

func (h *Handler) handleActionExecute(ctx context.Context, conn *websocket.Conn, env protocol.Envelope, dedup *dedupCache) {
	var payload actionExecutePayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrInvalidMessage, "payload de action.execute inválido")
		return
	}

	if cached, ok := dedup.get(env.RequestID); ok {
		_ = h.send(ctx, conn, protocol.TypeActionResult, env.RequestID, cached)
		return
	}

	action, ok := h.findAction(payload.ActionID)
	if !ok {
		_ = h.sendError(ctx, conn, env.RequestID, protocol.ErrActionNotFound, "ação desconhecida")
		return
	}

	_ = h.send(ctx, conn, protocol.TypeActionStarted, env.RequestID, actionStartedPayload{ActionID: action.ID})

	execCtx, cancel := context.WithTimeout(ctx, h.actionTimeout())
	result := h.Executor.Execute(execCtx, action)
	cancel()

	resultPayload := actionResultPayload{ActionID: action.ID, Status: result.Status, DurationMs: result.DurationMs}
	if result.ErrorCode != "" {
		resultPayload.ErrorCode = &result.ErrorCode
	}
	if result.Message != "" {
		resultPayload.Message = &result.Message
	}

	dedup.put(env.RequestID, resultPayload)
	_ = h.send(ctx, conn, protocol.TypeActionResult, env.RequestID, resultPayload)
}

// heartbeatLoop manda "ping" a cada pingInterval e fecha a conexão depois
// de maxMissedPings pongs perdidos em seguida (docs/PROTOCOL.md).
func (h *Handler) heartbeatLoop(ctx context.Context, conn *websocket.Conn, cancel context.CancelFunc, lastPong *atomicTime) {
	ticker := time.NewTicker(h.pingInterval())
	defer ticker.Stop()

	missed := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			beforePing := time.Now()
			if err := h.send(ctx, conn, protocol.TypePing, "", struct{}{}); err != nil {
				cancel()
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(h.pongTimeout()):
			}

			if lastPong.Get().Before(beforePing) {
				missed++
			} else {
				missed = 0
			}
			if missed >= h.maxMissedPings() {
				_ = conn.Close(websocket.StatusPolicyViolation, "heartbeat perdido")
				cancel()
				return
			}
		}
	}
}

func (h *Handler) send(ctx context.Context, conn *websocket.Conn, msgType, requestID string, payload any) error {
	return wsjson.Write(ctx, conn, newEnvelope(msgType, requestID, payload))
}

func (h *Handler) sendError(ctx context.Context, conn *websocket.Conn, requestID string, code protocol.ErrorCode, message string) error {
	return h.send(ctx, conn, protocol.TypeError, requestID, errorPayload{Code: code, Message: message})
}
