// Preferências não secretas de comportamento do app (PROJECT.md §10.2-D):
// vibração, manter tela ligada, modo imersivo, brilho reduzido. Mesmo
// padrão de storage.Preferences já usado em connectionConfig.ts/
// layoutStorage.ts — nada aqui é sensível, então não passa pelo Keystore.

import { Preferences } from "@capacitor/preferences";

export interface AppSettings {
  vibrationEnabled: boolean;
  keepAwakeEnabled: boolean;
  immersiveModeEnabled: boolean;
  dimBrightnessEnabled: boolean;
}

const STORAGE_KEY = "deskpanel.appSettings";

// Vibração e "manter tela ligada" começam habilitadas (PROJECT.md §10.1:
// "keep awake" é comportamento padrão do painel; vibração é "opcional" mas
// o padrão dela ligada casa com a entrega da Fase 3/6). Imersivo e brilho
// reduzido começam desligados — são opt-in.
export const DEFAULT_APP_SETTINGS: AppSettings = {
  vibrationEnabled: true,
  keepAwakeEnabled: true,
  immersiveModeEnabled: false,
  dimBrightnessEnabled: false,
};

function isValidSettings(value: unknown): value is AppSettings {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.vibrationEnabled === "boolean" &&
    typeof v.keepAwakeEnabled === "boolean" &&
    typeof v.immersiveModeEnabled === "boolean" &&
    typeof v.dimBrightnessEnabled === "boolean"
  );
}

export async function loadAppSettings(): Promise<AppSettings> {
  const { value } = await Preferences.get({ key: STORAGE_KEY });
  if (value) {
    try {
      const parsed = JSON.parse(value);
      if (isValidSettings(parsed)) return parsed;
    } catch {
      // cai para o padrão abaixo
    }
  }
  return DEFAULT_APP_SETTINGS;
}

export async function saveAppSettings(settings: AppSettings): Promise<void> {
  await Preferences.set({ key: STORAGE_KEY, value: JSON.stringify(settings) });
}
