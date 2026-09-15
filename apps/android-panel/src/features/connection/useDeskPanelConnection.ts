// Hook central que liga armazenamento local, cliente HTTP e cliente
// WebSocket em um único fluxo de conexão (PROJECT.md §10.2-A/B, §9). As
// dependências externas entram por parâmetro com um valor padrão real —
// mesmo padrão do Executor injetável usado no agente Go — para que os
// testes de componente não precisem mockar módulos inteiros.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  checkHealth,
  fetchActions,
  pairDevice,
  type ActionSummary,
  type HealthResponse,
  type HttpEndpoint,
  type HttpResult,
  type PairRequest,
} from "../../protocol/httpClient";
import {
  loadOrCreateConnectionConfig,
  saveConnectionConfig,
  type ConnectionConfig,
} from "../../storage/connectionConfig";
import { loadLayout, resetLayout, saveLayout } from "../../storage/layoutStorage";
import { readAccessToken, saveAccessToken } from "../../services/secureTokenStorage";
import type { DashboardConfig } from "../../storage/layout";
import {
  WsClient,
  type ActionResult,
  type ConnectionStatus,
  type WsClientHandlers,
} from "../../protocol/wsClient";
import { APP_VERSION } from "../../version";

export type ConnectionPhase = "loading" | "pairing" | "connecting" | "ready";

export interface WsClientLike {
  connect(): void;
  disconnect(): void;
  executeAction(actionId: string): Promise<ActionResult>;
}

export interface PairInput {
  deviceName: string;
  host: string;
  port: number;
  code: string;
}

export type PairOutcome = { ok: true } | { ok: false; message: string };

export interface DeskPanelConnectionDeps {
  loadConnectionConfig: () => Promise<ConnectionConfig>;
  saveConnectionConfig: (config: ConnectionConfig) => Promise<void>;
  readAccessToken: (deviceId: string) => Promise<string | null>;
  saveAccessToken: (deviceId: string, token: string) => Promise<void>;
  loadLayout: () => Promise<DashboardConfig>;
  saveLayout: (config: DashboardConfig) => Promise<void>;
  resetLayout: () => Promise<DashboardConfig>;
  checkHealth: (endpoint: HttpEndpoint) => Promise<HttpResult<HealthResponse>>;
  pairDevice: (
    endpoint: HttpEndpoint,
    request: PairRequest & { code: string },
  ) => ReturnType<typeof pairDevice>;
  fetchActions: (
    endpoint: HttpEndpoint,
    accessToken: string,
  ) => Promise<HttpResult<ActionSummary[]>>;
  createWsClient: (
    config: {
      host: string;
      port: number;
      deviceId: string;
      accessToken: string;
      appVersion: string;
    },
    handlers: WsClientHandlers,
  ) => WsClientLike;
}

const defaultDeps: DeskPanelConnectionDeps = {
  loadConnectionConfig: loadOrCreateConnectionConfig,
  saveConnectionConfig,
  readAccessToken,
  saveAccessToken,
  loadLayout,
  saveLayout,
  resetLayout,
  checkHealth,
  pairDevice: (endpoint, request) => pairDevice(endpoint, request),
  fetchActions,
  createWsClient: (config, handlers) => new WsClient(config, handlers),
};

export interface UseDeskPanelConnectionResult {
  phase: ConnectionPhase;
  status: ConnectionStatus;
  connectionConfig: ConnectionConfig | null;
  layout: DashboardConfig | null;
  actionsCatalog: Record<string, ActionSummary>;
  macName: string | null;
  agentVersion: string | null;
  connectionError: string | null;
  testConnection: (host: string, port: number) => Promise<HttpResult<HealthResponse>>;
  pair: (input: PairInput) => Promise<PairOutcome>;
  executeAction: (actionId: string) => Promise<ActionResult>;
  updateLayout: (next: DashboardConfig) => Promise<void>;
  resetLayoutToDefault: () => Promise<void>;
}

function friendlyPairMessage(code: string, fallback: string): string {
  switch (code) {
    case "PAIRING_CLOSED":
      return "Nenhum pareamento está aberto no Mac agora. Rode 'deskpanel-agent pair' e tente novamente.";
    case "PAIRING_CODE_EXPIRED":
      return "O código expirou. Gere um novo código no Mac.";
    case "PAIRING_CODE_INVALID":
      return "Código incorreto. Confira os seis dígitos e tente de novo.";
    case "RATE_LIMITED":
      return "Muitas tentativas. Aguarde um pouco antes de tentar novamente.";
    case "PROTOCOL_UNSUPPORTED":
      return "Este app não é compatível com a versão do agente instalada no Mac.";
    case "NETWORK_ERROR":
      return "Não foi possível alcançar o Mac. Confira o IP, a porta e a rede Wi-Fi.";
    case "TIMEOUT":
      return "O Mac não respondeu a tempo. Confira se ele está ligado e na mesma rede.";
    default:
      return fallback || "Não foi possível parear com o Mac.";
  }
}

