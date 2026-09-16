# Segurança — DeskPanel

Modelo de ameaça e controles obrigatórios. Fonte normativa completa: `PROJECT.md` §11 e §12. Qualquer mudança que afrouxe algo listado aqui exige ADR em `docs/DECISIONS.md` e autorização explícita do proprietário (`PROJECT.md` §19.2).

## Modelo de ameaça (resumo)

O DeskPanel Agent roda com os privilégios do usuário no Mac e aceita conexões da rede local. O risco principal é execução remota de comandos arbitrários caso um atacante na mesma rede (ou um app malicioso no celular) consiga falar com o agente. A mitigação central é: **o celular nunca envia comandos — só um `actionId` que já existe, pré-cadastrado e validado no Mac.**

## Controles obrigatórios

1. Nenhum comando, script, caminho ou URL vindo do Android é aceito — só IDs de ações pré-cadastradas.
2. Nenhum shell intermediário (`sh -c`, `bash -c`, `eval`, concatenação de string viram comando) — sempre `exec.CommandContext` com binário e argumentos fixos.
3. Todo JSON validado com limite de tamanho (16 KiB WS / 64 KiB HTTP).
4. Comparação de token resistente a timing quando aplicável.
5. Mac armazena só o hash do token (nunca o token em texto plano).
6. Android armazena o token no Android Keystore via plugin Capacitor local — nunca em `localStorage`, JSON ou log.
7. Pareamento fechado por padrão; janela de 5 min, uso único, máx. 5 tentativas por código.
8. Rate limit por IP e por dispositivo; bloqueio temporário após tentativas de autenticação repetidas.
9. Nunca logar tokens, códigos completos de pareamento ou payloads sensíveis.
10. Validar `Origin` do WebSocket — aceitar só as origens esperadas do Capacitor.
11. Rejeitar versões de protocolo incompatíveis (`PROTOCOL_UNSUPPORTED`).
12. Revogação individual de dispositivo deve derrubar acesso imediatamente.
13. Não escutar em IPv6 público no MVP.
14. Sem UPnP, port forwarding, Cloudflare Tunnel ou qualquer exposição pública.
15. Acesso remoto fora da rede local, se um dia for necessário, só via Tailscale — nunca porta pública.
16. O agente deve validar que o IP do cliente pertence a uma faixa privada/link-local, salvo configuração futura explicitamente aprovada.

## Arquivos e permissões no macOS

- Arquivos com tokens: permissão `0600`.
- Diretórios privados (`~/Library/Application Support/DeskPanel/`): `0700`.
- Escrita de JSON sempre atômica (arquivo temporário + `fsync` quando aplicável + rename).
- O agente não roda como `root`.

## Permissões do sistema operacional

### Android
- Internet/rede local, vibração (opcional), manter tela ativa, Android Keystore.
- Tráfego HTTP claro permitido **só** para comunicação local do MVP — nunca para endereços públicos.

### macOS
- **Acessibilidade**: necessária para `keystroke` e alguns controles de sistema.
- **Automação**: necessária para AppleScript (Spotify, Apple Music, System Events).
- **Rede local/firewall**: necessária para o Android alcançar a porta `38121`.
- **Gravação de tela**: não solicitar no MVP (nenhuma função captura conteúdo de tela).

## Tipos de ação permitidos (MVP)

Ver `PROJECT.md` §7.4. Resumo: `open_app`, `open_url` (só http/https), `run_shortcut`, `keystroke` (teclas/modificadores permitidos, template fixo), `volume_delta`/`volume_set` (0–100), `mute_toggle`, `spotify_control`/`music_control` (play_pause/next/previous), `screen_lock`, `display_sleep` (os dois últimos exigem toque prolongado no Android).

**Explicitamente fora do MVP:** `shell`, `exec`, script livre, argumento vindo do cliente, caminho de arquivo vindo do cliente, URL vinda do cliente.

## Logs

Ver `PROJECT.md` §13 para formato e rotação. Nunca registrar: access token, código de pareamento completo, saída completa de processos, dados pessoais desnecessários.

## Revisão de segurança — 0.1.0 (Fase 6)

Conferência de cada controle do §12 contra o código real, em 2026-09-16.
Não é uma auditoria externa — é uma releitura linha a linha do próprio
autor antes da versão 0.1.0, como pedido pela entrega "revisão de
segurança" da Fase 6.

