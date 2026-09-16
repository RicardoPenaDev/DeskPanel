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

---

## ADR-0005 — Editor do painel: reorganizar por toque, ícones locais, trava de segurança e backup de layout
Data: 2026-09-15
Status: aceita

Contexto:
`PROJECT.md` §17 (Fase 4) e §10.2-C pedem um editor com: reorganizar botões "por arrastar e soltar"; trocar ícone "dentre os ícones empacotados"; marcar uma ação como perigosa para exigir toque prolongado; persistência com migração de schema; e o §17 exige que "config corrompida usa backup ou padrão sem travar". Nenhum desses itens tinha uma implementação prévia (Fase 3 só tinha o layout padrão fixo, sem editor).

Decisão:
- **Reorganizar**: em vez de arrastar-e-soltar via HTML5 Drag and Drop (historicamente pouco confiável em WebViews Android por toque), o editor usa um modo "Reorganizar botões" com interação de tocar-para-selecionar / tocar-para-trocar: o usuário toca em um botão para selecioná-lo, depois toca no destino (vazio ou ocupado) para mover ou trocar. É mais confiável em touch, mais simples de testar (sem simular gestos de arrastar em jsdom) e cumpre a mesma função (reorganizar a grade) pedida pelo `PROJECT.md`.
- **Ícones**: pacote local mínimo (`components/Icon.tsx`), ~17 formas SVG desenhadas à mão, sem nenhuma dependência externa — chaves alinhadas com o campo `icon` já usado em `configs/config.example.json`. Resolve tanto §10.1 ("ícones empacotados localmente") quanto a entrega da Fase 4 ("trocar ícone dentre os ícones empacotados"), que antes só existia como texto no `STATUS.md` ("placeholder").
- **Trava de segurança**: uma ação cujo `kind` no agente exige toque prolongado (`screen_lock`, `display_sleep` — `ActionSummary.requireLongPress === true`) não pode ter esse requisito destravado pelo editor. O checkbox "Exigir toque prolongado" fica marcado e desabilitado nesse caso; o editor só permite ao usuário **adicionar** a exigência a uma ação que originalmente não teria (endurecer, nunca afrouxar). Isso é reforçado em duas camadas: `EditorScreen` força `requireLongPress: true` ao salvar, e `DashboardScreen` (Fase 3) já fazia um OR entre `button.requireLongPress` e `catalogAction.requireLongPress` ao renderizar — então mesmo um bug futuro no editor não bypassaria a exigência de toque prolongado na tela real.
- **Persistência/backup/migração**: `storage/layoutStorage.ts` agora mantém uma segunda chave de backup (`deskpanel.layout.backup`), promovida só depois de uma gravação válida bem-sucedida. Ao carregar: cópia principal válida → usa; senão cópia de backup válida → usa; senão layout padrão. `CURRENT_SCHEMA_VERSION = 1` é o único aceito hoje; um schemaVersion futuro entra como um novo caso de migração sem mexer no resto do módulo — não inventamos uma v2 fictícia agora.
- Páginas/botões continuam usando só o perfil único `"default"` (`PROJECT.md` não pede múltiplos perfis na Fase 4) — criar/excluir afeta só `profiles[0].pages`.

Consequências:
- Se o toque-para-reorganizar se mostrar pouco intuitivo no teste físico do Moto G60 (Fase 5), trocar para arrastar-e-soltar de verdade é uma mudança isolada em `EditorScreen`/`DashboardButton`, sem impacto em armazenamento ou protocolo.
- O conjunto de 17 ícones é deliberadamente pequeno; um pacote maior/mais bonito fica para a Fase 6 (polimento), mas já não há mais placeholder de texto puro nos botões.
- Um `schemaVersion` novo exigirá escrever a função de migração real quando o formato realmente mudar — o "seam" já existe, a migração em si não.

## ADR-0006 — Scripts de instalação macOS/Android: não-destrutivo por padrão, LaunchAgent com restart-on-crash
Data: 2026-09-16
Status: aceita

Contexto:
`PROJECT.md` §14.2/§14.3 e §17 (Fase 5) pedem `scripts/install-macos.sh`, `scripts/uninstall-macos.sh` e `scripts/build-apk.sh` reais. §19.3 proíbe qualquer instalação real, carregamento de LaunchAgent real, pareamento com o Moto G60 físico ou execução de ação real sem autorização explícita do proprietário — então os scripts precisavam ser escritos e validados no que fosse possível (sintaxe, `--help`, guarda de plataforma, o caminho `build-apk.sh --web-only` completo) sem serem de fato executados contra o Mac real neste ambiente (sandbox Linux, sem `launchctl`/Gradle/Android SDK).

