import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  useDeskPanelConnection,
  type DeskPanelConnectionDeps,
  type WsClientLike,
} from "./useDeskPanelConnection";
import type { WsClientHandlers } from "../../protocol/wsClient";

function makeDeps(overrides: Partial<DeskPanelConnectionDeps> = {}) {
  const wsHandlers: Partial<WsClientHandlers> = {};
  const fakeWsClient: WsClientLike = {
    connect: vi.fn(),
    disconnect: vi.fn(),
    executeAction: vi.fn(async () => ({
      actionId: "app.chrome",
      status: "success",
      durationMs: 10,
      errorCode: null,
      message: null,
    })),
  };

  const deps: DeskPanelConnectionDeps = {
    loadConnectionConfig: vi.fn(async () => ({
      deviceId: "device-1",
      deviceName: "Moto G60 do Ricardo",
      host: "",
      port: 38121,
    })),
    saveConnectionConfig: vi.fn(async () => {}),
    readAccessToken: vi.fn(async () => null),
    saveAccessToken: vi.fn(async () => {}),
    loadLayout: vi.fn(async () => ({
      schemaVersion: 1 as const,
      activeProfileId: "default",
      profiles: [],
    })),
    saveLayout: vi.fn(async () => {}),
    resetLayout: vi.fn(async () => ({
      schemaVersion: 1 as const,
      activeProfileId: "default",
      profiles: [],
    })),
    checkHealth: vi.fn(async () => ({
      ok: true as const,
      data: { status: "ok", service: "deskpanel-agent", protocolVersion: 1 },
    })),
    pairDevice: vi.fn(async () => ({
      ok: true as const,
      data: { deviceId: "device-1", accessToken: "tok-123", protocolVersion: 1 },
    })),
    fetchActions: vi.fn(async () => ({ ok: true as const, data: [] })),
    createWsClient: vi.fn((_config, handlers: WsClientHandlers) => {
      Object.assign(wsHandlers, handlers);
      return fakeWsClient;
    }),
    ...overrides,
  };

  return { deps, wsHandlers, fakeWsClient };
}

