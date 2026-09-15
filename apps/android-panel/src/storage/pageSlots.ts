// Modelo de "slots" usado só pelo editor (Fase 4, PROJECT.md §10.2-C): uma
// página vira um array de tamanho fixo columns*rows, cada posição com o
// DashboardButton daquele lugar ou null (vazio). Facilita adicionar,
// remover, mover e trocar botões de posição sem lidar com o array esparso
// de `buttons` diretamente. `applySlots` volta para o formato persistido,
// recalculando `position` a partir do índice do slot.

import type { DashboardButton, DashboardPage } from "./layout";

export function pageSlots(page: DashboardPage): Array<DashboardButton | null> {
  const total = page.columns * page.rows;
  const slots: Array<DashboardButton | null> = new Array(total).fill(null);
  for (const button of page.buttons) {
    if (button.position >= 0 && button.position < total) {
      slots[button.position] = button;
    }
  }
  return slots;
}

export function applySlots(
  page: DashboardPage,
  slots: Array<DashboardButton | null>,
): DashboardPage {
  const buttons: DashboardButton[] = [];
  slots.forEach((button, index) => {
    if (!button) return;
    buttons.push({ ...button, position: index });
  });
  return { ...page, buttons };
}