Decisão:
- **`install-macos.sh`**: por padrão nunca sobrescreve `config.json` existente (só cria a partir de `configs/config.example.json` se ausente); `--force` é obrigatório para sobrescrever. Aceita `--binary <caminho>` para pular `go build` (permite instalar um binário cross-compilado). O LaunchAgent (`dev.ricardopena.deskpanel.agent.plist`) é gerado com `$INSTALLED_BINARY`/`$HOME` resolvidos em tempo de instalação — nunca um nome de usuário fixo no template. `KeepAlive.SuccessfulExit = false`: o launchd reinicia o processo se ele cair/crashar, mas não fica reiniciando em loop se o agente sair com código 0 de propósito (ex.: um futuro comando administrativo de "parar o serviço"). `RunAtLoad = true` garante que o agente volta sozinho após logout/login — é o mecanismo que cobre o critério de saída da Fase 5 "agente volta após logout/login", sem precisar de um daemon separado de supervisão.
- **`uninstall-macos.sh`**: remove serviço + binário por padrão, preserva `config.json`/`devices.json`/`state.json`/logs. Remover dados exige `--purge`, que por sua vez pede confirmação interativa (`read -p`) a menos que `--yes`/`-y` também seja passado — evita apagar pareamentos/configuração por engano rodando o comando errado.
- **`build-apk.sh`**: flag `--web-only` roda só `pnpm install`/`test`/`build`/`cap sync`, sem invocar Gradle — permitiu validar de ponta a ponta o pipeline de build neste sandbox (sem Android SDK) e serve como verificação rápida de regressão em CI que não tenha o SDK Android instalado. As etapas de teste/lint continuam sendo as mesmas do `Makefile` (`pnpm test`, `pnpm build`), sem duplicar lógica.
- Todos os três scripts: `set -euo pipefail`, sem nenhum shell intermediário sobre entrada externa (as únicas entradas aceitas são flags fixas, nunca comando/caminho/URL livre — mesma disciplina de segurança do agente Go, aplicada aqui aos scripts de instalação).

Consequências:
- A execução real dos três scripts contra o Mac do usuário, o carregamento do LaunchAgent de verdade, o pareamento com o Moto G60 físico e o build via Gradle real ficam pendentes de autorização explícita e de rodar num Mac de verdade — não fazem parte desta mudança. `docs/STATUS.md` reflete isso como pendência da Fase 5.
- Se o critério "Moto G60 reconecta automaticamente" exigir mudança de comportamento (ex.: reconexão mais agressiva no `wsClient`), isso é um ajuste em `apps/android-panel`, não nos scripts de instalação.
- `--purge` sem `--yes` é interativo por desenho; scripts de automação/CI que precisarem purgar dados devem passar `--yes` explicitamente.

## ADR-0007 — Polimento do MVP: rotação de logs, doctor completo, haptics oficial, controles nativos de tela
Data: 2026-09-16
Status: aceita

Contexto:
`PROJECT.md` §17 (Fase 6) pede: correções do teste físico (impossível sem
hardware real — ver ressalva abaixo), feedback háptico, diagnóstico de
conexão, rotação de logs, acessibilidade básica, mais ícones locais,
revisão de segurança e versão `0.1.0`. §10.2-D especifica a tela de
Configurações completa (reconectar, refazer pareamento, vibração, manter
tela ligada, modo imersivo, brilho reduzido, diagnóstico, versões).

Decisão:
- **Rotação de logs**: implementada dentro do próprio agente
  (`internal/logging/rotate.go`, `RotatingWriter`), não delegada ao
  `launchd`/shell — redirecionar `stdout` para um arquivo via
  `StandardOutPath` no LaunchAgent (como a Fase 5 já fazia) nunca rotaciona
  sozinho. `cmdServe` agora abre o logger com `logging.NewFile` em
  `~/Library/Logs/DeskPanel/agent.log`, ~5 MiB por arquivo, até 3 arquivos
  (PROJECT.md §13); se o arquivo não puder ser aberto, cai para stdout e
  reporta o erro — nunca impede o agente de subir.
