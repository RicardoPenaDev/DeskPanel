// Modelo de dados do layout local do painel (PROJECT.md §10.3). Persistido
// via Capacitor Preferences — nunca o token de acesso, que fica no Android
// Keystore por um plugin separado (ver docs/SECURITY.md).

export interface DashboardConfig {
  schemaVersion: 1;
  activeProfileId: string;
  profiles: DashboardProfile[];
}

export interface DashboardProfile {
  id: string;
  name: string;
  pages: DashboardPage[];
}

export interface DashboardPage {
  id: string;
  name: string;
  columns: 4;
  rows: 2;
  buttons: DashboardButton[];
}

export type ButtonColor = "neutral" | "blue" | "green" | "orange" | "red" | "purple";

export interface DashboardButton {
  id: string;
  actionId: string;
  position: number;
  labelOverride?: string;
  iconOverride?: string;
  color?: ButtonColor;
  requireLongPress: boolean;
}
