// Wrapper TypeScript do plugin nativo DeviceControlPlugin.java (keep
// awake, modo imersivo, brilho reduzido — PROJECT.md §10.2-D). Segue o
// mesmo padrão do secureTokenStorage.ts: interface do plugin + fallback
// web para pnpm dev/testes, onde essas chamadas não fazem sentido (não há
// janela nativa fora do WebView Android).

import { registerPlugin } from "@capacitor/core";

export interface DeviceControlPlugin {
  setKeepAwake(options: { enabled: boolean }): Promise<void>;
  setImmersiveMode(options: { enabled: boolean }): Promise<void>;
  setDimBrightness(options: { enabled: boolean }): Promise<void>;
}

const DeviceControl = registerPlugin<DeviceControlPlugin>("DeviceControl", {
  web: () => import("./deviceControlWeb").then((m) => new m.DeviceControlWeb()),
});

export async function setKeepAwake(enabled: boolean): Promise<void> {
  await DeviceControl.setKeepAwake({ enabled });
}

export async function setImmersiveMode(enabled: boolean): Promise<void> {
  await DeviceControl.setImmersiveMode({ enabled });
}

export async function setDimBrightness(enabled: boolean): Promise<void> {
  await DeviceControl.setDimBrightness({ enabled });
}
