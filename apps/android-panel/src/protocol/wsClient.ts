// Cliente WebSocket do DeskPanel (PROJECT.md §9, docs/PROTOCOL.md). Espelha
// apps/mac-agent/internal/websocket/handler.go: autenticação por
// auth.hello/auth.accepted, execução de ações por requestId, heartbeat
// ping/pong e reconexão com backoff 1s/2s/4s/8s/15s. Nunca envia nada além
// de actionId — nenhum comando, caminho ou parâmetro livre.

import { PROTOCOL_VERSION, type Envelope, type ErrorCode, type MessageType } from "./messages";
import type { StateSnapshot } from "./httpClient";

export interface WsClientConfig {
  host: string;
  port: number;
  deviceId: string;
  accessToken: string;
  appVersion: string;
}

export type ConnectionStatus =
  "disconnected" | "connecting" | "authenticating" | "connected" | "reconnecting";

export interface ActionResult {
  actionId: string;
  status: string;
  durationMs: number;
  errorCode: string | null;
  message: string | null;
}

export interface WsClientHandlers {
  onStatusChange?: (status: ConnectionStatus) => void;
  onStateSnapshot?: (state: StateSnapshot) => void;
  onStateChanged?: (state: Partial<StateSnapshot>) => void;
  onActionStarted?: (actionId: string) => void;
  onAuthError?: (code: ErrorCode | "UNKNOWN", message: string) => void;
}

export type WebSocketFactory = (url: string) => WebSocket;

/** Backoff de reconexão fixo pedido pelo PROJECT.md §9.5. Jitter pequeno
 * evita que múltiplos clientes (improvável aqui, mas barato de evitar)
 * reconectem exatamente no mesmo instante. */
export const RECONNECT_DELAYS_MS = [1000, 2000, 4000, 8000, 15000];
const RECONNECT_JITTER_MS = 250;
const AUTH_TIMEOUT_MS = 5000;
// Levemente acima do timeout de ação do agente (10s, PROJECT.md §7.5) para
// que o agente tenha chance de responder ACTION_TIMEOUT antes do cliente
// desistir localmente.
const ACTION_TIMEOUT_MS = 12000;

interface PendingAction {
  resolve: (result: ActionResult) => void;
  reject: (err: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

function generateRequestId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  // Fallback só para ambientes sem crypto.randomUUID (não usado em produção
  // no WebView do Android, mas mantém o código robusto).
  return `req-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function buildEnvelope<T>(type: MessageType, requestId: string, payload: T): Envelope<T> {
  return {
    type,
    protocolVersion: PROTOCOL_VERSION,
    requestId,
    timestamp: new Date().toISOString(),
    payload,
  };
}

export class WsClient {
  private socket: WebSocket | null = null;
  private status: ConnectionStatus = "disconnected";
  private manuallyClosed = false;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private authTimer: ReturnType<typeof setTimeout> | null = null;
  private pendingActions = new Map<string, PendingAction>();

  constructor(
    private readonly config: WsClientConfig,
    private readonly handlers: WsClientHandlers = {},
    private readonly wsFactory: WebSocketFactory = (url) => new WebSocket(url),
  ) {}

  getStatus(): ConnectionStatus {
    return this.status;
  }

  connect(): void {
    this.manuallyClosed = false;
    this.reconnectAttempt = 0;
    this.openSocket();
  }

  disconnect(): void {
    this.manuallyClosed = true;
    this.clearReconnectTimer();
    this.clearAuthTimer();
    this.rejectAllPending(new Error("Conexão encerrada pelo usuário"));
    if (this.socket) {
      this.socket.onopen = null;
      this.socket.onmessage = null;
      this.socket.onerror = null;
      this.socket.onclose = null;
      this.socket.close();
      this.socket = null;
    }
    this.setStatus("disconnected");
  }

  executeAction(actionId: string): Promise<ActionResult> {
    if (this.status !== "connected" || !this.socket) {
      return Promise.reject(new Error("NOT_CONNECTED"));
    }
    const requestId = generateRequestId();
    const envelope = buildEnvelope("action.execute", requestId, { actionId });

    return new Promise<ActionResult>((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pendingActions.delete(requestId);
        reject(new Error("ACTION_TIMEOUT"));
      }, ACTION_TIMEOUT_MS);
      this.pendingActions.set(requestId, { resolve, reject, timer });
      this.socket!.send(JSON.stringify(envelope));
    });
  }

  private setStatus(status: ConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    this.handlers.onStatusChange?.(status);
  }

  private wsUrl(): string {
    return `ws://${this.config.host}:${this.config.port}/api/v1/ws`;
  }

