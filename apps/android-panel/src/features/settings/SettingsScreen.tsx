// Tela de Configurações (PROJECT.md §10.2-D): nome do dispositivo,
// IP/porta do Mac, reconectar, refazer pareamento, vibração, manter tela
// ligada, modo imersivo, brilho reduzido, diagnóstico de conexão e
// versões do app/agente/protocolo. Formulário de conexão usa estado local
// com botão "Salvar" explícito (mesmo padrão de PairingScreen) — os
// toggles aplicam na hora, tanto na preferência persistida quanto no
// controle nativo (services/deviceControl.ts).

import { useState } from "react";
import type { ConnectionConfig } from "../../storage/connectionConfig";
import type { AppSettings } from "../../storage/appSettings";
import type { HealthResponse, HttpResult } from "../../protocol/httpClient";
import type { ConnectionStatus } from "../../protocol/wsClient";
import { PROTOCOL_VERSION } from "../../protocol/messages";
import { APP_VERSION } from "../../version";
import { setDimBrightness, setImmersiveMode, setKeepAwake } from "../../services/deviceControl";

export interface SettingsScreenProps {
  connectionConfig: ConnectionConfig;
  appSettings: AppSettings;
  status: ConnectionStatus;
  macName: string | null;
  agentVersion: string | null;
  connectionError?: string | null;
  testConnection: (host: string, port: number) => Promise<HttpResult<HealthResponse>>;
  updateConnectionConfig: (next: ConnectionConfig) => Promise<void>;
  updateAppSettings: (next: AppSettings) => Promise<void>;
  reconnect: () => void;
  forgetPairing: () => Promise<void>;
  onClose: () => void;
}

const STATUS_LABEL: Record<ConnectionStatus, string> = {
  disconnected: "Desconectado",
  connecting: "Conectando…",
  authenticating: "Autenticando…",
  connected: "Online",
  reconnecting: "Reconectando…",
};

