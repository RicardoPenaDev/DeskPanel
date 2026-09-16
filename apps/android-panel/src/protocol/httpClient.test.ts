import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  checkHealth,
  fetchActions,
  fetchAppIcon,
  fetchApps,
  fetchState,
  pairDevice,
  fetchVersion,
} from "./httpClient";

const endpoint = { host: "192.168.1.50", port: 38121 };

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("httpClient", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("retorna dados em uma checagem de saúde bem-sucedida", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(200, { status: "ok", service: "deskpanel-agent", protocolVersion: 1 }),
    );

    const result = await checkHealth(endpoint);

    expect(result).toEqual({
      ok: true,
      data: { status: "ok", service: "deskpanel-agent", protocolVersion: 1 },
    });
    expect(fetch).toHaveBeenCalledWith(
      "http://192.168.1.50:38121/api/v1/health",
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("mapeia uma resposta de erro do agente para o código estruturado", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(401, { error: { code: "AUTH_INVALID", message: "token inválido" } }),
    );

    const result = await fetchActions(endpoint, "algum-token");

    expect(result).toEqual({
      ok: false,
      error: { code: "AUTH_INVALID", message: "token inválido" },
    });
  });

  it("converte uma falha de rede em NETWORK_ERROR sem vazar a exceção", async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new TypeError("Failed to fetch"));

    const result = await fetchState(endpoint, "algum-token");

    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.error.code).toBe("NETWORK_ERROR");
    }
  });

  it("converte um abort de timeout em TIMEOUT", async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new DOMException("aborted", "AbortError"));

    const result = await fetchVersion(endpoint);

    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.error.code).toBe("TIMEOUT");
    }
  });

  it("envia o código de pareamento e a versão de protocolo corretamente", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(200, { deviceId: "device-1", accessToken: "tok", protocolVersion: 1 }),
    );

    await pairDevice(endpoint, {
      code: "483921",
      deviceId: "device-1",
      deviceName: "Moto G60 do Ricardo",
      appVersion: "0.1.0",
    });

    const [, init] = vi.mocked(fetch).mock.calls[0];
    const sentBody = JSON.parse((init as RequestInit).body as string);
    expect(sentBody).toEqual({
      code: "483921",
      deviceId: "device-1",
      deviceName: "Moto G60 do Ricardo",
      appVersion: "0.1.0",
      protocolVersion: 1,
    });
  });

  it("nunca inclui o token de acesso na URL", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(200, { macName: "Mac", agentVersion: "0.1.0" }),
    );

    await fetchState(endpoint, "segredo-super-secreto");

    const [url] = vi.mocked(fetch).mock.calls[0];
    expect(String(url)).not.toContain("segredo-super-secreto");
  });

  it("busca a lista de apps instalados com Bearer token", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(200, [{ id: "app:aaaa", name: "Notion", hasIcon: true }]),
    );

    const result = await fetchApps(endpoint, "algum-token");

    expect(result).toEqual({ ok: true, data: [{ id: "app:aaaa", name: "Notion", hasIcon: true }] });
    expect(fetch).toHaveBeenCalledWith(
      "http://192.168.1.50:38121/api/v1/apps",
      expect.objectContaining({ headers: { Authorization: "Bearer algum-token" } }),
    );
  });

  it("converte o ícone de um app em uma object URL, com o token no header e não na URL", async () => {
    // jsdom não implementa URL.createObjectURL — precisa ser criado e
    // restaurado manualmente em vez de vi.spyOn (que exige o método já
    // existir no objeto).
    const originalCreateObjectURL = URL.createObjectURL;
    URL.createObjectURL = vi.fn().mockReturnValue("blob:fake-icon-url");
    try {
      const blob = new Blob(["fake-png-bytes"], { type: "image/png" });
      vi.mocked(fetch).mockResolvedValueOnce(new Response(blob, { status: 200 }));

      const iconUrl = await fetchAppIcon(endpoint, "segredo-super-secreto", "app:aaaa");

      expect(iconUrl).toBe("blob:fake-icon-url");
      const [url, init] = vi.mocked(fetch).mock.calls[0];
      expect(String(url)).toBe("http://192.168.1.50:38121/api/v1/apps/app%3Aaaaa/icon");
      expect(String(url)).not.toContain("segredo-super-secreto");
      expect((init as RequestInit).headers).toEqual({
        Authorization: "Bearer segredo-super-secreto",
      });
    } finally {
      URL.createObjectURL = originalCreateObjectURL;
    }
  });

  it("fetchAppIcon retorna null em vez de lançar quando o ícone não existe", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 404 }));

    const iconUrl = await fetchAppIcon(endpoint, "algum-token", "app:desconhecido");

    expect(iconUrl).toBeNull();
  });

  it("fetchAppIcon retorna null em vez de lançar quando a rede falha", async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new TypeError("Failed to fetch"));

    const iconUrl = await fetchAppIcon(endpoint, "algum-token", "app:aaaa");

    expect(iconUrl).toBeNull();
  });
});
