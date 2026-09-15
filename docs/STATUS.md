# Status — DeskPanel

Última atualização: 2026-09-15

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, Makefile, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume_delta/set, mute_toggle, spotify/music_control) via `osascript`/`open`/`shortcuts`, sem shell intermediário; `FakeExecutor` mantido para testes; CLI `serve`/`status`/`doctor` agora carregam e validam `config.json` de verdade; `doctor` checa binários do macOS e permissão do diretório de dados. 36 testes Go + 1 teste React, todos passando via runner injetável (nenhum comando real é executado nos testes).

## Não feito / próximo

- **Fase 2 — API, WebSocket e segurança**: endpoints HTTP, WebSocket, pareamento real, tokens/revogação, rate limit, heartbeat — é o que falta pra `serve`, `pair`, `devices`, `revoke`, `status` funcionarem de verdade (hoje `serve` só carrega config e para).
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor`.
- Projeto Android nativo (`apps/android-panel/android`) nunca foi compilado com Gradle real (sem SDK neste ambiente).
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ver `docs/DECISIONS.md` ADR-0002); no Mac real, usar Homebrew normalmente.
