import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const store = new Map<string, string>();

vi.mock("@capacitor/preferences", () => ({
  Preferences: {
    get: vi.fn(async ({ key }: { key: string }) => ({ value: store.get(key) ?? null })),
    set: vi.fn(async ({ key, value }: { key: string; value: string }) => {
      store.set(key, value);
    }),
    remove: vi.fn(async ({ key }: { key: string }) => {
      store.delete(key);
    }),
  },
}));

import {
  DEFAULT_DEVICE_NAME,
  DEFAULT_PORT,
  clearConnectionConfig,
  loadOrCreateConnectionConfig,
  saveConnectionConfig,
} from "./connectionConfig";

describe("connectionConfig", () => {
  beforeEach(() => {
    store.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("cria e persiste uma configuração padrão quando nada foi salvo", async () => {
    const config = await loadOrCreateConnectionConfig();

    expect(config.deviceName).toBe(DEFAULT_DEVICE_NAME);
    expect(config.port).toBe(DEFAULT_PORT);
    expect(config.host).toBe("");
    expect(config.deviceId).toMatch(/.+/);

    const second = await loadOrCreateConnectionConfig();
    expect(second.deviceId).toBe(config.deviceId);
  });

  it("mantém o mesmo deviceId entre reinícios simulados", async () => {
    const first = await loadOrCreateConnectionConfig();
    await saveConnectionConfig({ ...first, host: "192.168.1.50" });

    const reloaded = await loadOrCreateConnectionConfig();
    expect(reloaded.deviceId).toBe(first.deviceId);
    expect(reloaded.host).toBe("192.168.1.50");
  });

  it("volta ao padrão quando o valor salvo está corrompido", async () => {
    store.set("deskpanel.connectionConfig", "{ isso não é json");

    const config = await loadOrCreateConnectionConfig();
    expect(config.deviceName).toBe(DEFAULT_DEVICE_NAME);
  });

  it("nunca persiste um token de acesso junto da configuração de conexão", async () => {
    const config = await loadOrCreateConnectionConfig();
    await saveConnectionConfig({ ...config, host: "192.168.1.50" });

    const raw = store.get("deskpanel.connectionConfig") ?? "";
    expect(raw).not.toContain("accessToken");
    expect(raw).not.toContain("token");
  });

  it("remove a configuração ao limpar", async () => {
    await loadOrCreateConnectionConfig();
    await clearConnectionConfig();
    expect(store.has("deskpanel.connectionConfig")).toBe(false);
  });
});
