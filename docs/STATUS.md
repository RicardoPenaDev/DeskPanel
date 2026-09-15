# Status — DeskPanel

Última atualização: 2026-09-15

## Fase atual: Fase 0 — Fundação — concluída

Repositório inicializado, documentação base escrita, esqueletos Go e React/Capacitor compilando, `make verify` passando limpo. Nenhuma funcionalidade real foi implementada (nenhuma ação executa de fato, sem servidor real, sem pareamento real) — isso é esperado para a Fase 0.

## O que foi implementado

- Estrutura completa do repositório (`apps/mac-agent`, `apps/android-panel`, `configs`, `docs`, `scripts`, `test/fixtures`).
- `apps/mac-agent` (Go 1.23, módulo `deskpanel-agent`): CLI com os 7 subcomandos (`serve`, `pair`, `devices`, `revoke`, `status`, `doctor`, `version`); pacotes `protocol`, `actions`, `config`, `executor` (`FakeExecutor` funcional + `MacOSExecutor` stub), `auth` (hash/comparação de token), `pairing` (geração de código), `state`, `logging`, `api` (`/health`); 21 testes unitários reais cobrindo validação de catálogo/config, hashing de token, geração de código de pareamento e o handler de `/health`.
- `apps/android-panel` (React 18 + TypeScript + Vite 5 + Capacitor 6): casca do app, tipos do protocolo espelhando o Go, modelo de dados do layout (`DashboardConfig`), projeto nativo Android gerado (`apps/android-panel/android`), 1 teste de componente.
- `Makefile` com `setup`, `dev-agent`, `dev-android`, `test`, `lint`, `build-agent`, `build-apk`, `verify`.
- `docs/DECISIONS.md`, `docs/PROTOCOL.md`, `docs/SECURITY.md`, `docs/INSTALL-MACOS.md`, `docs/INSTALL-ANDROID.md`, `docs/tasks/0000-fase-0.md`.
- Scripts stub (`scripts/install-macos.sh`, `uninstall-macos.sh`, `build-apk.sh`) — apontam claramente para a Fase 5, ainda não implementados de verdade.

## `make verify` — resultado

```
lint-agent   : gofmt limpo, go vet sem avisos
lint-android : eslint sem erros, prettier sem divergências
test-agent   : 21 testes Go, todos ok (go test ./...)
test-android : 1 teste (vitest run), ok
build-agent  : go build ./cmd/deskpanel-agent — ok
build-android: tsc -b && vite build — ok (142 KB / 46 KB gzip)
```

`git diff --check` está limpo para todo arquivo escrito por nós. Os únicos avisos de espaço em branco vêm de arquivos gerados automaticamente pelo Capacitor dentro de `apps/android-panel/android/` (ex.: `gradlew.bat`, que usa CRLF por convenção do Windows) — não foram editados manualmente, e editá-los poderia quebrar o wrapper do Gradle.

## Riscos e limitações conhecidas

- Ambiente de build da IA (sandbox Linux do Cowork) não tem Homebrew nem acesso root; Go 1.23.9 e pnpm foram instalados localmente sem privilégios elevados (ver `docs/DECISIONS.md` ADR-0002). No Mac real do usuário, a instalação via Homebrew segue normal e é a recomendada para desenvolvimento contínuo.
- Nenhuma permissão do macOS (Acessibilidade/Automação) foi concedida ainda — não se aplica até a Fase 1/5.
- `MacOSExecutor` é só um stub que retorna `ACTION_NOT_ALLOWED` — execução real de ações é escopo da Fase 1.
- `apps/android-panel/android` foi gerado pelo Capacitor mas nunca compilado com Gradle real (exige Android SDK, que não está disponível neste sandbox) — a validação de que o projeto Android nativo builda de verdade fica para quando houver acesso a um ambiente com SDK/Gradle (idealmente o Mac real do usuário, na Fase 5).

## Próxima fase recomendada

Fase 1 — Agente e ações locais (`config.json` real, catálogo de ações carregado de disco, `MacOSExecutor` implementado para os 11 `kind`s do MVP, CLI `serve`/`status`/`doctor` funcionais de verdade).
