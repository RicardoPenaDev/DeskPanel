import { describe, expect, it } from "vitest";
import { defaultDashboardConfig } from "./defaultLayout";

describe("defaultDashboardConfig", () => {
  it("gera duas páginas de 4x2 com ids de botão únicos", () => {
    const config = defaultDashboardConfig();
    expect(config.profiles).toHaveLength(1);

    const [profile] = config.profiles;
    expect(profile.pages).toHaveLength(2);

    const allButtonIds = profile.pages.flatMap((page) => page.buttons.map((b) => b.id));
    expect(new Set(allButtonIds).size).toBe(allButtonIds.length);

    for (const page of profile.pages) {
      expect(page.columns).toBe(4);
      expect(page.rows).toBe(2);
      expect(page.buttons).toHaveLength(8);
    }
  });

  it("marca apenas apagar monitor e bloquear Mac como toque prolongado", () => {
    const config = defaultDashboardConfig();
    const mediaPage = config.profiles[0].pages[1];
    const longPressIds = mediaPage.buttons.filter((b) => b.requireLongPress).map((b) => b.actionId);

    expect(longPressIds.sort()).toEqual(["system.display_sleep", "system.lock"]);
  });
});
