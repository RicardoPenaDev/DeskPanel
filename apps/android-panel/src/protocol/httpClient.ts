// Cliente HTTP para a API do DeskPanel Agent (PROJECT.md §8, docs/PROTOCOL.md).
// Cobre apenas os endpoints REST auxiliares — a execução de ações em tempo
// real acontece pelo WebSocket (wsClient.ts). Nenhuma função aqui aceita
// comando, caminho ou URL arbitrários vindos de fora do dispositivo: os
// parâmetros são sempre host/porta configurados pelo usuário e dados de
// pareamento digitados por ele.

import { PROTOCOL_VERSION, type ErrorCode } from "./messages";

export interface HttpEndpoint {
  host: string;
  port: number;
}

export interface ActionSummary {
  id: string;
  label: string;
  icon: string;
  kind: string;
  requireLongPress: boolean;
  // Preenchido no cliente (não vem do /actions) quando esta entrada é um
  // app da varredura ao vivo de /Applications — ver fetchAppIcon.
  iconUrl?: string;
}

export interface AppSummary {
  id: string;
  name: string;
  hasIcon: boolean;
}

export interface StateSnapshot {
  macName: string;
  agentVersion: string;
}

export interface WeatherSnapshot {
  city: string;
  tempC: number;
  description: string;
  updatedAt: string;
}

export interface HealthResponse {
  status: string;
  service: string;
  protocolVersion: number;
}

export interface VersionResponse {
  agentVersion: string;
  protocolVersion: number;
}

export interface PairRequest {
  code: string;
  deviceId: string;
  deviceName: string;
  appVersion: string;
}

export interface PairResponse {
  deviceId: string;
  accessToken: string;
  protocolVersion: number;
}

export type HttpFailureCode = ErrorCode | "NETWORK_ERROR" | "TIMEOUT" | "INVALID_RESPONSE";

export interface HttpFailure {
  code: HttpFailureCode;
  message: string;
}

export type HttpResult<T> = { ok: true; data: T } | { ok: false; error: HttpFailure };

const DEFAULT_TIMEOUT_MS = 5000;

function baseUrl(endpoint: HttpEndpoint): string {
  return `http://${endpoint.host}:${endpoint.port}/api/v1`;
}

/**
 * fetchJson centraliza timeout, parsing de erro e conversão de exceções de
 * rede em um HttpResult tipado, para que as telas nunca precisem lidar com
 * exceções cruas de `fetch`.
 */
async function fetchJson<T>(
  url: string,
  init: RequestInit,
  timeoutMs: number = DEFAULT_TIMEOUT_MS,
): Promise<HttpResult<T>> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, { ...init, signal: controller.signal });
    const text = await response.text();
    const body = text.length > 0 ? safeParse(text) : undefined;

    if (!response.ok) {
      const code = errorCodeFromBody(body) ?? "INTERNAL_ERROR";
      const message = messageFromBody(body) ?? `Falha HTTP ${response.status}`;
      return { ok: false, error: { code, message } };
    }

    if (body === undefined) {
      return {
        ok: false,
        error: { code: "INVALID_RESPONSE", message: "Resposta vazia do agente" },
      };
    }

    return { ok: true, data: body as T };
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      return { ok: false, error: { code: "TIMEOUT", message: "Tempo esgotado ao contatar o Mac" } };
    }
    return {
      ok: false,
      error: { code: "NETWORK_ERROR", message: "Não foi possível alcançar o Mac na rede local" },
    };
  } finally {
    clearTimeout(timer);
  }
}

function safeParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

function errorCodeFromBody(body: unknown): ErrorCode | undefined {
  if (body && typeof body === "object" && "error" in body) {
    const err = (body as { error?: unknown }).error;
    if (err && typeof err === "object" && "code" in err) {
      const code = (err as { code?: unknown }).code;
      if (typeof code === "string") return code as ErrorCode;
    }
  }
  return undefined;
}

function messageFromBody(body: unknown): string | undefined {
  if (body && typeof body === "object" && "error" in body) {
    const err = (body as { error?: unknown }).error;
    if (err && typeof err === "object" && "message" in err) {
      const message = (err as { message?: unknown }).message;
      if (typeof message === "string") return message;
    }
  }
  return undefined;
}

export function checkHealth(endpoint: HttpEndpoint): Promise<HttpResult<HealthResponse>> {
  return fetchJson<HealthResponse>(`${baseUrl(endpoint)}/health`, { method: "GET" }, 3000);
}

export function fetchVersion(endpoint: HttpEndpoint): Promise<HttpResult<VersionResponse>> {
  return fetchJson<VersionResponse>(`${baseUrl(endpoint)}/version`, { method: "GET" }, 3000);
}

export function pairDevice(
  endpoint: HttpEndpoint,
  request: PairRequest,
): Promise<HttpResult<PairResponse>> {
  return fetchJson<PairResponse>(`${baseUrl(endpoint)}/pair`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...request, protocolVersion: PROTOCOL_VERSION }),
  });
}

export function fetchActions(
  endpoint: HttpEndpoint,
  accessToken: string,
): Promise<HttpResult<ActionSummary[]>> {
  return fetchJson<ActionSummary[]>(`${baseUrl(endpoint)}/actions`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export function fetchState(
  endpoint: HttpEndpoint,
  accessToken: string,
): Promise<HttpResult<StateSnapshot>> {
  return fetchJson<StateSnapshot>(`${baseUrl(endpoint)}/state`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * fetchWeather busca o clima atual (geolocalizado pelo próprio Mac) para
 * a tela ambiente do Android. Sem internet no Mac ou geolocalização
 * recusada vira HttpResult com ok:false — o relógio ambiente só não
 * mostra o clima, nunca quebra a tela.
 */
export function fetchWeather(
  endpoint: HttpEndpoint,
  accessToken: string,
): Promise<HttpResult<WeatherSnapshot>> {
  return fetchJson<WeatherSnapshot>(`${baseUrl(endpoint)}/weather`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * fetchApps busca a varredura ao vivo de aplicativos instalados no Mac
 * (/Applications e afins) — a lista completa que o editor oferece para
 * criar um atalho, além do catálogo curado de /actions.
 */
export function fetchApps(
  endpoint: HttpEndpoint,
  accessToken: string,
): Promise<HttpResult<AppSummary[]>> {
  return fetchJson<AppSummary[]>(`${baseUrl(endpoint)}/apps`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

/**
 * fetchAppIcon busca o PNG do ícone real de um app e devolve uma object
 * URL local pronta para <img src>. Não usa fetchJson (resposta binária,
 * não JSON) nem token na URL (o <img> nunca vê o Bearer) — falha vira
 * `null` em vez de lançar, já que um ícone ausente nunca deve quebrar a
 * tela.
 */
export async function fetchAppIcon(
  endpoint: HttpEndpoint,
  accessToken: string,
  appId: string,
): Promise<string | null> {
  try {
    const response = await fetch(`${baseUrl(endpoint)}/apps/${encodeURIComponent(appId)}/icon`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    if (!response.ok) return null;
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  } catch {
    return null;
  }
}
