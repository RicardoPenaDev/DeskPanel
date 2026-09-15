// Configuração de conexão não secreta (IP/hostname do Mac, porta, nome e
// id do dispositivo) — persistida com Capacitor Preferences, conforme
// PROJECT.md §10.3. O token de acesso NUNCA passa por aqui: fica só no
// Android Keystore via services/secureTokenStorage.ts.

import { Preferences } from "@capacitor/preferences";

export interface ConnectionConfig {
  deviceId: string;
  deviceName: string;
  host: string;
  port: number;
}

const STORAGE_KEY = "deskpanel.connectionConfig";
export const DEFAULT_DEVICE_NAME = "Moto G60 do Ricardo";
export const DEFAULT_PORT = 38121;

function generateDeviceId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `device-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function isValidConfig(value: unknown): value is ConnectionConfig {
  if (!value || typeof value !== "object") return false;
  const v = value as Record<string, unknown>;
  return (
    typeof v.deviceId === "string" &&
    typeof v.deviceName === "string" &&
    typeof v.host === "string" &&
    typeof v.port === "number"
  );
}

/**
 * loadOrCreateConnectionConfig lê a configuração salva; se não existir ou
 * estiver corrompida, cria uma nova (com um novo deviceId estável) e já a
 * persiste, para que o mesmo deviceId sobreviva a reinícios do app mesmo
 * antes do primeiro pareamento.
 */
export async function loadOrCreateConnectionConfig(): Promise<ConnectionConfig> {
  const { value } = await Preferences.get({ key: STORAGE_KEY });
  if (value) {
    try {
      const parsed = JSON.parse(value);
      if (isValidConfig(parsed)) return parsed;
    } catch {
      // cai para criar uma configuração nova abaixo
    }
  }

  const fresh: ConnectionConfig = {
    deviceId: generateDeviceId(),
    deviceName: DEFAULT_DEVICE_NAME,
    host: "",
    port: DEFAULT_PORT,
  };
  await Preferences.set({ key: STORAGE_KEY, value: JSON.stringify(fresh) });
  return fresh;
}

export async function saveConnectionConfig(config: ConnectionConfig): Promise<void> {
  await Preferences.set({ key: STORAGE_KEY, value: JSON.stringify(config) });
}

export async function clearConnectionConfig(): Promise<void> {
  await Preferences.remove({ key: STORAGE_KEY });
}
