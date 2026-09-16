# Status — DeskPanel

Última atualização: 2026-09-16

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume, spotify/music) via `osascript`/`open`/`shortcuts`; `FakeExecutor` para testes; CLI carregando `config.json` de verdade.
- **Fase 2 — API, WebSocket e segurança**: servidor HTTP+WebSocket real (`serve`), pareamento com código temporário, tokens com hash salvo em `devices.json`, revogação imediata, rate limit, filtro de IP privado, heartbeat com dedup por `requestId`, socket administrativo para `pair`/`devices`/`revoke`/`status`.
- **Fase 3 — Painel Android funcional**: cliente HTTP/WS real (reconexão 1s/2s/4s/8s/15s, heartbeat), storage não secreto de conexão/layout, token no Android Keystore via plugin Capacitor local (ADR-0004), tela de pareamento, painel 4×2 com duas páginas e os 5 estados de botão, orientação paisagem e tela sempre ativa.
- **Fase 4 — Editor do painel**: `EditorScreen` completo — criar/excluir páginas, adicionar/remover botão só do catálogo autoritativo do Mac, reorganizar por tocar-para-selecionar/trocar (ADR-0005), trocar nome/ícone/cor, trava de toque prolongado para ações perigosas, mover botão entre páginas, restaurar layout padrão. 17 ícones locais. Backup automático de layout.
- **Fase 5 — Instalação real (scripts e documentação)**: `scripts/install-macos.sh`/`uninstall-macos.sh`/`build-apk.sh` reais (build/instala/desinstala LaunchAgent, `--purge`, `--web-only`). Docs `INSTALL-MACOS.md`/`INSTALL-ANDROID.md`/README atualizados. ADR-0006. Validado neste ambiente: sintaxe, `--help`, guarda de plataforma, `build-apk.sh --web-only` de ponta a ponta.
- **Fase 6 — Polimento do MVP (parte software)**: rotação de logs no próprio agente (`internal/logging`, ~5 MiB × 3 arquivos), `doctor` completo (porta, LaunchAgent, conectividade local, versão do protocolo), vibração via `@capacitor/haptics` (liga/desliga em Configurações), tela de Configurações completa (`features/settings/SettingsScreen.tsx` — nome/host/porta, reconectar, refazer pareamento, vibração, manter tela ligada, modo imersivo, brilho reduzido, diagnóstico de conexão, versões), plugin nativo local `DeviceControlPlugin.java` para os três toggles de tela, foco visível global (acessibilidade), revisão de segurança completa em `docs/SECURITY.md` (todos os controles do §12 conferidos contra o código), versão `0.1.0` embutida no binário via `-ldflags`. ADR-0007. 127 testes Vitest + `tsc -b` + `eslint` + `prettier` + `pnpm build`; Go com `go vet`/`gofmt`/testes limpos.

## Não feito / próximo

- **Fase 6 — parte física** (requer autorização explícita do proprietário, `PROJECT.md` §19.3, e hardware real): "correções do teste físico" (não há teste físico ainda para corrigir), teste contínuo de duas horas, validação real de vibração/keep-awake/modo imersivo/brilho no Moto G60, tag local `v0.1.0` (autorização separada, §17).
- **Fase 5 — execução real**: instalar no Mac de verdade, carregar o LaunchAgent, compilar o APK com Gradle/Android SDK reais, parear com o Moto G60 físico.
- Critérios de aceite do §18 que dependem de hardware físico continuam pendentes (APK instalado, ações básicas no setup real, etc.) — ver checklist completo em `PROJECT.md` §18.
- O lado nativo Android (`SecureTokenStoragePlugin`, `DeviceControlPlugin`, `MainActivity`/`AndroidManifest`) continua não compilado neste ambiente (sem Android SDK/Gradle) — só valida na execução real.
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor` (editor já trata como ações cadastráveis/perigosas; execução real pendente desde a Fase 1).
- Reorganizar por toque (ADR-0005) ainda não validado com dedo de verdade no Moto G60.
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ADR-0002); `pnpm` só funciona via `corepack pnpm` aqui (não afeta o Mac real).
