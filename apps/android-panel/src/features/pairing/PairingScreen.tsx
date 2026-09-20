// Tela de primeiro acesso (PROJECT.md §10.2-A): nome do dispositivo, IP ou
// hostname do Mac, porta e código de pareamento; ações "Testar conexão" e
// "Parear"; erros objetivos. Nenhum dado sensível fica em estado local
// além do que o usuário está digitando agora — o token só existe depois
// que pair() retorna e é salvo no Keystore pelo hook.

import { useState, type FormEvent } from "react";
import type { ConnectionConfig } from "../../storage/connectionConfig";
import type { HealthResponse, HttpResult } from "../../protocol/httpClient";
import type { PairInput, PairOutcome } from "../connection/useDeskPanelConnection";
import QrScanner from "./QrScanner";

export interface PairingScreenProps {
  initialConfig: ConnectionConfig;
  testConnection: (host: string, port: number) => Promise<HttpResult<HealthResponse>>;
  pair: (input: PairInput) => Promise<PairOutcome>;
  connectionError?: string | null;
}

export default function PairingScreen({
  initialConfig,
  testConnection,
  pair,
  connectionError,
}: PairingScreenProps) {
  const [deviceName, setDeviceName] = useState(initialConfig.deviceName);
  const [host, setHost] = useState(initialConfig.host);
  const [port, setPort] = useState(String(initialConfig.port || 38121));
  const [code, setCode] = useState("");
  const [testResult, setTestResult] = useState<string | null>(null);
  const [testing, setTesting] = useState(false);
  const [pairing, setPairing] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [scanning, setScanning] = useState(false);

  function handleQrCode(value: string): void {
    try {
      const parsed = new URL(value);
      if (parsed.protocol !== "deskpanel:" || parsed.hostname !== "pair") throw new Error("scheme");
      const nextHost = parsed.searchParams.get("host") ?? "";
      const nextPort = parsed.searchParams.get("port") ?? "38121";
      const nextCode = parsed.searchParams.get("code") ?? "";
      if (!nextHost || !/^\d{1,5}$/.test(nextPort) || !/^\d{6}$/.test(nextCode)) throw new Error("payload");
      setHost(nextHost);
      setPort(nextPort);
      setCode(nextCode);
      setScanning(false);
      setErrorMessage("Dados preenchidos pelo QR Code. Toque em Parear.");
    } catch {
      setErrorMessage("QR Code inválido. Use um código gerado pelo DeskPanel no Mac.");
    }
  }

  const portNumber = Number(port);
  const hostValid = host.trim().length > 0;
  const portValid = Number.isInteger(portNumber) && portNumber > 0 && portNumber <= 65535;
  const codeValid = /^\d{6}$/.test(code);
  const canSubmit = hostValid && portValid && codeValid && !pairing;

  async function handleTestConnection(): Promise<void> {
    if (!hostValid || !portValid) return;
    setTesting(true);
    setTestResult(null);
    const result = await testConnection(host.trim(), portNumber);
    setTesting(false);
    setTestResult(
      result.ok
        ? `Conectado: ${result.data.service} (protocolo v${result.data.protocolVersion})`
        : result.error.message,
    );
  }

  async function handlePair(event: FormEvent): Promise<void> {
    event.preventDefault();
    if (!canSubmit) return;
    setPairing(true);
    setErrorMessage(null);
    const outcome = await pair({
      deviceName: deviceName.trim(),
      host: host.trim(),
      port: portNumber,
      code,
    });
    setPairing(false);
    if (!outcome.ok) {
      setErrorMessage(outcome.message);
    }
  }

  if (scanning) return <QrScanner onCode={handleQrCode} onClose={() => setScanning(false)} />;

  return (
    <main className="dp-pairing-screen">
      <h1>DeskPanel</h1>
      <p className="dp-muted">Conecte este dispositivo ao Mac pela rede local.</p>

      <form onSubmit={handlePair} className="dp-pairing-form">
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
          <input
            value={host}
            onChange={(e) => setHost(e.target.value)}
            placeholder="192.168.1.50"
            autoComplete="off"
          />
        </label>

        <label className="dp-field">
          Porta
          <input value={port} onChange={(e) => setPort(e.target.value)} inputMode="numeric" />
        </label>

        <label className="dp-field">
          Código de pareamento
          <input
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
            inputMode="numeric"
            placeholder="000000"
          />
        </label>

        <div className="dp-pairing-actions">
          <button type="button" onClick={() => setScanning(true)} disabled={pairing}>
            Escanear QR Code
          </button>
          <button
            type="button"
            onClick={handleTestConnection}
            disabled={testing || !hostValid || !portValid}
          >
            {testing ? "Testando…" : "Testar conexão"}
          </button>
          <button type="submit" disabled={!canSubmit}>
            {pairing ? "Pareando…" : "Parear"}
          </button>
        </div>

        {testResult && <p className="dp-pairing-feedback">{testResult}</p>}
        {(errorMessage ?? connectionError) && (
          <p className="dp-pairing-error" role="alert">
            {errorMessage ?? connectionError}
          </p>
        )}
      </form>
    </main>
  );
}
