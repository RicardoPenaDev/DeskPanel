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

import { loadLayout, resetLayout, saveLayout } from "./layoutStorage";
import { defaultDashboardConfig } from "./defaultLayout";

describe("layoutStorage", () => {
  beforeEach(() => {
    store.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("retorna o layout padrão quando nada foi salvo", async () => {
    const layout = await loadLayout();
    expect(layout).toEqual(defaultDashboardConfig());
  });

  it("persiste e recarrega um layout modificado", async () => {
    const layout = defaultDashboardConfig();
    layout.profiles[0].pages[0].buttons[0].labelOverride = "Google Chrome";
    await saveLayout(layout);

    const reloaded = await loadLayout();
    expect(reloaded.profiles[0].pages[0].buttons[0].labelOverride).toBe("Google Chrome");
  });

  it("volta ao padrão quando o schema salvo é incompatível", async () => {
    store.set("deskpanel.layout", JSON.stringify({ schemaVersion: 2, profiles: [] }));
    const layout = await loadLayout();
    expect(layout).toEqual(defaultDashboardConfig());
  });

  it("volta ao padrão quando o JSON salvo está corrompido", async () => {
    store.set("deskpanel.layout", "{ inválido");
    const layout = await loadLayout();
    expect(layout).toEqual(defaultDashboardConfig());
  });

  it("resetLayout restaura e persiste o padrão", async () => {
    await saveLayout({ ...defaultDashboardConfig(), activeProfileId: "outro" });
    const restored = await resetLayout();
    expect(restored).toEqual(defaultDashboardConfig());

    const reloaded = await loadLayout();
    expect(reloaded).toEqual(defaultDashboardConfig());
  });
});