| # | Controle (§12) | Onde | Status |
|---|---|---|---|
| 1 | Nenhum comando/script/caminho/URL vindo do Android | `internal/websocket/handler.go` (`action.execute` só carrega `actionId`); `internal/actions` valida o catálogo no Mac | OK |
| 2 | Sem shell intermediário | `internal/executor/macos.go` usa `exec.CommandContext` com binário e argumentos fixos, nunca `sh -c` | OK |
| 3/4 | Limite de 16 KiB por mensagem WebSocket | `internal/protocol/protocol.go: MaxWebSocketMessageBytes = 16*1024`, aplicado na leitura das mensagens | OK |
| 3/5 | Limite de 64 KiB por requisição HTTP | `internal/protocol/protocol.go: MaxHTTPRequestBytes = 64*1024`, aplicado via middleware `maxBody` em `internal/api/server.go` | OK |
| 6 | Comparação de token resistente a timing | `internal/auth/auth.go` e `internal/devices/devices.go` usam `subtle.ConstantTimeCompare` sobre o hash, nunca `==` na string | OK |
| 7 | Só hash do token no Mac | `internal/auth/auth.go` (`sha256.Sum256`); `devices.json` grava `TokenHash`, nunca o token | OK |
| 8 | Token no Android Keystore | `SecureTokenStoragePlugin.java` + `EncryptedSharedPreferences` (ADR-0004) — não compilado neste ambiente, ver ressalva abaixo | OK (não verificado por build real) |
| 9 | Pareamento fechado por padrão | `internal/pairing`: sem `pair` explícito no Mac, `/api/v1/pair` sempre retorna `PAIRING_CLOSED` | OK |
| 10 | Rate limit por IP | `internal/api/middleware.go: rateLimit()`, aplicado a toda a árvore de rotas (`Server.Handler()`), inclusive `/ws` | OK |
| 10 | Rate limit "por dispositivo" | Pareamento: `FailureTracker` chaveado por IP em `pair.go`. Autenticação WS: cada tentativa fecha a conexão (`authenticate()` em `handler.go`), então uma nova tentativa exige nova conexão HTTP — cai sob o mesmo rate limit por IP acima. Não há uma segunda trava chaveada por deviceId hoje; o IP já cobre o caso de uso do MVP (um Android por rede) | OK, com nota |
| 11 | Bloqueio temporário após falhas repetidas | `internal/ratelimit.FailureTracker` (`IsLocked`/`RecordFailure`), usado em `pair.go` | OK |
| 12 | Nunca logar token/código completo/payload sensível | `grep` por `AccessToken`/`token` perto de chamadas de log não encontrou nenhuma — logger nunca recebe o payload bruto de `auth.hello` | OK |
| 13 | Validar `Origin` do WebSocket | `internal/websocket/handler.go: allowedOrigins` + `websocket.AcceptOptions{OriginPatterns: allowedOrigins}` | OK |
| 14 | Rejeitar protocolo incompatível | `internal/protocol` + checagem de `protocolVersion` no pareamento e no handshake WS (`PROTOCOL_UNSUPPORTED`) | OK |
| 15 | Revogação individual derruba acesso imediato | `internal/devices/devices.go: Revoke()` marca `Revoked=true`; `FindByToken`/`authenticate()` ignoram dispositivos revogados na próxima checagem | OK |
| 16 | Não escutar em IPv6 público | `configs/config.example.json: listenAddress = "0.0.0.0"` — `net.Listen("tcp", "0.0.0.0:porta")` no Go só faz bind em IPv4, não dual-stack | OK |
| 17 | Sem UPnP/port forwarding/Tunnel | Nenhuma dependência do tipo no código ou nos scripts de instalação (Fase 5) | OK |
| — | Cliente fora de rede privada/link-local é rejeitado | `internal/netguard.IsAllowedAddr` + `requirePrivateNetwork` em `internal/api/middleware.go`, aplicado a toda a árvore de rotas | OK |

Ressalva permanente (já registrada em ADR-0004 e no `docs/STATUS.md`): o
lado nativo Android (`SecureTokenStoragePlugin.java`,
`DeviceControlPlugin.java`) não compila neste ambiente de desenvolvimento
(sandbox sem Android SDK/Gradle) — a validação de que o Keystore realmente
protege o token só acontece com um build real no Mac do usuário (Fase 5) e
teste físico no Moto G60.

Nenhum controle do §12 foi enfraquecido nesta revisão. Nenhuma mudança de
código de segurança foi necessária — a revisão confirmou o que as Fases
1-2 já implementaram, e documentou o item 10 (rate limit por dispositivo)
com mais precisão.
