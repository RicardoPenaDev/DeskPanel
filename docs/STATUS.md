# Status — DeskPanel

Última atualização: 2026-09-16

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume, spotify/music) via `osascript`/`open`/`shortcuts`; `FakeExecutor` para testes; CLI carregando `config.json` de verdade.
- **Fase 2 — API, WebSocket e segurança**: servidor HTTP+WebSocket real (`serve`), pareamento com código temporário, tokens com hash salvo em `devices.json`, revogação imediata, rate limit, filtro de IP privado, heartbeat com dedup por `requestId`, socket administrativo para `pair`/`devices`/`revoke`/`status`.
- **Fase 3 — Painel Android funcional**: cliente HTTP/WS real (reconexão 1s/2s/4s/8s/15s, heartbeat), storage não secreto de conexão/layout, token no Android Keystore via plugin Capacitor local (ADR-0004), tela de pareamento, painel 4×2 com duas páginas e os 5 estados de botão, orientação paisagem e tela sempre ativa.
- **Fase 4 — Editor do painel**: `EditorScreen` completo — criar/excluir páginas, adicionar/remover botão só do catálogo autoritativo do Mac, reorganizar por tocar-para-selecionar/trocar (ADR-0005), trocar nome/ícone/cor, trava de toque prolongado para ações perigosas, mover botão entre páginas, restaurar layout padrão. 17 ícones locais. Backup automático de layout. 94 testes Vitest + `tsc -b` + `eslint` + `prettier` + `pnpm build` passando.
- **Fase 5 — Instalação real (scripts e documentação)**: `scripts/install-macos.sh` (build ou `--binary`, diretórios, config.json só se ausente, LaunchAgent sem usuário fixo, `launchctl bootstrap`, `--force`/`--skip-launchagent`/`--no-load`), `scripts/uninstall-macos.sh` (preserva dados por padrão, `--purge`/`--yes` para apagar), `scripts/build-apk.sh` (`pnpm install/test/build` + `cap sync` + Gradle, `--release`/`--web-only`/`--skip-install`/`--skip-tests`). Docs `INSTALL-MACOS.md`/`INSTALL-ANDROID.md`/README atualizados (ADB e instalação manual do APK). Validado neste ambiente: sintaxe dos 3 scripts, `--help`, guarda "só roda no macOS" (aborta corretamente em Linux), e `build-apk.sh --web-only` executado de ponta a ponta com sucesso. ADR-0006.

## Não feito / próximo

- **Fase 5 — execução real** (requer autorização explícita do proprietário, `PROJECT.md` §19.3, e um Mac/Moto G60 de verdade): rodar `install-macos.sh` no Mac real, carregar o LaunchAgent de verdade, compilar o APK com Gradle/Android SDK reais, parear com o Moto G60 físico, validar reconexão automática e ações básicas no setup real, confirmar que o agente volta após logout/login.
- Tela de Configurações (§10.2-D: refazer pareamento, vibração, modo imersivo, diagnóstico) ainda não existe.
- O lado nativo Android (plugin de token, `MainActivity`/`AndroidManifest`, da Fase 3) continua não compilado neste ambiente (sem Android SDK/Gradle) — só valida na execução real da Fase 5.
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor` (editor já trata como ações cadastráveis/perigosas; execução real no Mac pendente desde a Fase 1).
- Reorganizar por toque (ADR-0005) ainda não validado com dedo de verdade no Moto G60.
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ADR-0002); `pnpm` só funciona via `corepack pnpm` aqui (não afeta o Mac real).
