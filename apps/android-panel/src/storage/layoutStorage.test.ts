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

  it("volta ao padrão quando o schema salvo é incompatível e não há backup", async () => {
    store.set("deskpanel.layout", JSON.stringify({ schemaVersion: 2, profiles: [] }));
    const layout = await loadLayout();
    expect(layout).toEqual(defaultDashboardConfig());
  });

  it("volta ao padrão quando o JSON salvo está corrompido e não há backup", async () => {
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

  it("cada saveLayout bem-sucedido atualiza o backup", async () => {
    const first = defaultDashboardConfig();
    first.profiles[0].pages[0].buttons[0].labelOverride = "Versão 1";
    await saveLayout(first);

    const second = { ...first, activeProfileId: first.activeProfileId };
    second.profiles[0].pages[0].buttons[0].labelOverride = "Versão 2";
    await saveLayout(second);

    expect(store.get("deskpanel.layout.backup")).toContain("Versão 2");
  });

  it("usa o backup quando a cópia principal fica corrompida", async () => {
    const good = defaultDashboardConfig();
    good.profiles[0].pages[0].buttons[0].labelOverride = "Backup bom";
    await saveLayout(good);

    // corrompe só a cópia principal, sem tocar no backup
    store.set("deskpanel.layout", "{ isso quebrou");

    const layout = await loadLayout();
    expect(layout.profiles[0].pages[0].buttons[0].labelOverride).toBe("Backup bom");
  });

  it("usa o backup quando a cópia principal tem schemaVersion incompatível", async () => {
    const good = defaultDashboardConfig();
    await saveLayout(good);

    store.set("deskpanel.layout", JSON.stringify({ schemaVersion: 99, profiles: [] }));

    const layout = await loadLayout();
    expect(layout).toEqual(good);
  });

  it("nunca promove um layout inválido para o backup", async () => {
    const good = defaultDashboardConfig();
    good.profiles[0].pages[0].buttons[0].labelOverride = "Bom";
    await saveLayout(good);

    // @ts-expect-error -- simula um bug interno tentando salvar algo inválido
    await saveLayout({ schemaVersion: 1, activeProfileId: "x" });

    const backup = JSON.parse(store.get("deskpanel.layout.backup") ?? "{}");
    expect(backup.profiles?.[0]?.pages?.[0]?.buttons?.[0]?.labelOverride).toBe("Bom");
  });
});
