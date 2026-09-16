// Hook central que liga armazenamento local, cliente HTTP e cliente
// WebSocket em um único fluxo de conexão (PROJECT.md §10.2-A/B, §9). As
// dependências externas entram por parâmetro com um valor padrão real —
// mesmo padrão do Executor injetável usado no agente Go — para que os
// testes de componente não precisem mockar módulos inteiros.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  checkHealth,
  fetchActions,
  fetchAppIcon,
  fetchApps,
  pairDevice,
  type ActionSummary,
  type AppSummary,
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
import {
  clearAccessToken,
  readAccessToken,
  saveAccessToken,
} from "../../services/secureTokenStorage";
import {
  DEFAULT_APP_SETTINGS,
  loadAppSettings,
  saveAppSettings,
  type AppSettings,
} from "../../storage/appSettings";
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
  clearAccessToken: (deviceId: string) => Promise<void>;
  loadLayout: () => Promise<DashboardConfig>;
  saveLayout: (config: DashboardConfig) => Promise<void>;
  resetLayout: () => Promise<DashboardConfig>;
  loadAppSettings: () => Promise<AppSettings>;
  saveAppSettings: (settings: AppSettings) => Promise<void>;
  checkHealth: (endpoint: HttpEndpoint) => Promise<HttpResult<HealthResponse>>;
  pairDevice: (
    endpoint: HttpEndpoint,
    request: PairRequest & { code: string },
  ) => ReturnType<typeof pairDevice>;
  fetchActions: (
    endpoint: HttpEndpoint,
    accessToken: string,
  ) => Promise<HttpResult<ActionSummary[]>>;
  // Varredura ao vivo de /Applications no Mac — mesclada no mesmo
  // actionsCatalog para o editor deixar escolher qualquer app instalado,
  // não só os curados em config.json.
  fetchApps: (endpoint: HttpEndpoint, accessToken: string) => Promise<HttpResult<AppSummary[]>>;
  fetchAppIcon: (
    endpoint: HttpEndpoint,
    accessToken: string,
    appId: string,
  ) => Promise<string | null>;
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
  clearAccessToken,
  loadLayout,
  saveLayout,
  resetLayout,
  loadAppSettings,
  saveAppSettings,
  checkHealth,
  pairDevice: (endpoint, request) => pairDevice(endpoint, request),
  fetchActions,
  fetchApps,
  fetchAppIcon,
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
  appSettings: AppSettings;
  testConnection: (host: string, port: number) => Promise<HttpResult<HealthResponse>>;
  pair: (input: PairInput) => Promise<PairOutcome>;
  executeAction: (actionId: string) => Promise<ActionResult>;
  updateLayout: (next: DashboardConfig) => Promise<void>;
  resetLayoutToDefault: () => Promise<void>;
  updateAppSettings: (next: AppSettings) => Promise<void>;
  updateConnectionConfig: (next: ConnectionConfig) => Promise<void>;
  reconnect: () => void;
  forgetPairing: () => Promise<void>;
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
  const [appSettings, setAppSettings] = useState<AppSettings>(DEFAULT_APP_SETTINGS);

  const wsClientRef = useRef<WsClientLike | null>(null);

  // Carrega estado local (config de conexão, layout, token) uma única vez.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      const [config, currentLayout, settings] = await Promise.all([
        deps.loadConnectionConfig(),
        deps.loadLayout(),
        deps.loadAppSettings(),
      ]);
      if (cancelled) return;
      setConnectionConfig(config);
      setLayout(currentLayout);
      setAppSettings(settings);

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

  // Busca o catálogo autoritativo de ações do Mac assim que autenticado, e
  // mescla nele a varredura ao vivo de /Applications: cada app instalado
  // vira uma entrada open_app no mesmo catálogo, para o editor deixar
  // escolher qualquer um. As duas buscas terminam num único setState para
  // não haver corrida entre elas (uma nunca sobrescreve a outra).
  useEffect(() => {
    if (phase !== "ready" || !connectionConfig || !accessToken) return;
    let cancelled = false;
    const endpoint = { host: connectionConfig.host, port: connectionConfig.port };

    Promise.all([
      deps.fetchActions(endpoint, accessToken),
      deps.fetchApps(endpoint, accessToken),
    ]).then(([actionsResult, appsResult]) => {
      // Uma falha passageira de rede nunca deve apagar um catálogo que já
      // tinha carregado — sem isso, uma única requisição perdida deixava o
      // painel inteiro "indisponível" até o app ser reaberto (este efeito
      // só roda uma vez por conexão, nunca tenta de novo sozinho).
      if (cancelled || !actionsResult.ok) return;

      const catalog: Record<string, ActionSummary> = {};
      for (const action of actionsResult.data) catalog[action.id] = action;

      const apps = appsResult.ok ? appsResult.data : [];
      for (const app of apps) {
        catalog[app.id] = {
          id: app.id,
          label: app.name,
          icon: "app",
          kind: "open_app",
          requireLongPress: false,
        };
      }
      setActionsCatalog(catalog);

      // Ícones reais chegam depois, um a um, sem travar a exibição do
      // catálogo — cada app sem ícone continua com o ícone genérico.
      for (const app of apps) {
        if (!app.hasIcon) continue;
        deps.fetchAppIcon(endpoint, accessToken, app.id).then((iconUrl) => {
          if (cancelled || !iconUrl) return;
          setActionsCatalog((prev) => {
            const existing = prev[app.id];
            if (!existing) return prev;
            return { ...prev, [app.id]: { ...existing, iconUrl } };
          });
        });
      }
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

  const updateAppSettings = useCallback(async (next: AppSettings): Promise<void> => {
    await deps.saveAppSettings(next);
    setAppSettings(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Atualiza nome/host/porta a partir da tela de Configurações. Muda a
  // identidade de connectionConfig, o que já faz o efeito de WebSocket
  // reconectar sozinho no novo host/porta (mesmo deviceId/token — não
  // exige repareamento, só que seja o mesmo Mac já pareado).
  const updateConnectionConfig = useCallback(async (next: ConnectionConfig): Promise<void> => {
    await deps.saveConnectionConfig(next);
    setConnectionConfig(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Reconectar: reaproveita a mesma instância de WsClient (mesmo backoff,
  // mesmas credenciais) em vez de recriar tudo — útil quando o usuário quer
  // forçar uma nova tentativa na tela de Configurações sem esperar o
  // backoff automático.
  const reconnect = useCallback((): void => {
    const client = wsClientRef.current;
    if (!client) return;
    client.disconnect();
    client.connect();
  }, []);

  // Refazer pareamento: limpa só o token (mantém deviceId/host/porta/nome
  // já digitados) e volta para a tela de pareamento. O efeito de conexão
  // fecha o WebSocket sozinho quando accessToken vira null (dependência do
  // efeito).
  const forgetPairing = useCallback(async (): Promise<void> => {
    if (connectionConfig) {
      await deps.clearAccessToken(connectionConfig.deviceId);
    }
    setAccessToken(null);
    setConnectionError(null);
    setPhase("pairing");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connectionConfig]);

  return {
    phase,
    status,
    connectionConfig,
    layout,
    actionsCatalog,
    macName,
    agentVersion,
    connectionError,
    appSettings,
    testConnection,
    pair,
    executeAction,
    updateLayout,
    resetLayoutToDefault,
    updateAppSettings,
    updateConnectionConfig,
    reconnect,
    forgetPairing,
  };
}
