// Persistência do layout do painel (PROJECT.md §10.3, Fase 4 §17:
// "persistência e migração de schema") via Capacitor Preferences.
//
// Guarda duas cópias: a atual (`deskpanel.layout`) e a última cópia válida
// conhecida (`deskpanel.layout.backup`), atualizada só depois de uma
// gravação bem-sucedida. Se a cópia atual estiver corrompida ou tiver um
// schemaVersion desconhecido, tentamos a cópia de backup antes de cair no
// layout padrão — nunca lançamos exceção nem travamos o app por causa de
// um layout salvo inválido.
//
// CURRENT_SCHEMA_VERSION é o único valor aceito hoje; um schemaVersion
// futuro (2, 3, ...) entra aqui como um novo `case` em `migrate` sem
// alterar o resto do módulo.

import { Preferences } from "@capacitor/preferences";
import type { DashboardButton, DashboardConfig, DashboardPage, DashboardProfile } from "./layout";
import { defaultDashboardConfig } from "./defaultLayout";

const STORAGE_KEY = "deskpanel.layout";
const BACKUP_KEY = "deskpanel.layout.backup";
export const CURRENT_SCHEMA_VERSION = 1;

function isButton(value: unknown): value is DashboardButton {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.id === "string" &&
    typeof v.actionId === "string" &&
    typeof v.position === "number" &&
    typeof v.requireLongPress === "boolean"
  );
}

function isPage(value: unknown): value is DashboardPage {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.id === "string" &&
    typeof v.name === "string" &&
    typeof v.columns === "number" &&
    typeof v.rows === "number" &&
    Array.isArray(v.buttons) &&
    v.buttons.every(isButton)
  );
}

function isProfile(value: unknown): value is DashboardProfile {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.id === "string" &&
    typeof v.name === "string" &&
    Array.isArray(v.pages) &&
    v.pages.every(isPage)
  );
}

/**
 * Valida a forma de um layout de schemaVersion 1. Um schemaVersion
 * diferente (ou ausente) é tratado como incompatível — sem migração
 * definida ainda, cai para backup/padrão em vez de arriscar interpretar
 * um formato desconhecido.
 */
function isValidLayoutV1(value: unknown): value is DashboardConfig {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    v.schemaVersion === 1 &&
    typeof v.activeProfileId === "string" &&
    Array.isArray(v.profiles) &&
    v.profiles.length > 0 &&
    v.profiles.every(isProfile)
  );
}

async function readValidLayout(key: string): Promise<DashboardConfig | null> {
  const { value } = await Preferences.get({ key });
  if (!value) return null;

  try {
    const parsed = JSON.parse(value);
    return isValidLayoutV1(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

export async function loadLayout(): Promise<DashboardConfig> {
  const primary = await readValidLayout(STORAGE_KEY);
  if (primary) return primary;

  const backup = await readValidLayout(BACKUP_KEY);
  if (backup) return backup;

  return defaultDashboardConfig();
}

export async function saveLayout(config: DashboardConfig): Promise<void> {
  const serialized = JSON.stringify(config);
  await Preferences.set({ key: STORAGE_KEY, value: serialized });
  // só promove a backup depois que a gravação principal deu certo, e só se
  // o que estamos salvando é de fato válido — nunca fazemos backup de lixo.
  if (isValidLayoutV1(config)) {
    await Preferences.set({ key: BACKUP_KEY, value: serialized });
  }
}

export async function resetLayout(): Promise<DashboardConfig> {
  const fresh = defaultDashboardConfig();
  await saveLayout(fresh);
  return fresh;
}
