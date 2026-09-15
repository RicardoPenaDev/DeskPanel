# Status — DeskPanel

Última atualização: 2026-09-15

## Feito

- **Fase 0 — Fundação**: repo, docs base, esqueletos Go e React/Capacitor, `make verify` funcionando.
- **Fase 1 — Agente e ações locais**: `MacOSExecutor` real (open_app, open_url, run_shortcut, volume, spotify/music) via `osascript`/`open`/`shortcuts`; `FakeExecutor` para testes; CLI carregando `config.json` de verdade.
- **Fase 2 — API, WebSocket e segurança**: servidor HTTP+WebSocket real (`serve`), pareamento com código temporário, tokens com hash salvo em `devices.json`, revogação imediata, rate limit, filtro de IP privado, heartbeat com dedup por `requestId`, socket administrativo para `pair`/`devices`/`revoke`/`status`. Testado com `make verify` + smoke test manual ponta a ponta contra o binário real.
- **Fase 3 — Painel Android funcional**: cliente HTTP (`httpClient.ts`) e WebSocket (`wsClient.ts`, com reconexão 1s/2s/4s/8s/15s e heartbeat) espelhando o protocolo do agente; storage não secreto de conexão/layout via Capacitor Preferences; token de acesso via plugin Capacitor **local** (`SecureTokenStoragePlugin.java`, Android Keystore/EncryptedSharedPreferences — ver ADR-0004); tela de pareamento (`PairingScreen`); painel principal com grade 4×2, duas páginas, troca por gesto, indicador de conexão/página e os 5 estados de botão (normal/pressionado/executando/sucesso/erro), incluindo toque prolongado para ações perigosas; hook `useDeskPanelConnection` orquestrando tudo; orientação paisagem e tela sempre ativa (`MainActivity`). `configs/config.example.json` ampliado para as 16 ações do layout padrão (§10.4) e validado contra o agente real (`serve` sobe com "16 ações registradas"). 64 testes Vitest + `tsc -b` + `eslint` + `prettier` passando.

## Não feito / próximo

- **Fase 4 — Editor do painel**: criar/excluir páginas, adicionar/remover/reorganizar botões, trocar ação/nome/ícone/cor, migração de schema.
- Tela de Configurações (§10.2-D: refazer pareamento, vibração, modo imersivo, diagnóstico) ainda não existe — não fazia parte das entregas da Fase 3.
- `SecureTokenStoragePlugin.java`, a dependência `androidx.security:security-crypto` e os ajustes em `MainActivity.java`/`AndroidManifest.xml` **não foram compilados nem testados** neste ambiente (sem Android SDK/Gradle) — validação real só na Fase 5, com Gradle de verdade e teste físico no Moto G60.
- `keystroke`, `screen_lock`, `display_sleep` ainda não implementados no `MacOSExecutor` (ficam pendurados como `ACTION_NOT_ALLOWED` até então).
- Ícones ainda são só texto/placeholder — pacote de ícones locais fica para a Fase 6 (polimento).
- Ambiente de build da IA não tem Homebrew/root — Go e pnpm instalados localmente sem privilégios (ver ADR-0002); no Mac real, usar Homebrew normalmente. Neste ambiente, `pnpm` só funciona via `corepack pnpm` (não afeta o Mac real, que tem pnpm no PATH via Homebrew/corepack normal).
