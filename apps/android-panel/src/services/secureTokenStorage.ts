// Interface TypeScript para o plugin Capacitor local que guarda o token de
// acesso no Android Keystore (PROJECT.md §10.3, §12.8). A implementação
// nativa fica em
// android/app/src/main/java/dev/ricardopena/deskpanel/SecureTokenStoragePlugin.java;
// o fallback web (secureTokenStorageWeb.ts) só existe para `pnpm dev` e
// testes, e não é seguro — ver aviso no próprio arquivo.
//
// O token NUNCA passa por storage/connectionConfig.ts, Capacitor
// Preferences, localStorage ou qualquer log.

import { registerPlugin } from "@capacitor/core";

export interface SecureTokenStoragePlugin {
  setToken(options: { deviceId: string; token: string }): Promise<void>;
  getToken(options: { deviceId: string }): Promise<{ token: string | null }>;
  clearToken(options: { deviceId: string }): Promise<void>;
}

const SecureTokenStorage = registerPlugin<SecureTokenStoragePlugin>("SecureTokenStorage", {
  web: () => import("./secureTokenStorageWeb").then((m) => new m.SecureTokenStorageWeb()),
});

export async function saveAccessToken(deviceId: string, token: string): Promise<void> {
  await SecureTokenStorage.setToken({ deviceId, token });
}

export async function readAccessToken(deviceId: string): Promise<string | null> {
  const { token } = await SecureTokenStorage.getToken({ deviceId });
  return token;
}

export async function clearAccessToken(deviceId: string): Promise<void> {
  await SecureTokenStorage.clearToken({ deviceId });
}

export default SecureTokenStorage;
