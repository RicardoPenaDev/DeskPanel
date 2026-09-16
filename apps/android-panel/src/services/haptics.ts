// Vibração curta em ação bem-sucedida/erro (PROJECT.md §10.1: "vibração
// curta opcional em ação bem-sucedida"; §10.2-D permite ligar/desligar).
// Usa o plugin oficial @capacitor/haptics — mapeado para o padrão de
// vibração do sistema no Android, sem plugin nativo próprio. Cada função
// engole erro silenciosamente: aparelho sem motor de vibração (ou rodando
// fora do Android) não pode derrubar a execução da ação.

import { Haptics, NotificationType } from "@capacitor/haptics";

export async function vibrateSuccess(): Promise<void> {
  try {
    await Haptics.notification({ type: NotificationType.Success });
  } catch {
    // sem suporte a haptics neste dispositivo/ambiente — ignora.
  }
}

export async function vibrateError(): Promise<void> {
  try {
    await Haptics.notification({ type: NotificationType.Error });
  } catch {
    // idem.
  }
}