- **`doctor` completo**: as checagens que faltavam (§13) foram adicionadas
  sem tocar nas existentes — porta livre/ocupada (distinguindo "é o
  próprio agente" via socket administrativo de "outro processo"), status
  do LaunchAgent (`launchctl print`, só no macOS), endereços IPv4 locais
  alcançáveis (reaproveita `internal/netguard.IsPrivateOrLoopback`) e
  versão do protocolo/agente.
- **Vibração**: plugin oficial `@capacitor/haptics` (mesma major version 6
  dos outros plugins Capacitor já usados), não um plugin nativo próprio —
  ao contrário do token (Keystore) e dos controles de janela abaixo, não
  há nenhuma lógica específica do DeskPanel aqui, só chamar uma API padrão
  do sistema. `services/haptics.ts` expõe `vibrateSuccess`/`vibrateError`,
  cada uma engolindo erro silenciosamente (aparelho sem motor de vibração
  não pode derrubar a execução de uma ação). Controlado por
  `AppSettings.vibrationEnabled` (padrão ligado), passado como prop
  `vibrationEnabled` até `DashboardButton` — sem estado global implícito,
  mesmo padrão de dependência explícita já usado no resto do app.
- **Manter tela ligada / modo imersivo / brilho reduzido**: um plugin
  Capacitor local novo, `DeviceControlPlugin.java` (mesmo padrão do
  `SecureTokenStoragePlugin` da Fase 3 — registrado em `MainActivity`, sem
  publicar como pacote separado). Nenhum dos três pede permissão especial
  do Android: "keep awake" e "imersivo" são flags da própria janela
  (`WindowManager.LayoutParams`/`WindowInsetsController`), e o brilho
  reduzido usa `LayoutParams.screenBrightness` (atributo da janela do
  app), não `Settings.System` — por isso não precisa de
  `android.permission.WRITE_SETTINGS`, que exigiria um fluxo de permissão
  especial fora do MVP. Modo imersivo usa `WindowInsetsController` em
  Android 11+ (API 30) com fallback para as flags legadas de
  `SystemUiVisibility` em versões mais antigas.
- **Tela de Configurações** (`features/settings/SettingsScreen.tsx`): usa
  estado local com um botão "Salvar" explícito para nome/host/porta (mesmo
  padrão de `PairingScreen`, não o padrão totalmente controlado do
  `EditorScreen` — aqui não há necessidade de refletir cada tecla de volta
  no app em tempo real). Os toggles (vibração/keep-awake/imersivo/brilho)
  aplicam na hora: persistem via `useDeskPanelConnection.updateAppSettings`
  e, para os três que têm efeito nativo, chamam `DeviceControlPlugin` na
  sequência. "Reconectar" reaproveita a mesma instância de `WsClient`
  (`disconnect()`+`connect()`) em vez de recriar o hook inteiro.  "Refazer
  pareamento" limpa só o token (mantém deviceId/host/porta/nome já
  digitados) e volta para `PairingScreen` — não é preciso digitar tudo de
  novo, só o código. "Diagnóstico de conexão" reaproveita `checkHealth` já
  existente, medindo a latência no cliente com `performance.now()`.
- **Acessibilidade básica**: focus-visible global para todo elemento
  interativo (`button`, `input`, `[role="tab"]`, `[tabindex]`) em
  `global.css` — antes só os campos de formulário tinham contorno de foco
  visível; `role="alert"` no erro do botão do painel, para leitor de tela
  anunciar imediatamente. O resto (aria-label nos botões, `aria-hidden` nos
  ícones decorativos, `role="status"`/`aria-live` no indicador de conexão,
  `role="tablist"`/`role="tab"` no indicador de página, `lang="pt-BR"`,
  toque mínimo de 48px) já vinha das Fases 3/4 e foi só conferido, não
  refeito.
- **Versão 0.1.0**: `Makefile` ganhou `VERSION := 0.1.0` e passa
  `-ldflags "-X main.version=$(VERSION)"` em `build-agent`;
  `scripts/install-macos.sh` faz o mesmo na sua própria chamada de
  `go build`. `apps/android-panel/package.json`/`version.ts` já estavam em
  `0.1.0` desde o bootstrap.

Consequências:
- **"Correções do teste físico" não foi feito** — não existe teste físico
  ainda para corrigir (Fase 5 não rodou de verdade no Mac/Moto G60 real,
  por exigir autorização explícita do proprietário, `PROJECT.md` §19.3).
  Essa entrega da Fase 6 fica pendente até depois da instalação real.
- `DeviceControlPlugin.java` e o registro dele em `MainActivity.java` têm
  a mesma ressalva do `SecureTokenStoragePlugin` (ADR-0004): não compilados
  nem testados neste ambiente (sem Android SDK/Gradle). A revisão de
  segurança (nova seção em `docs/SECURITY.md`) documenta essa ressalva
  explicitamente.
- Tag local `v0.1.0` **não foi criada** — `PROJECT.md` §17 exige
  autorização explícita do proprietário para isso, separada da autorização
  geral de avançar de fase.
- Os critérios de aceite do §18 que dependem de hardware físico (APK
  instalado no Moto G60, teste contínuo de duas horas, etc.) continuam
  pendentes — ver `docs/STATUS.md`.
