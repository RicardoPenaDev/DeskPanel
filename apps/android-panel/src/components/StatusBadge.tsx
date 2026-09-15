// Indicador de conexão exibido no topo do painel principal (PROJECT.md
// §10.2-B: "Indicador de conexão no topo").

import type { ConnectionStatus } from "../protocol/wsClient";

export interface StatusBadgeProps {
  status: ConnectionStatus;
  macName?: string | null;
}

const STATUS_LABELS: Record<ConnectionStatus, string> = {
  disconnected: "Desconectado",
  connecting: "Conectando…",
  authenticating: "Autenticando…",
  connected: "Online",
  reconnecting: "Reconectando…",
};

export default function StatusBadge({ status, macName }: StatusBadgeProps) {
  const label = status === "connected" && macName ? macName : STATUS_LABELS[status];

  return (
    <div className={`dp-status-badge dp-status-badge--${status}`} role="status" aria-live="polite">
      <span className="dp-status-dot" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}
