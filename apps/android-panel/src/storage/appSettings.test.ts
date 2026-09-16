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

import { DEFAULT_APP_SETTINGS, loadAppSettings, saveAppSettings } from "./appSettings";

describe("appSettings", () => {
  beforeEach(() => {
    store.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("devolve o padrão quando nada foi salvo ainda", async () => {
    const settings = await loadAppSettings();
    expect(settings).toEqual(DEFAULT_APP_SETTINGS);
  });

  it("persiste e recarrega as mudanças", async () => {
    await saveAppSettings({
      vibrationEnabled: false,
      keepAwakeEnabled: true,
      immersiveModeEnabled: true,
      dimBrightnessEnabled: true,
    });

    const reloaded = await loadAppSettings();
    expect(reloaded).toEqual({
      vibrationEnabled: false,
      keepAwakeEnabled: true,
      immersiveModeEnabled: true,
      dimBrightnessEnabled: true,
    });
  });

  it("volta ao padrão quando o valor salvo está corrompido", async () => {
    store.set("deskpanel.appSettings", "{ isso não é json");
    const settings = await loadAppSettings();
    expect(settings).toEqual(DEFAULT_APP_SETTINGS);
  });

  it("volta ao padrão quando falta um campo esperado", async () => {
    store.set("deskpanel.appSettings", JSON.stringify({ vibrationEnabled: false }));
    const settings = await loadAppSettings();
    expect(settings).toEqual(DEFAULT_APP_SETTINGS);
  });
});