describe("useDeskPanelConnection", () => {
  it("vai para 'pairing' quando não há token salvo", async () => {
    const { deps } = makeDeps();
    const { result } = renderHook(() => useDeskPanelConnection(deps));

    await waitFor(() => expect(result.current.phase).toBe("pairing"));
  });

  it("conecta automaticamente com token+host salvos e fica 'ready' após autenticar", async () => {
    const { deps, wsHandlers } = makeDeps({
      loadConnectionConfig: vi.fn(async () => ({
        deviceId: "device-1",
        deviceName: "Moto G60 do Ricardo",
        host: "192.168.1.50",
        port: 38121,
      })),
      readAccessToken: vi.fn(async () => "tok-existente"),
      fetchActions: vi.fn(async () => ({
        ok: true as const,
        data: [
          {
            id: "app.chrome",
            label: "Chrome",
            icon: "chrome",
            kind: "open_app",
            requireLongPress: false,
          },
        ],
      })),
    });

    const { result } = renderHook(() => useDeskPanelConnection(deps));

    await waitFor(() => expect(result.current.phase).toBe("connecting"));
    expect(deps.createWsClient).toHaveBeenCalled();

    act(() => {
      wsHandlers.onStatusChange?.("connected");
    });

    await waitFor(() => expect(result.current.phase).toBe("ready"));
    await waitFor(() => expect(result.current.actionsCatalog["app.chrome"]).toBeDefined());
  });

  it("repassa state.snapshot para macName/agentVersion", async () => {
    const { deps, wsHandlers } = makeDeps({
      loadConnectionConfig: vi.fn(async () => ({
        deviceId: "device-1",
        deviceName: "Moto G60",
        host: "192.168.1.50",
        port: 38121,
      })),
      readAccessToken: vi.fn(async () => "tok-existente"),
    });

    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("connecting"));

    act(() => {
      wsHandlers.onStateSnapshot?.({ macName: "MacBook de Ricardo", agentVersion: "0.1.0" });
    });

    await waitFor(() => expect(result.current.macName).toBe("MacBook de Ricardo"));
    expect(result.current.agentVersion).toBe("0.1.0");
  });

  it("pair() salva token e config e avança para 'connecting'", async () => {
    const { deps } = makeDeps();
    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("pairing"));

    let outcome: Awaited<ReturnType<typeof result.current.pair>> | undefined;
    await act(async () => {
      outcome = await result.current.pair({
        deviceName: "Moto G60",
        host: "192.168.1.50",
        port: 38121,
        code: "483921",
      });
    });

    expect(outcome).toEqual({ ok: true });
    expect(deps.saveAccessToken).toHaveBeenCalledWith("device-1", "tok-123");
    expect(deps.saveConnectionConfig).toHaveBeenCalledWith(
      expect.objectContaining({ host: "192.168.1.50", port: 38121 }),
    );
    await waitFor(() => expect(result.current.phase).toBe("connecting"));
  });

  it("pair() traduz PAIRING_CODE_INVALID em mensagem amigável e mantém 'pairing'", async () => {
    const { deps } = makeDeps({
      pairDevice: vi.fn(async () => ({
        ok: false as const,
        error: { code: "PAIRING_CODE_INVALID" as const, message: "" },
      })),
    });
    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("pairing"));

    let outcome: Awaited<ReturnType<typeof result.current.pair>> | undefined;
    await act(async () => {
      outcome = await result.current.pair({
        deviceName: "Moto G60",
        host: "192.168.1.50",
        port: 38121,
        code: "000000",
      });
    });

    expect(outcome?.ok).toBe(false);
    if (outcome && !outcome.ok) {
      expect(outcome.message).toContain("Código incorreto");
    }
    expect(result.current.phase).toBe("pairing");
  });

  it("executeAction delega para o wsClient quando conectado", async () => {
    const { deps, wsHandlers, fakeWsClient } = makeDeps({
      loadConnectionConfig: vi.fn(async () => ({
        deviceId: "device-1",
        deviceName: "Moto G60",
        host: "192.168.1.50",
        port: 38121,
      })),
      readAccessToken: vi.fn(async () => "tok-existente"),
    });

    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("connecting"));

    act(() => {
      wsHandlers.onStatusChange?.("connected");
    });
    await waitFor(() => expect(result.current.phase).toBe("ready"));

    await act(async () => {
      await result.current.executeAction("app.chrome");
    });

    expect(fakeWsClient.executeAction).toHaveBeenCalledWith("app.chrome");
  });

  it("executeAction rejeita quando não há conexão ativa", async () => {
    const { deps } = makeDeps();
    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("pairing"));

    await expect(result.current.executeAction("app.chrome")).rejects.toThrow("NOT_CONNECTED");
  });

  it("volta para 'pairing' e descarta o token quando o servidor revoga a autenticação", async () => {
    const { deps, wsHandlers } = makeDeps({
      loadConnectionConfig: vi.fn(async () => ({
        deviceId: "device-1",
        deviceName: "Moto G60",
        host: "192.168.1.50",
        port: 38121,
      })),
      readAccessToken: vi.fn(async () => "tok-revogado"),
    });

    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("connecting"));

    act(() => {
      wsHandlers.onAuthError?.("AUTH_REVOKED", "dispositivo revogado");
    });

    await waitFor(() => expect(result.current.phase).toBe("pairing"));
    expect(result.current.connectionError).toBe("dispositivo revogado");
  });

  it("updateLayout salva e atualiza o layout em memória", async () => {
    const { deps } = makeDeps();
    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("pairing"));

    const novoLayout = {
      schemaVersion: 1 as const,
      activeProfileId: "default",
      profiles: [{ id: "default", name: "Padrão", pages: [] }],
    };

    await act(async () => {
      await result.current.updateLayout(novoLayout);
    });

    expect(deps.saveLayout).toHaveBeenCalledWith(novoLayout);
    expect(result.current.layout).toEqual(novoLayout);
  });

  it("resetLayoutToDefault chama deps.resetLayout e atualiza o layout em memória", async () => {
    const layoutPadrao = {
      schemaVersion: 1 as const,
      activeProfileId: "default",
      profiles: [{ id: "default", name: "Padrão", pages: [] }],
    };
    const { deps } = makeDeps({ resetLayout: vi.fn(async () => layoutPadrao) });
    const { result } = renderHook(() => useDeskPanelConnection(deps));
    await waitFor(() => expect(result.current.phase).toBe("pairing"));

    await act(async () => {
      await result.current.resetLayoutToDefault();
    });

    expect(deps.resetLayout).toHaveBeenCalledTimes(1);
    expect(result.current.layout).toEqual(layoutPadrao);
  });
});
