# Status — DeskPanel

Última atualização: 2026-09-15

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume, spotify/music) via `osascript`/`open`/`shortcuts`; `FakeExecutor` para testes; CLI carregando `config.json` de verdade.
- **Fase 2 — API, WebSocket e segurança**: servidor HTTP+WebSocket real (`serve`), pareamento com código temporário, tokens com hash salvo em `devices.json`, revogação imediata, rate limit, filtro de IP privado, heartbeat com dedup por `requestId`, socket administrativo para `pair`/`devices`/`revoke`/`status`.
- **Fase 3 — Painel Android funcional**: cliente HTTP/WS real (reconexão 1s/2s/4s/8s/15s, heartbeat), storage não secreto de conexão/layout, token no Android Keystore via plugin Capacitor local (ADR-0004), tela de pareamento, painel 4×2 com duas páginas e os 5 estados de botão, orientação paisagem e tela sempre ativa.
- **Fase 4 — Editor do painel**: `EditorScreen` completo — criar/excluir páginas (mínimo 1 página sempre), adicionar/remover botão escolhendo ação só do catálogo autoritativo do Mac (`ActionPicker`, nunca uma ação inventada), reorganizar por tocar-para-selecionar/tocar-para-trocar (ver ADR-0005 sobre por que não é drag-and-drop literal), trocar nome/ícone/cor, marcar toque prolongado (travado como obrigatório quando a ação já é perigosa no agente — não pode ser destravado pelo editor), mover botão entre páginas (com erro claro se a página de destino estiver cheia), restaurar layout padrão com confirmação. Pacote de 17 ícones locais (`components/Icon.tsx`, sem dependência externa) usado tanto no editor quanto nos botões do painel. `storage/layoutStorage.ts` ganhou backup automático + validação mais estrita, com fallback padrão→backup→layout de fábrica sem nunca travar. Entrada no editor por botão "Editar" no topo do painel ou toque prolongado em um slot vazio da grade. 94 testes Vitest + `tsc -b` + `eslint` + `prettier` + `pnpm build` passando.

## Não feito / próximo

- **Fase 5 — Instalação real**: scripts de instalação/desinstalação do macOS, LaunchAgent, build de APK real, teste físico no Moto G60.
- Tela de Configurações (§10.2-D: refazer pareamento, vibração, modo imersivo, diagnóstico) ainda não existe — não fazia parte das entregas das Fases 3/4.
- O lado nativo Android (`SecureTokenStoragePlugin.java`, `androidx.security:security-crypto`, `MainActivity`/`AndroidManifest`, da Fase 3) **continua não compilado nem testado** neste ambiente (sem Android SDK/Gradle) — só valida de verdade na Fase 5.
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor` (o editor já lida com `screen_lock`/`display_sleep` como ações cadastráveis e perigosas, mas a execução real no Mac continua pendente desde a Fase 1).
- Reorganizar por toque (ADR-0005) ainda não foi validado com um dedo de verdade no Moto G60 — só testado em Vitest/jsdom.
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ADR-0002); `pnpm` só funciona via `corepack pnpm` aqui (não afeta o Mac real).
