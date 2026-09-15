# Status — DeskPanel

Última atualização: 2026-09-15

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume, spotify/music) via `osascript`/`open`/`shortcuts`; `FakeExecutor` para testes; CLI carregando `config.json` de verdade.
- **Fase 2 — API, WebSocket e segurança**: servidor HTTP+WebSocket real (`serve`), pareamento com código temporário, tokens com hash salvo em `devices.json`, revogação imediata, rate limit, filtro de IP privado, heartbeat com dedup por `requestId`, socket administrativo para `pair`/`devices`/`revoke`/`status`. Testado com `make verify` + smoke test manual ponta a ponta contra o binário real (pareou, autenticou, executou, revogou — tudo funcionou).

## Não feito / próximo

- **Fase 3 — Painel Android funcional**: telas reais (pareamento, painel 4×2, estados visuais), cliente HTTP/WS de verdade no app, reconexão, storage seguro do token no Keystore.
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor`.
- Projeto Android nativo nunca foi compilado com Gradle real (sem SDK neste ambiente).
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ver `docs/DECISIONS.md` ADR-0002); no Mac real, usar Homebrew normalmente.
