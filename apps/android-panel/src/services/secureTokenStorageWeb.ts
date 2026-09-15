import { WebPlugin } from "@capacitor/core";
import type { SecureTokenStoragePlugin } from "./secureTokenStorage";

/**
 * Implementação web do plugin, usada só em `pnpm dev`/testes fora do
 * Android. NÃO é segura: guarda o token em memória do processo, sem
 * Keystore e sem persistir entre recarregamentos — de propósito, para
 * nunca criar a falsa impressão de que um token ficou protegido fora do
 * dispositivo real. Em produção o app roda sempre dentro do WebView
 * Android, onde SecureTokenStoragePlugin.java responde no lugar desta
 * classe (ver android/app/src/main/java/.../SecureTokenStoragePlugin.java).
 */
export class SecureTokenStorageWeb extends WebPlugin implements SecureTokenStoragePlugin {
  private tokens = new Map<string, string>();

  async setToken(options: { deviceId: string; token: string }): Promise<void> {
    this.tokens.set(options.deviceId, options.token);
  }

  async getToken(options: { deviceId: string }): Promise<{ token: string | null }> {
    return { token: this.tokens.get(options.deviceId) ?? null };
  }

  async clearToken(options: { deviceId: string }): Promise<void> {
    this.tokens.delete(options.deviceId);
  }
}
