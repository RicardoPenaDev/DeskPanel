# Decisões de arquitetura (ADR) — DeskPanel

Registro de decisões técnicas tomadas durante o desenvolvimento. Toda decisão que se afasta ou detalha algo definido em `PROJECT.md` deve ser registrada aqui, com data, contexto e justificativa. Não editar decisões fixadas em `PROJECT.md` §4 sem autorização explícita do proprietário — este arquivo é para complementar, não substituir.

Formato de cada entrada:

```
## ADR-000X — Título
Data: AAAA-MM-DD
Status: proposta | aceita | substituída por ADR-000Y

Contexto:
...

Decisão:
...

Consequências:
...
```

---

## ADR-0001 — Decisões fixas do MVP (linha de base)
Data: 2026-09-15
Status: aceita

Contexto:
`PROJECT.md` §4 fixa a stack e as decisões estruturais do MVP antes de qualquer código ser escrito.

Decisão:
- Interface Android: React + TypeScript + Vite, empacotado com Capacitor.
- Agente macOS: Go, sem framework HTTP externo pesado (usar `net/http` da stdlib + roteador leve, decidir na Fase 2).
- Comunicação: WebSocket para execução de ações em tempo real; HTTP JSON versionado (`/api/v1`) para pareamento, listagem de ações e estado inicial.
- Sem banco de dados no MVP — JSON local com escrita atômica.
- Autenticação por pareamento manual + token por dispositivo, hash do token armazenado no Mac.
- Execução via `exec.CommandContext`, nunca `sh -c`/`bash -c`/`eval`.
- Porta padrão `38121`, sem exposição pública.

Consequências:
- Qualquer ação não listada em §7.4 exige uma nova ADR antes de ser implementada.
- Mudança de stack (ex.: trocar Capacitor por outra solução) exige ADR e autorização do proprietário.

---

## ADR-0002 — Ambiente de desenvolvimento (Fase 0)
Data: 2026-09-15
Status: aceita

Contexto:
O ambiente de build usado pela IA (sandbox Linux dentro do Cowork, montando a pasta `atalho` do Mac do usuário) não tem acesso root nem alcance de rede a `golang.org`/`dl.google.com` (bloqueados pelo proxy de saída do sandbox). Também não há Homebrew disponível nesse sandbox — Homebrew só existe no macOS nativo do usuário, fora do bridge.

Decisão:
- Go 1.23.9 (linux/arm64) foi baixado a partir dos releases oficiais espelhados pelo `actions/go-versions` (GitHub Releases, mesmos binários da distribuição oficial, apenas hospedados em outro domínio acessível) e instalado localmente em `$HOME/go-sdk` dentro do sandbox, sem privilégios de root.
- pnpm foi instalado via `npm install -g` com prefixo customizado (`$HOME/.npm-global`), também sem root.
- Essas instalações valem só para o sandbox usado pela IA nesta sessão. **No Mac real do usuário**, a instalação recomendada continua sendo via Homebrew (`brew install go`, `brew install pnpm` ou `corepack enable`), conforme perguntado e confirmado com o proprietário.

Consequências:
- `make verify` roda integralmente dentro do sandbox usando esse toolchain local.
- O usuário deve garantir Go e pnpm instalados (via Homebrew ou outro gerenciador) no Mac real antes de rodar `make dev-agent`, `make build-agent` ou instalar o LaunchAgent fora deste ambiente assistido.

---

## ADR-0003 — Biblioteca de WebSocket e desenho do socket administrativo
Data: 2026-09-15
Status: aceita

Contexto:
A stdlib do Go não implementa WebSocket. `PROJECT.md` §4 pede "sem framework HTTP externo pesado", deixando a escolha da lib de WS para a Fase 2. Também foi preciso decidir como `pair`/`devices`/`revoke`/`status` (comandos de CLI de curta duração) conversam com o processo `serve` (de longa duração), já que a config e o catálogo de dispositivos só existem na memória do `serve` — `PROJECT.md` §7.1 já previa isso via socket Unix.

