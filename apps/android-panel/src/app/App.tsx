// Casca do painel: escolhe entre a tela de pareamento (PROJECT.md
// §10.2-A), o painel principal (§10.2-B) e o editor (§10.2-C, Fase 4)
// conforme o estado de conexão vindo de useDeskPanelConnection e um modo
// de UI local (dashboard ↔ editor). As Fases 5/6 acrescentam instalação
// real e polimento — este componente não antecipa nenhuma delas.

import { useState } from "react";
import { useDeskPanelConnection } from "../features/connection/useDeskPanelConnection";
import PairingScreen from "../features/pairing/PairingScreen";
import DashboardScreen from "../features/dashboard/DashboardScreen";
import EditorScreen from "../features/editor/EditorScreen";

type PanelMode = "dashboard" | "editor";

export default function App() {
  const connection = useDeskPanelConnection();
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

  return (
    <DashboardScreen
      layout={connection.layout}
      actionsCatalog={connection.actionsCatalog}
      status={connection.status}
      macName={connection.macName}
      executeAction={connection.executeAction}
      onOpenEditor={() => setMode("editor")}
    />
  );
}
