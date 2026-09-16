// Fallback web do DeviceControlPlugin — usado em `pnpm dev`/testes fora do
// WebView Android, onde não existe janela nativa para ajustar. Cada método
// é um no-op deliberado (não lança, não bloqueia o fluxo do app).

import { WebPlugin } from "@capacitor/core";
import type { DeviceControlPlugin } from "./deviceControl";

export class DeviceControlWeb extends WebPlugin implements DeviceControlPlugin {
  async setKeepAwake(_options: { enabled: boolean }): Promise<void> {
    // no-op fora do Android — nada para ajustar num navegador comum.
  }

  async setImmersiveMode(_options: { enabled: boolean }): Promise<void> {
    // no-op fora do Android.
  }

  async setDimBrightness(_options: { enabled: boolean }): Promise<void> {
    // no-op fora do Android.
  }
}
