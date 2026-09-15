// Layout inicial do painel (PROJECT.md §10.4): duas páginas fixas usadas na
// primeira execução ou quando o layout salvo estiver corrompido. Os
// `actionId` aqui precisam existir no catálogo autoritativo do Mac
// (configs/config.example.json traz o conjunto correspondente); um
// actionId sem correspondência no Mac simplesmente aparece indisponível no
// botão — o editor (Fase 4) permitirá trocar a ação sem editar código.

import type { DashboardButton, DashboardConfig, DashboardPage } from "./layout";

function button(
  id: string,
  actionId: string,
  position: number,
  requireLongPress = false,
): DashboardButton {
  return { id, actionId, position, requireLongPress };
}

function appsPage(): DashboardPage {
  return {
    id: "page-apps",
    name: "Aplicativos",
    columns: 4,
    rows: 2,
    buttons: [
      button("btn-app-chrome", "app.chrome", 0),
      button("btn-app-whatsapp", "app.whatsapp", 1),
      button("btn-app-finder", "app.finder", 2),
      button("btn-app-terminal", "app.terminal", 3),
      button("btn-app-spotify", "app.spotify", 4),
      button("btn-app-vscode", "app.vscode", 5),
      button("btn-app-screenshot", "app.screenshot", 6),
      button("btn-shortcut-work", "shortcut.work", 7),
    ],
  };
}

function mediaPage(): DashboardPage {
  return {
    id: "page-media",
    name: "Mídia e sistema",
    columns: 4,
    rows: 2,
    buttons: [
      button("btn-media-previous", "media.spotify.previous", 0),
      button("btn-media-playpause", "media.spotify.playpause", 1),
      button("btn-media-next", "media.spotify.next", 2),
      button("btn-system-mute", "system.mute", 3),
      button("btn-volume-down", "volume.down", 4),
      button("btn-volume-up", "volume.up", 5),
      button("btn-display-sleep", "system.display_sleep", 6, true),
      button("btn-screen-lock", "system.lock", 7, true),
    ],
  };
}

export function defaultDashboardConfig(): DashboardConfig {
  return {
    schemaVersion: 1,
    activeProfileId: "default",
    profiles: [
      {
        id: "default",
        name: "Padrão",
        pages: [appsPage(), mediaPage()],
      },
    ],
  };
}