export function useDeskPanelConnection(
  overrides: Partial<DeskPanelConnectionDeps> = {},
): UseDeskPanelConnectionResult {
  const deps: DeskPanelConnectionDeps = { ...defaultDeps, ...overrides };

  const [phase, setPhase] = useState<ConnectionPhase>("loading");
  const [status, setStatus] = useState<ConnectionStatus>("disconnected");
  const [connectionConfig, setConnectionConfig] = useState<ConnectionConfig | null>(null);
  const [layout, setLayout] = useState<DashboardConfig | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [actionsCatalog, setActionsCatalog] = useState<Record<string, ActionSummary>>({});
  const [macName, setMacName] = useState<string | null>(null);
  const [agentVersion, setAgentVersion] = useState<string | null>(null);
  const [connectionError, setConnectionError] = useState<string | null>(null);

  const wsClientRef = useRef<WsClientLike | null>(null);

  // Carrega estado local (config de conexão, layout, token) uma única vez.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      const [config, currentLayout] = await Promise.all([
        deps.loadConnectionConfig(),
        deps.loadLayout(),
      ]);
      if (cancelled) return;
      setConnectionConfig(config);
      setLayout(currentLayout);

      const token = await deps.readAccessToken(config.deviceId);
      if (cancelled) return;

      if (token && config.host) {
        setAccessToken(token);
        setPhase("connecting");
      } else {
        setPhase("pairing");
      }
    })();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Abre o WebSocket assim que houver token + config; fecha ao desmontar
  // ou quando o dispositivo precisar ser pareado de novo.
  useEffect(() => {
    if (!connectionConfig || !accessToken) return;

    const client = deps.createWsClient(
      {
        host: connectionConfig.host,
        port: connectionConfig.port,
        deviceId: connectionConfig.deviceId,
        accessToken,
        appVersion: APP_VERSION,
      },
      {
        onStatusChange: (nextStatus) => {
          setStatus(nextStatus);
          if (nextStatus === "connected") {
            setConnectionError(null);
            setPhase("ready");
          }
        },
        onStateSnapshot: (state) => {
          setMacName(state.macName);
          setAgentVersion(state.agentVersion);
        },
        onStateChanged: (state) => {
          if (state.macName) setMacName(state.macName);
          if (state.agentVersion) setAgentVersion(state.agentVersion);
        },
        onAuthError: (code, message) => {
          setConnectionError(message || `Falha de autenticação (${code})`);
          if (code === "AUTH_REVOKED" || code === "AUTH_INVALID") {
            setAccessToken(null);
            setPhase("pairing");
          }
        },
      },
    );

    wsClientRef.current = client;
    client.connect();

    return () => {
      client.disconnect();
      wsClientRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectionConfig, accessToken]);

  // Busca o catálogo autoritativo de ações do Mac assim que autenticado.
  useEffect(() => {
    if (phase !== "ready" || !connectionConfig || !accessToken) return;
    let cancelled = false;

    deps
      .fetchActions({ host: connectionConfig.host, port: connectionConfig.port }, accessToken)
      .then((result) => {
        if (cancelled || !result.ok) return;
        const catalog: Record<string, ActionSummary> = {};
        for (const action of result.data) catalog[action.id] = action;
        setActionsCatalog(catalog);
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [phase, connectionConfig, accessToken]);

  const testConnection = useCallback(
    (host: string, port: number) => deps.checkHealth({ host, port }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const pair = useCallback(
    async (input: PairInput): Promise<PairOutcome> => {
      if (!connectionConfig) {
        return { ok: false, message: "Configuração de conexão ainda não carregada" };
      }

      const endpoint: HttpEndpoint = { host: input.host, port: input.port };
      const result = await deps.pairDevice(endpoint, {
        code: input.code,
        deviceId: connectionConfig.deviceId,
        deviceName: input.deviceName,
        appVersion: APP_VERSION,
      });

      if (!result.ok) {
        return { ok: false, message: friendlyPairMessage(result.error.code, result.error.message) };
      }

      const updatedConfig: ConnectionConfig = {
        ...connectionConfig,
        deviceName: input.deviceName,
        host: input.host,
        port: input.port,
      };

      await deps.saveConnectionConfig(updatedConfig);
      await deps.saveAccessToken(connectionConfig.deviceId, result.data.accessToken);

      setConnectionConfig(updatedConfig);
      setAccessToken(result.data.accessToken);
      setConnectionError(null);
      setPhase("connecting");
      return { ok: true };
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [connectionConfig],
  );

  const executeAction = useCallback((actionId: string): Promise<ActionResult> => {
    const client = wsClientRef.current;
    if (!client) return Promise.reject(new Error("NOT_CONNECTED"));
    return client.executeAction(actionId);
  }, []);

  const updateLayout = useCallback(
    async (next: DashboardConfig): Promise<void> => {
      await deps.saveLayout(next);
      setLayout(next);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const resetLayoutToDefault = useCallback(async (): Promise<void> => {
    const fresh = await deps.resetLayout();
    setLayout(fresh);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    phase,
    status,
    connectionConfig,
    layout,
    actionsCatalog,
    macName,
    agentVersion,
    connectionError,
    testConnection,
    pair,
    executeAction,
    updateLayout,
    resetLayoutToDefault,
  };
}
