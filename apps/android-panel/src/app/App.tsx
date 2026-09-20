// Casca do painel: escolhe entre a tela de pareamento (PROJECT.md
// §10.2-A), o painel principal (§10.2-B), o editor (§10.2-C, Fase 4) e as
// Configurações (§10.2-D, Fase 6) conforme o estado de conexão vindo de
// useDeskPanelConnection e um modo de UI local (dashboard ↔ editor ↔
// settings). A instalação real fica nos scripts da Fase 5 — este
// componente não antecipa nenhuma delas.

import { useEffect, useRef, useState } from "react";
import {
  useDeskPanelConnection,
  type DeskPanelConnectionDeps,
} from "../features/connection/useDeskPanelConnection";
import PairingScreen from "../features/pairing/PairingScreen";
import DashboardScreen from "../features/dashboard/DashboardScreen";
import EditorScreen from "../features/editor/EditorScreen";
import SettingsScreen from "../features/settings/SettingsScreen";
import { setDimBrightness, setImmersiveMode, setKeepAwake } from "../services/deviceControl";

type PanelMode = "dashboard" | "editor" | "settings";

type ConnectionNotice = { kind: "offline" | "online"; message: string };

export interface AppProps {
  // Só para testes de componente injetarem deps falsas (mesmo padrão de
  // useDeskPanelConnection) sem precisar mockar módulos inteiros.
  connectionOverrides?: Partial<DeskPanelConnectionDeps>;
}

export default function App({ connectionOverrides }: AppProps = {}) {
  const connection = useDeskPanelConnection(connectionOverrides);
  const [mode, setMode] = useState<PanelMode>("dashboard");
  const previousStatus = useRef(connection.status);
  const [connectionNotice, setConnectionNotice] = useState<ConnectionNotice | null>(null);

  useEffect(() => {
    const previous = previousStatus.current;
    const next = connection.status;
    previousStatus.current = next;

    if (previous === next) return;
    if (next === "connected" && previous !== "connected") {
      setConnectionNotice({ kind: "online", message: "Conectado ao Mac" });
    } else if (
      previous === "connected" &&
      (next === "disconnected" || next === "reconnecting" || next === "connecting")
    ) {
      setConnectionNotice({ kind: "offline", message: "Sem conexão com o Mac" });
    }
  }, [connection.status]);

  useEffect(() => {
    if (!connectionNotice) return;
    const timer = window.setTimeout(() => setConnectionNotice(null), 3500);
    return () => window.clearTimeout(timer);
  }, [connectionNotice]);

  // Aplica keep-awake/imersivo/brilho reduzido assim que as preferências
  // salvas terminam de carregar — sem isso, o modo imersivo (e os outros
  // dois) só entravam em vigor se o usuário mexesse no toggle das
  // Configurações NESTA sessão; reabrir o app sempre voltava pro padrão
  // "sem imersivo" do Android, mostrando a barra de status por cima do
  // recorte da câmera até alguém abrir Configurações de novo.
  useEffect(() => {
    if (connection.phase === "loading") return;
    void setKeepAwake(connection.appSettings.keepAwakeEnabled);
    void setImmersiveMode(connection.appSettings.immersiveModeEnabled);
    void setDimBrightness(connection.appSettings.dimBrightnessEnabled);
  }, [connection.phase, connection.appSettings]);

  if (connection.phase === "loading" || !connection.connectionConfig || !connection.layout) {
    return (
      <main className="dp-splash">
        <h1>DeskPanel</h1>
        <p className="dp-muted">Carregando…</p>
      </main>
    );
  }

  const notice = connectionNotice ? (
    <div
      className={`dp-connection-notice dp-connection-notice--${connectionNotice.kind}`}
      role="status"
      aria-live="polite"
    >
      <span className="dp-connection-notice__dot" aria-hidden="true" />
      {connectionNotice.message}
    </div>
  ) : null;

  if (connection.phase === "pairing") {
    return (
      <>
        {notice}
        <PairingScreen
          initialConfig={connection.connectionConfig}
          testConnection={connection.testConnection}
          pair={connection.pair}
          connectionError={connection.connectionError}
        />
      </>
    );
  }

  if (mode === "editor") {
    return (
      <>
        {notice}
        <EditorScreen
          layout={connection.layout}
          actionsCatalog={connection.actionsCatalog}
          onSave={connection.updateLayout}
          onRestoreDefault={connection.resetLayoutToDefault}
          onClose={() => setMode("dashboard")}
        />
      </>
    );
  }

  if (mode === "settings") {
    return (
      <>
        {notice}
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
      </>
    );
  }

  return (
    <>
      {notice}
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
    </>
  );
}