  private openSocket(): void {
    this.clearReconnectTimer();
    this.setStatus(this.reconnectAttempt > 0 ? "reconnecting" : "connecting");

    let socket: WebSocket;
    try {
      socket = this.wsFactory(this.wsUrl());
    } catch {
      this.scheduleReconnect();
      return;
    }
    this.socket = socket;

    socket.onopen = () => this.handleOpen();
    socket.onmessage = (event: MessageEvent) => this.handleMessage(event);
    socket.onerror = () => {
      /* o evento close subsequente carrega o motivo tratável */
    };
    socket.onclose = () => this.handleClose();
  }

  private handleOpen(): void {
    this.setStatus("authenticating");
    const requestId = generateRequestId();
    const hello = buildEnvelope("auth.hello", requestId, {
      deviceId: this.config.deviceId,
      accessToken: this.config.accessToken,
      appVersion: this.config.appVersion,
    });
    this.socket?.send(JSON.stringify(hello));

    this.authTimer = setTimeout(() => {
      this.handlers.onAuthError?.("UNKNOWN", "Sem resposta de autenticação do Mac");
      this.socket?.close();
    }, AUTH_TIMEOUT_MS);
  }

  private clearAuthTimer(): void {
    if (this.authTimer) {
      clearTimeout(this.authTimer);
      this.authTimer = null;
    }
  }

  private handleMessage(event: MessageEvent): void {
    let envelope: Envelope;
    try {
      envelope = JSON.parse(String(event.data)) as Envelope;
    } catch {
      return;
    }

    switch (envelope.type) {
      case "auth.accepted":
        this.clearAuthTimer();
        this.reconnectAttempt = 0;
        this.setStatus("connected");
        break;
      case "ping":
        this.sendPong();
        break;
      case "action.started": {
        const payload = envelope.payload as { actionId: string };
        this.handlers.onActionStarted?.(payload.actionId);
        break;
      }
      case "action.result":
        this.resolvePending(envelope.requestId, envelope.payload as ActionResult);
        break;
      case "state.snapshot":
        this.handlers.onStateSnapshot?.(envelope.payload as StateSnapshot);
        break;
      case "state.changed":
        this.handlers.onStateChanged?.(envelope.payload as Partial<StateSnapshot>);
        break;
      case "error":
        this.handleError(envelope);
        break;
      default:
        break;
    }
  }

  private sendPong(): void {
    const pong = buildEnvelope("pong", "", {});
    this.socket?.send(JSON.stringify(pong));
  }

  private resolvePending(requestId: string, result: ActionResult): void {
    const pending = this.pendingActions.get(requestId);
    if (!pending) return;
    clearTimeout(pending.timer);
    this.pendingActions.delete(requestId);
    pending.resolve(result);
  }

  private handleError(envelope: Envelope): void {
    const payload = envelope.payload as { code: ErrorCode; message: string };
    const pending = this.pendingActions.get(envelope.requestId);
    if (pending) {
      clearTimeout(pending.timer);
      this.pendingActions.delete(envelope.requestId);
      pending.reject(new Error(payload.code));
      return;
    }
    if (this.status !== "connected") {
      this.clearAuthTimer();
      this.handlers.onAuthError?.(payload.code, payload.message);
    }
  }

  private handleClose(): void {
    this.clearAuthTimer();
    this.socket = null;
    this.rejectAllPending(new Error("Conexão perdida"));
    if (this.manuallyClosed) {
      this.setStatus("disconnected");
      return;
    }
    this.scheduleReconnect();
  }

  private rejectAllPending(err: Error): void {
    for (const pending of this.pendingActions.values()) {
      clearTimeout(pending.timer);
      pending.reject(err);
    }
    this.pendingActions.clear();
  }

  private scheduleReconnect(): void {
    if (this.manuallyClosed) return;
    this.setStatus("reconnecting");
    const index = Math.min(this.reconnectAttempt, RECONNECT_DELAYS_MS.length - 1);
    const base = RECONNECT_DELAYS_MS[index];
    const jitter = Math.floor(Math.random() * RECONNECT_JITTER_MS);
    this.reconnectAttempt++;
    this.reconnectTimer = setTimeout(() => this.openSocket(), base + jitter);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}