Decisão:
- WebSocket: `github.com/coder/websocket` (sucessor do `nhooyr.io/websocket`) — API baseada em `context.Context`, sem dependências transitivas, mantida ativamente. É a única dependência externa do agente.
- Socket administrativo (`internal/adminsocket`): protocolo simples de request/response em JSON sobre `net.Listen("unix", ...)`, arquivo do socket com permissão 0700 no diretório. `serve` registra handlers para `pair`/`devices`/`revoke`/`status`; os subcomandos de CLI só discam o socket e formatam a resposta.
- `github.com/coder/websocket` não estava acessível via `proxy.golang.org` (bloqueado pelo proxy de saída do sandbox usado pela IA) — foi baixado com `GOPROXY=direct GOSUMDB=off`, indo direto ao GitHub via git. `go.sum` foi gerado normalmente a partir desse download. No Mac real do usuário, `GOPROXY`/`GOSUMDB` padrão devem funcionar sem ajuste.

Consequências:
- Qualquer diretório de dados do agente (`config.json`, `devices.json`, `agent.sock`) que já exista com permissão mais aberta que 0700 é corrigido automaticamente na primeira escrita — não basta confiar em `MkdirAll`, que não reajusta permissão de diretório pré-existente.
- `serve` precisa estar rodando para `pair`/`devices`/`revoke`/`status` funcionarem; sem ele, a CLI retorna um erro claro em vez de falhar silenciosamente.

---

## ADR-0004 — Armazenamento do token no Android (plugin Capacitor local) e "keep awake"
Data: 2026-09-15
Status: aceita

Contexto:
`PROJECT.md` §10.3/§12.8 exige que o token de acesso fique protegido pelo Android Keystore, nunca em `localStorage`, Capacitor Preferences, arquivo ou log. A API `Preferences` do Capacitor não é armazenamento seguro. Não há plugin oficial do Capacitor 6 para Keystore; os plugins de terceiros disponíveis não são auditáveis dentro deste projeto. `PROJECT.md` §10.1 também exige manter a tela ativa enquanto o painel estiver em primeiro plano.

Decisão:
- Token: plugin Capacitor **local** (não publicado como pacote separado) — `SecureTokenStoragePlugin.java`, registrado diretamente em `MainActivity.java` via `registerPlugin(...)`. Usa `androidx.security.crypto.EncryptedSharedPreferences` (`androidx.security:security-crypto:1.1.0-alpha06`, adicionada em `android/app/build.gradle`) com uma entrada por `deviceId`. Escrito em Java (não Kotlin) para não introduzir o toolchain Kotlin no projeto gerado pelo `cap add android`, que só tinha suporte a Java configurado.
- No lado TypeScript, `src/services/secureTokenStorage.ts` expõe a interface do plugin via `registerPlugin` do `@capacitor/core`, com fallback web (`secureTokenStorageWeb.ts`) só para `pnpm dev`/testes — esse fallback guarda o token em memória, sem persistência, e nunca deve rodar em produção (o app real roda sempre dentro do WebView Android).
- Preferências não secretas (IP/porta/nome do dispositivo, layout do painel) continuam via `@capacitor/preferences` (`storage/connectionConfig.ts`, `storage/layoutStorage.ts`), separado do token por desenho — nenhum dos dois módulos jamais lida com o token.
- "Keep awake": em vez de adicionar uma dependência (`@capacitor/keep-awake` ou similar), a tela é mantida ativa com `WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON` direto em `MainActivity.onCreate`, já que este é um dispositivo dedicado ao painel (nenhuma tela concorrente a proteger da tela sempre ligada).
- Orientação: `android:screenOrientation="landscape"` fixado em `AndroidManifest.xml` (além do `configChanges` já presente desde a Fase 0).

Consequências:
- `SecureTokenStoragePlugin.java`, a dependência `androidx.security:security-crypto` e as mudanças em `MainActivity.java`/`AndroidManifest.xml` **não puderam ser compiladas nem testadas neste ambiente** (sandbox sem Android SDK/Gradle — mesma limitação já registrada desde a Fase 0). A validação real só acontece na Fase 5, com Gradle de verdade no Mac do usuário e teste físico no Moto G60.
- O lado TypeScript do plugin (`secureTokenStorage.ts`, `secureTokenStorageWeb.ts`) é testado normalmente com Vitest, mockando `@capacitor/core`.
- Se o plugin nativo falhar ao compilar no Mac real, o sintoma mais provável é erro de resolução de `androidx.security:security-crypto` — conferir se `google()` está nos repositórios do projeto (já está, herdado do `cap add android`).
