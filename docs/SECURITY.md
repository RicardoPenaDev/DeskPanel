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
