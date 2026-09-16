// Casca do painel: escolhe entre a tela de pareamento (PROJECT.md
// §10.2-A), o painel principal (§10.2-B), o editor (§10.2-C, Fase 4) e as
// Configurações (§10.2-D, Fase 6) conforme o estado de conexão vindo de
// useDeskPanelConnection e um modo de UI local (dashboard ↔ editor ↔
// settings). A instalação real fica nos scripts da Fase 5 — este
// componente não antecipa nenhuma delas.

import { useState } from "react";
import {
  useDeskPanelConnection,
  type DeskPanelConnectionDeps,
} from "../features/connection/useDeskPanelConnection";
import PairingScreen from "../features/pairing/PairingScreen";
import DashboardScreen from "../features/dashboard/DashboardScreen";
import EditorScreen from "../features/editor/EditorScreen";
import SettingsScreen from "../features/settings/SettingsScreen";

type PanelMode = "dashboard" | "editor" | "settings";

export interface AppProps {
  // Só para testes de componente injetarem deps falsas (mesmo padrão de
  // useDeskPanelConnection) sem precisar mockar módulos inteiros.
  connectionOverrides?: Partial<DeskPanelConnectionDeps>;
}

export default function App({ connectionOverrides }: AppProps = {}) {
  const connection = useDeskPanelConnection(connectionOverrides);
  const [mode, setMode] = useState<PanelMode>("dashboard");

  if (connection.phase === "loading" || !connection.connectionConfig || !connection.layout) {
    return (
      <main className="dp-splash">
        <h1>DeskPanel</h1>
        <p className="dp-muted">Carregando…</p>
      </main>
    );
  }

  if (connection.phase === "pairing") {
    return (
      <PairingScreen
        initialConfig={connection.connectionConfig}
        testConnection={connection.testConnection}
        pair={connection.pair}
        connectionError={connection.connectionError}
      />
    );
  }

  if (mode === "editor") {
    return (
      <EditorScreen
        layout={connection.layout}
        actionsCatalog={connection.actionsCatalog}
        onSave={connection.updateLayout}
        onRestoreDefault={connection.resetLayoutToDefault}
        onClose={() => setMode("dashboard")}
      />
    );
  }

  if (mode === "settings") {
    return (
      <SettingsScreen
        connectionConfig={connection.connectionConfig}
        appSettings={connection.appSettings}
        status={connection.status}
        macName={connection.macName}
        agentVersion={connection.agentVersion}
        connectionError={connection.connectionError}
        testConnection={connection.testConnection}
        updateConnectionConfig={connection.updateConnectionConfig}
        updateAppSettings={connection.updateAppSettings}
        reconnect={connection.reconnect}
        forgetPairing={connection.forgetPairing}
        onClose={() => setMode("dashboard")}
      />
    );
  }

  return (
    <DashboardScreen
      layout={connection.layout}
      actionsCatalog={connection.actionsCatalog}
      status={connection.status}
      macName={connection.macName}
      vibrationEnabled={connection.appSettings.vibrationEnabled}
      weather={connection.weather}
      executeAction={connection.executeAction}
      onOpenEditor={() => setMode("editor")}
      onOpenSettings={() => setMode("settings")}
    />
  );
}
