// Espelho em TypeScript do protocolo descrito em docs/PROTOCOL.md e
// implementado em Go em apps/mac-agent/internal/protocol. Mudar um lado
// sem o outro é um bug — mantenha os dois em sincronia manualmente até
// existir geração de código compartilhada.

export const PROTOCOL_VERSION = 1;

export const MAX_WEBSOCKET_MESSAGE_BYTES = 16 * 1024;
export const MAX_HTTP_REQUEST_BYTES = 64 * 1024;

export type MessageType =
  | "auth.hello"
  | "auth.accepted"
  | "action.execute"
  | "action.started"
  | "action.result"
  | "ping"
  | "pong"
  | "state.snapshot"
  | "state.changed"
  | "error";

export interface Envelope<TPayload = unknown> {
  type: MessageType;
  protocolVersion: number;
  requestId: string;
  timestamp: string;
  payload: TPayload;
}

export type ErrorCode =
  | "AUTH_REQUIRED"
  | "AUTH_INVALID"
  | "AUTH_REVOKED"
  | "PAIRING_CLOSED"
  | "PAIRING_CODE_INVALID"
  | "PAIRING_CODE_EXPIRED"
  | "PROTOCOL_UNSUPPORTED"
  | "INVALID_MESSAGE"
  | "ACTION_NOT_FOUND"
  | "ACTION_NOT_ALLOWED"
  | "ACTION_TIMEOUT"
  | "ACTION_FAILED"
  | "RATE_LIMITED"
  | "INTERNAL_ERROR"
  | "WEATHER_UNAVAILABLE";