export default function SettingsScreen({
  connectionConfig,
  appSettings,
  status,
  macName,
  agentVersion,
  connectionError,
  testConnection,
  updateConnectionConfig,
  updateAppSettings,
  reconnect,
  forgetPairing,
  onClose,
}: SettingsScreenProps) {
  const [deviceName, setDeviceName] = useState(connectionConfig.deviceName);
  const [host, setHost] = useState(connectionConfig.host);
  const [port, setPort] = useState(String(connectionConfig.port));
  const [savingConnection, setSavingConnection] = useState(false);
  const [savedFeedback, setSavedFeedback] = useState<string | null>(null);

  const [diagnosing, setDiagnosing] = useState(false);
  const [diagnosis, setDiagnosis] = useState<string | null>(null);

  const [confirmingForget, setConfirmingForget] = useState(false);
  const [forgetting, setForgetting] = useState(false);

  const portNumber = Number(port);
  const hostValid = host.trim().length > 0;
  const portValid = Number.isInteger(portNumber) && portNumber > 0 && portNumber <= 65535;
  const canSaveConnection = hostValid && portValid && !savingConnection;

  async function handleSaveConnection(): Promise<void> {
    if (!canSaveConnection) return;
    setSavingConnection(true);
    setSavedFeedback(null);
    await updateConnectionConfig({
      ...connectionConfig,
      deviceName: deviceName.trim() || connectionConfig.deviceName,
      host: host.trim(),
      port: portNumber,
    });
    setSavingConnection(false);
    setSavedFeedback("Salvo. Reconectando com os novos dados…");
  }

  async function handleDiagnose(): Promise<void> {
    setDiagnosing(true);
    setDiagnosis(null);
    const startedAt = performance.now();
    const result = await testConnection(connectionConfig.host, connectionConfig.port);
    const latencyMs = Math.round(performance.now() - startedAt);
    setDiagnosing(false);
    setDiagnosis(
      result.ok
        ? `Mac respondeu em ${latencyMs} ms — ${result.data.service}, protocolo v${result.data.protocolVersion}`
        : `Falha (${latencyMs} ms): ${result.error.message}`,
    );
  }

  async function handleToggle(patch: Partial<AppSettings>): Promise<void> {
    const next: AppSettings = { ...appSettings, ...patch };
    await updateAppSettings(next);
    if ("keepAwakeEnabled" in patch) await setKeepAwake(next.keepAwakeEnabled);
    if ("immersiveModeEnabled" in patch) await setImmersiveMode(next.immersiveModeEnabled);
    if ("dimBrightnessEnabled" in patch) await setDimBrightness(next.dimBrightnessEnabled);
  }

  async function handleForgetPairing(): Promise<void> {
    setForgetting(true);
    await forgetPairing();
    setForgetting(false);
  }

  return (
    <main className="dp-settings-screen">
      <header className="dp-settings-header">
        <h1>Configurações</h1>
        <button type="button" className="dp-settings-back" onClick={onClose}>
          Voltar ao painel
        </button>
      </header>

      <section className="dp-settings-section">
        <h2>Conexão</h2>
        <label className="dp-field">
          Nome do dispositivo
          <input
            value={deviceName}
            onChange={(e) => setDeviceName(e.target.value)}
            autoComplete="off"
          />
        </label>
        <label className="dp-field">
          IP ou hostname do Mac
          <input value={host} onChange={(e) => setHost(e.target.value)} autoComplete="off" />
        </label>
        <label className="dp-field">
          Porta
          <input value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" />
        </label>

        <div className="dp-settings-actions">
          <button type="button" onClick={handleSaveConnection} disabled={!canSaveConnection}>
            {savingConnection ? "Salvando…" : "Salvar"}
          </button>
          <button type="button" onClick={reconnect}>
            Reconectar
          </button>
        </div>
        {savedFeedback && <p className="dp-settings-feedback">{savedFeedback}</p>}

        <p className="dp-settings-status">
          Status: <strong>{STATUS_LABEL[status]}</strong>
          {macName && ` — ${macName}`}
        </p>
        {connectionError && (
          <p className="dp-settings-error" role="alert">
            {connectionError}
          </p>
        )}
      </section>

      <section className="dp-settings-section">
        <h2>Pareamento</h2>
        {!confirmingForget ? (
          <button type="button" onClick={() => setConfirmingForget(true)}>
            Refazer pareamento
          </button>
        ) : (
          <div className="dp-settings-confirm">
            <p>Isso apaga o token salvo neste aparelho. Você vai precisar de um novo código.</p>
            <button type="button" onClick={handleForgetPairing} disabled={forgetting}>
              {forgetting ? "Removendo…" : "Confirmar"}
            </button>
            <button type="button" onClick={() => setConfirmingForget(false)} disabled={forgetting}>
              Cancelar
            </button>
          </div>
        )}
      </section>

      <section className="dp-settings-section">
        <h2>Comportamento</h2>
        <label className="dp-settings-toggle">
          <input
            type="checkbox"
            checked={appSettings.vibrationEnabled}
            onChange={(e) => void handleToggle({ vibrationEnabled: e.target.checked })}
          />
          Vibração em ação bem-sucedida ou com erro
        </label>
        <label className="dp-settings-toggle">
          <input
            type="checkbox"
            checked={appSettings.keepAwakeEnabled}
            onChange={(e) => void handleToggle({ keepAwakeEnabled: e.target.checked })}
          />
          Manter tela ligada
        </label>
        <label className="dp-settings-toggle">
          <input
            type="checkbox"
            checked={appSettings.immersiveModeEnabled}
            onChange={(e) => void handleToggle({ immersiveModeEnabled: e.target.checked })}
          />
          Modo imersivo (esconder barras do sistema)
        </label>
        <label className="dp-settings-toggle">
          <input
            type="checkbox"
            checked={appSettings.dimBrightnessEnabled}
            onChange={(e) => void handleToggle({ dimBrightnessEnabled: e.target.checked })}
          />
          Brilho reduzido enquanto o painel estiver ativo
        </label>
      </section>

      <section className="dp-settings-section">
        <h2>Diagnóstico de conexão</h2>
        <button type="button" onClick={handleDiagnose} disabled={diagnosing}>
          {diagnosing ? "Testando…" : "Testar conexão agora"}
        </button>
        {diagnosis && <p className="dp-settings-feedback">{diagnosis}</p>}
      </section>

      <section className="dp-settings-section">
        <h2>Versões</h2>
        <p>App: {APP_VERSION}</p>
        <p>Agente: {agentVersion ?? "desconhecida (sem conexão ainda)"}</p>
        <p>Protocolo: v{PROTOCOL_VERSION}</p>
      </section>
    </main>
  );
}
