# Protocolo DeskPanel v1

Fonte de verdade normativa: `PROJECT.md` §8 e §9. Este documento é uma referência de consulta rápida durante a implementação; em caso de divergência, `PROJECT.md` prevalece.

## Base

```text
http://<ip-do-mac>:38121/api/v1
ws://<ip-do-mac>:38121/api/v1/ws
```

## HTTP

| Método | Endpoint | Auth | Finalidade |
|---|---|---|---|
| GET | `/health` | não | liveness check, resposta mínima |
| POST | `/pair` | código temporário | pareia um novo dispositivo |
| GET | `/actions` | Bearer token | lista ações autorizadas |
| GET | `/state` | Bearer token | estado inicial (Mac + mídia) |
| GET | `/version` | não | versões do agente e do protocolo |
| GET | `/ws` | primeira mensagem WS | upgrade para o canal em tempo real |

Sem endpoint HTTP genérico de execução — execução normal é só via WebSocket.

### `/health`

```json
{ "status": "ok", "service": "deskpanel-agent", "protocolVersion": 1 }
```

Não deve vazar hostname, usuário, caminhos, IPs internos, tokens ou lista de ações.

### Pareamento

1. `deskpanel-agent pair` no Mac abre uma janela de 5 minutos e imprime um código de 6 dígitos.
2. Android envia `POST /api/v1/pair`:

```json
{
  "code": "483921",
  "deviceId": "uuid-gerado-no-android",
  "deviceName": "Moto G60 do Ricardo",
  "appVersion": "0.1.0",
  "protocolVersion": 1
}
```

3. Resposta:

```json
{
  "deviceId": "uuid-gerado-no-android",
  "accessToken": "token-aleatorio-de-alta-entropia",
  "protocolVersion": 1
}
```

Regras: código numérico criptograficamente gerado, expira em 5 min, uso único, máx. 5 tentativas por código, token com >= 32 bytes aleatórios, Mac guarda só o hash do token, uma entrada por dispositivo em `devices.json`, nunca logar código completo ou token.

## WebSocket

Após abrir o socket, o cliente tem 5s para enviar `auth.hello`; nenhuma outra mensagem é aceita antes da autenticação.

### Envelope padrão

```json
{
  "type": "action.execute",
  "protocolVersion": 1,
  "requestId": "uuid",
  "timestamp": "2026-09-15T18:30:00Z",
  "payload": {}
}
```

### Tipos de mensagem

- `auth.hello` / `auth.accepted` — autenticação inicial.
- `action.execute` / `action.started` / `action.result` — ciclo de execução de uma ação.
- `ping` / `pong` — heartbeat (servidor envia `ping` a cada 20s; cliente responde `pong` em até 10s; 2 falhas seguidas encerram o socket).
- `state.snapshot` / `state.changed` — estado do Mac e da mídia.
- `error` — qualquer código da tabela abaixo.

Reconexão do Android: backoff com jitter — 1s, 2s, 4s, 8s, máx. 15s. Ao reconectar, pedir novo `state.snapshot`.

### Códigos de erro

```text
AUTH_REQUIRED        AUTH_INVALID          AUTH_REVOKED
PAIRING_CLOSED        PAIRING_CODE_INVALID  PAIRING_CODE_EXPIRED
PROTOCOL_UNSUPPORTED  INVALID_MESSAGE       ACTION_NOT_FOUND
ACTION_NOT_ALLOWED    ACTION_TIMEOUT        ACTION_FAILED
RATE_LIMITED          INTERNAL_ERROR
```

Mensagens exibidas ao usuário devem ser amigáveis; detalhes técnicos ficam só no log local.

## Limites de mensagem

- WebSocket: 16 KiB por mensagem.
- HTTP: 64 KiB por requisição.

## Versionamento

Mudanças incompatíveis exigem incrementar `protocolVersion`. O agente deve rejeitar versões de protocolo que não suporta com `PROTOCOL_UNSUPPORTED`.
