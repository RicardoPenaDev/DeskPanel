// Casca do painel: escolhe entre a tela de pareamento (PROJECT.md
// §10.2-A) e o painel principal (§10.2-B) conforme o estado de conexão
// vindo de useDeskPanelConnection. As Fases 4/5/6 acrescentam editor,
// instalação real e polimento — este componente não antecipa nenhuma
// delas.

import { useDeskPanelConnection } from "../features/connection/useDeskPanelConnection";
import PairingScreen from "../features/pairing/PairingScreen";
import DashboardScreen from "../features/dashboard/DashboardScreen";

export default function App() {
  const connection = useDeskPanelConnection();

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

  return (
    <DashboardScreen
      layout={connection.layout}
      actionsCatalog={connection.actionsCatalog}
      status={connection.status}
      macName={connection.macName}
      executeAction={connection.executeAction}
    />
  );
}
