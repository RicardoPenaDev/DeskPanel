// Persistência do layout do painel (PROJECT.md §10.3) via Capacitor
// Preferences. Um layout ausente ou corrompido sempre cai de volta ao
// layout padrão em vez de travar o app (a Fase 4 tratará migração de
// schema com mais cuidado; por ora só validamos schemaVersion === 1).

import { Preferences } from "@capacitor/preferences";
import type { DashboardConfig } from "./layout";
import { defaultDashboardConfig } from "./defaultLayout";

const STORAGE_KEY = "deskpanel.layout";

function isValidLayout(value: unknown): value is DashboardConfig {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    v.schemaVersion === 1 && Array.isArray(v.profiles) && typeof v.activeProfileId === "string"
  );
}

export async function loadLayout(): Promise<DashboardConfig> {
  const { value } = await Preferences.get({ key: STORAGE_KEY });
  if (!value) return defaultDashboardConfig();

  try {
    const parsed = JSON.parse(value);
    return isValidLayout(parsed) ? parsed : defaultDashboardConfig();
  } catch {
    return defaultDashboardConfig();
  }
}

export async function saveLayout(config: DashboardConfig): Promise<void> {
  await Preferences.set({ key: STORAGE_KEY, value: JSON.stringify(config) });
}

export async function resetLayout(): Promise<DashboardConfig> {
  const fresh = defaultDashboardConfig();
  await saveLayout(fresh);
  return fresh;
}
