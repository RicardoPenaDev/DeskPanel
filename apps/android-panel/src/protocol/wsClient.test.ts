import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { RECONNECT_DELAYS_MS, WsClient, type WsClientHandlers } from "./wsClient";
import type { Envelope } from "./messages";

class FakeWebSocket {
  static instances: FakeWebSocket[] = [];
  url: string;
  sent: string[] = [];
  closed = false;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  onclose: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  send(data: string): void {
    this.sent.push(data);
  }

  close(): void {
    if (this.closed) return;
    this.closed = true;
    this.onclose?.();
  }

  triggerOpen(): void {
    this.onopen?.();
  }

  triggerMessage(
    payload: Omit<Envelope, "protocolVersion" | "timestamp"> & Partial<Envelope>,
  ): void {
    const envelope: Envelope = {
      protocolVersion: 1,
      timestamp: new Date().toISOString(),
      ...payload,
    } as Envelope;
    this.onmessage?.({ data: JSON.stringify(envelope) });
  }

  lastSent<T>(): (Envelope<T> & { type: string }) | undefined {
    const raw = this.sent.at(-1);
    return raw ? (JSON.parse(raw) as Envelope<T> & { type: string }) : undefined;
  }
}

function makeClient(handlers: WsClientHandlers = {}) {
  FakeWebSocket.instances = [];
  const client = new WsClient(
    {
      host: "192.168.1.50",
      port: 38121,
      deviceId: "device-1",
      accessToken: "tok",
      appVersion: "0.1.0",
    },
    handlers,
    (url: string) => new FakeWebSocket(url) as unknown as WebSocket,
  );
  return client;
}

describe("WsClient", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("envia auth.hello ao abrir e fica 'connected' após auth.accepted", () => {
    const statuses: string[] = [];
    const client = makeClient({ onStatusChange: (s) => statuses.push(s) });

    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();

    const hello = socket.lastSent<{ deviceId: string; accessToken: string }>();
    expect(hello?.type).toBe("auth.hello");
    expect(hello?.payload.deviceId).toBe("device-1");
    expect(hello?.payload.accessToken).toBe("tok");

    socket.triggerMessage({
      type: "auth.accepted",
      requestId: hello!.requestId,
      payload: { agentVersion: "0.1.0", macName: "Mac" },
    });

    expect(client.getStatus()).toBe("connected");
    expect(statuses).toEqual(["connecting", "authenticating", "connected"]);
  });

  it("nunca envia o token de acesso fora da mensagem auth.hello", () => {
    const client = makeClient();
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();
    socket.triggerMessage({
      type: "auth.accepted",
      requestId: "r1",
      payload: { agentVersion: "0.1.0", macName: "Mac" },
    });

    void client.executeAction("app.chrome");

    const lastMessage = socket.lastSent<{ actionId: string }>();
    expect(lastMessage?.type).toBe("action.execute");
    expect(JSON.stringify(lastMessage)).not.toContain("tok");
  });

  it("resolve executeAction quando chega action.result com o mesmo requestId", async () => {
    const client = makeClient();
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();
    socket.triggerMessage({
      type: "auth.accepted",
      requestId: "r1",
      payload: { agentVersion: "0.1.0", macName: "Mac" },
    });

    const promise = client.executeAction("app.chrome");
    const sent = socket.lastSent<{ actionId: string }>();

    socket.triggerMessage({
      type: "action.result",
      requestId: sent!.requestId,
      payload: {
        actionId: "app.chrome",
        status: "success",
        durationMs: 42,
        errorCode: null,
        message: null,
      },
    });

    await expect(promise).resolves.toEqual({
      actionId: "app.chrome",
      status: "success",
      durationMs: 42,
      errorCode: null,
      message: null,
    });
  });

  it("rejeita executeAction imediatamente quando não está conectado", async () => {
    const client = makeClient();
    await expect(client.executeAction("app.chrome")).rejects.toThrow("NOT_CONNECTED");
  });

  it("responde pong a um ping do servidor", () => {
    const client = makeClient();
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();
    socket.triggerMessage({
      type: "auth.accepted",
      requestId: "r1",
      payload: { agentVersion: "0.1.0", macName: "Mac" },
    });

    socket.sent = [];
    socket.triggerMessage({ type: "ping", requestId: "", payload: {} });

    const pong = socket.lastSent();
    expect(pong?.type).toBe("pong");
  });

  it("repassa state.snapshot para o handler", () => {
    const onStateSnapshot = vi.fn();
    const client = makeClient({ onStateSnapshot });
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();
    socket.triggerMessage({
      type: "state.snapshot",
      requestId: "",
      payload: { macName: "MacBook de Ricardo", agentVersion: "0.1.0" },
    });

    expect(onStateSnapshot).toHaveBeenCalledWith({
      macName: "MacBook de Ricardo",
      agentVersion: "0.1.0",
    });
  });

  it("reconecta com backoff após queda inesperada e reseta ao reconectar", () => {
    const statuses: string[] = [];
    const client = makeClient({ onStatusChange: (s) => statuses.push(s) });
    client.connect();
    const first = FakeWebSocket.instances[0];
    first.triggerOpen();
    first.triggerMessage({
      type: "auth.accepted",
      requestId: "r1",
      payload: { agentVersion: "0.1.0", macName: "Mac" },
    });

    first.close();
    expect(client.getStatus()).toBe("reconnecting");

    vi.advanceTimersByTime(RECONNECT_DELAYS_MS[0] + 500);
    expect(FakeWebSocket.instances.length).toBe(2);
  });

  it("não reconecta depois de disconnect() manual", () => {
    const client = makeClient();
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();

    client.disconnect();
    expect(client.getStatus()).toBe("disconnected");

    vi.advanceTimersByTime(20000);
    expect(FakeWebSocket.instances.length).toBe(1);
  });

  it("aciona onAuthError se auth.accepted não chegar a tempo", () => {
    const onAuthError = vi.fn();
    const client = makeClient({ onAuthError });
    client.connect();
    const socket = FakeWebSocket.instances[0];
    socket.triggerOpen();

    vi.advanceTimersByTime(6000);

    expect(onAuthError).toHaveBeenCalled();
    expect(socket.closed).toBe(true);
  });
});
