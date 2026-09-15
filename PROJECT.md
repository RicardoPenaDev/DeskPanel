# DeskPanel — painel Android para controlar o macOS

## 1. Finalidade deste documento

Este arquivo é a fonte de verdade para o desenvolvimento do **DeskPanel**, um sistema que transforma um celular Android dedicado — inicialmente um Motorola Moto G60 — em um painel de atalhos semelhante a um Stream Deck para controlar um Mac.

A IA responsável pelo desenvolvimento deve:

1. Ler este documento completamente antes de alterar qualquer arquivo.
2. Implementar o projeto por fases, na ordem definida aqui.
3. Não substituir as decisões arquiteturais sem registrar a justificativa em `docs/DECISIONS.md`.
4. Não criar endpoints de execução arbitrária de shell.
5. Não expor o agente diretamente à internet.
6. Executar testes e registrar evidências antes de considerar uma fase concluída.
7. Interromper e pedir decisão humana quando uma mudança afetar segurança, escopo, arquitetura ou compatibilidade.

---

## 2. Visão do produto

O DeskPanel terá duas aplicações:

- **DeskPanel Android:** interface visual instalada no Moto G60, usada na horizontal sobre a mesa.
- **DeskPanel Agent:** serviço instalado no Mac, responsável por receber ações autorizadas e executá-las no macOS.

Fluxo principal:

```text
Usuário toca em um botão no Moto G60
        ↓
Aplicativo envia apenas o ID da ação
        ↓
Agente valida dispositivo, token e ação
        ↓
Agente executa uma ação previamente cadastrada
        ↓
Resultado retorna ao celular
        ↓
Botão mostra sucesso ou erro
```

O celular nunca enviará comandos de terminal, AppleScript, nomes de aplicativos ou URLs para execução. Ele enviará somente um `actionId`. Toda ação executável será definida e validada no Mac.

---

## 3. Objetivos

### 3.1 Objetivos do MVP

- Funcionar no Motorola Moto G60 em modo paisagem.
- Controlar um Mac conectado à mesma rede local.
- Abrir aplicativos e URLs previamente configurados.
- Executar Atalhos do macOS previamente configurados.
- Controlar volume e reprodução de mídia.
- Bloquear a tela ou apagar o monitor mediante confirmação por toque prolongado.
- Mostrar se o Mac está online.
- Exibir retorno visual de sucesso, andamento e falha.
- Suportar várias páginas de botões.
- Permitir reorganizar botões e escolher ações existentes.
- Reconectar automaticamente após perda temporária de Wi-Fi.
- Iniciar o agente automaticamente após login no Mac.
- Manter a tela do Android ativa enquanto o painel estiver aberto.
- Trabalhar integralmente na rede local, sem depender de servidor externo.

### 3.2 Objetivos posteriores ao MVP

- Descoberta automática do Mac por mDNS/Bonjour.
- Exibição da música atual, artista, capa e estado de reprodução.
- Perfis automáticos conforme o aplicativo ativo no Mac.
- Ações encadeadas e macros.
- Ícones personalizados enviados pelo usuário.
- Aplicativo de menu nativo no macOS.
- Sincronização opcional das páginas entre dispositivos.
- Integração opcional com Home Assistant e serviços do homelab.
- Acesso remoto opcional apenas por Tailscale, nunca por porta pública.

### 3.3 Fora do escopo do MVP

- Publicação na Google Play ou Mac App Store.
- Controle remoto pela internet aberta.
- Espelhamento da tela do Mac.
- Controle de mouse por touchpad.
- Transmissão de áudio ou vídeo.
- Execução de shell livre digitado pelo celular.
- Sistema multiusuário ou serviço SaaS.
- Suporte oficial a Windows e Linux.
- Marketplace de plugins.

---

## 4. Decisões técnicas fixadas

| Componente | Decisão |
|---|---|
| Interface Android | React + TypeScript + Vite |
| Empacotamento Android | Capacitor |
| Linguagem do agente macOS | Go |
| Comunicação em tempo real | WebSocket |
| API auxiliar | HTTP JSON versionado em `/api/v1` |
| Banco de dados | Nenhum no MVP; arquivos JSON locais com escrita atômica |
| Armazenamento do layout | Local no Android |
| Armazenamento das ações | Configuração autoritativa no Mac |
| Autenticação | Pareamento manual e token individual por dispositivo |
| Execução no macOS | `exec.Command` com binários e argumentos previamente autorizados |
| Inicialização do agente | `launchd` por meio de um LaunchAgent do usuário |
| Descoberta no MVP | IP e porta configurados manualmente |
| Porta padrão | TCP `38121` |
| Acesso externo | Proibido no MVP |
| Telemetria externa | Nenhuma |

As dependências devem usar versões estáveis e compatíveis disponíveis no início da implementação. As versões efetivas devem ficar fixadas em `go.mod`, `go.sum`, `package.json` e `pnpm-lock.yaml`. Não usar dependências sem versão travada no resultado final.

---

## 5. Arquitetura

```text
┌───────────────────────────────────────┐
│ Moto G60 — DeskPanel Android          │
│                                       │
│ React + TypeScript                    │
│ Capacitor                             │
│ Layout e preferências locais          │
│ Token protegido pelo Android Keystore │
└───────────────────┬───────────────────┘
                    │ HTTP + WebSocket
                    │ rede local
┌───────────────────▼───────────────────┐
│ Mac — DeskPanel Agent                 │
│                                       │
│ API HTTP                              │
│ servidor WebSocket                    │
│ autenticação e rate limit             │
│ catálogo autoritativo de ações        │
│ executores macOS                      │
│ estado e logs locais                  │
└───────────────────┬───────────────────┘
                    │
        ┌───────────┼─────────────┐
        ▼           ▼             ▼
   `open`      `shortcuts`    `osascript`
```

### 5.1 Princípios

- **Local first:** o sistema funciona sem internet.
- **Servidor autoritativo:** o Mac decide o que pode ser executado.
- **Menor privilégio:** nenhum comando arbitrário será aceito.
- **Falha segura:** ação desconhecida, token inválido ou payload incorreto resultam em rejeição.
- **Interface desacoplada:** o Android não precisa conhecer detalhes de AppleScript ou caminhos do macOS.
- **Protocolo versionado:** mudanças incompatíveis exigem nova versão.
- **Observabilidade local:** logs suficientes para diagnóstico, sem registrar segredos.

---

## 6. Estrutura do repositório

```text
deskpanel/
├── PROJECT.md
├── README.md
├── Makefile
├── .editorconfig
├── .gitignore
├── apps/
│   ├── android-panel/
│   │   ├── android/
│   │   ├── src/
│   │   │   ├── app/
│   │   │   ├── components/
│   │   │   ├── features/
│   │   │   │   ├── connection/
│   │   │   │   ├── dashboard/
│   │   │   │   ├── editor/
│   │   │   │   ├── pairing/
│   │   │   │   └── settings/
│   │   │   ├── protocol/
│   │   │   ├── services/
│   │   │   ├── storage/
│   │   │   └── styles/
│   │   ├── capacitor.config.ts
│   │   ├── package.json
│   │   ├── pnpm-lock.yaml
│   │   └── vite.config.ts
│   └── mac-agent/
│       ├── cmd/
│       │   └── deskpanel-agent/
│       │       └── main.go
│       ├── internal/
│       │   ├── actions/
│       │   ├── api/
│       │   ├── auth/
│       │   ├── config/
│       │   ├── executor/
│       │   ├── logging/
│       │   ├── pairing/
│       │   ├── protocol/
│       │   ├── state/
│       │   └── websocket/
│       ├── go.mod
│       └── go.sum
├── configs/
│   └── config.example.json
├── docs/
│   ├── DECISIONS.md
│   ├── STATUS.md
│   ├── SECURITY.md
│   ├── PROTOCOL.md
│   ├── INSTALL-MACOS.md
│   ├── INSTALL-ANDROID.md
│   └── tasks/
├── scripts/
│   ├── install-macos.sh
│   ├── uninstall-macos.sh
│   └── build-apk.sh
└── test/
    └── fixtures/
```

---

## 7. DeskPanel Agent para macOS

### 7.1 Modos de execução

O mesmo binário deverá oferecer:

```bash
deskpanel-agent serve
deskpanel-agent pair
deskpanel-agent devices
deskpanel-agent revoke <device-id>
deskpanel-agent status
deskpanel-agent doctor
deskpanel-agent version
```

- `serve`: inicia API, WebSocket e execução de ações.
- `pair`: abre uma janela de pareamento por cinco minutos e mostra um código de seis dígitos.
- `devices`: lista dispositivos pareados sem mostrar tokens.
- `revoke`: remove a autorização de um dispositivo.
- `status`: informa porta, estado e conexões.
- `doctor`: verifica configuração, permissões, arquivos, porta e comandos do macOS.
- `version`: mostra versão e versão do protocolo.

O serviço `serve` deverá expor um socket Unix local para receber os comandos administrativos. O socket não poderá aceitar conexões de outros usuários.

Local sugerido:

```text
~/Library/Application Support/DeskPanel/agent.sock
```

### 7.2 Diretórios macOS

```text
~/Library/Application Support/DeskPanel/
├── bin/deskpanel-agent
├── config.json
├── devices.json
├── state.json
└── agent.sock

~/Library/Logs/DeskPanel/
└── agent.log

~/Library/LaunchAgents/
└── dev.ricardopena.deskpanel.agent.plist
```

Requisitos:

- Arquivos que contenham tokens devem usar permissão `0600`.
- Diretórios privados devem usar `0700`.
- Escritas de JSON devem ser atômicas: escrever em arquivo temporário, executar `fsync` quando aplicável e renomear.
- O agente não deve exigir execução como `root`.

### 7.3 Configuração do agente

Exemplo de `config.json`:

```json
{
  "schemaVersion": 1,
  "server": {
    "listenAddress": "0.0.0.0",
    "port": 38121,
    "pairingWindowSeconds": 300,
    "maxConnections": 5
  },
  "security": {
    "allowPublicNetworks": false,
    "requestsPerMinute": 120,
    "failedAuthLimit": 10
  },
  "actions": [
    {
      "id": "app.chrome",
      "label": "Chrome",
      "icon": "chrome",
      "kind": "open_app",
      "parameters": {
        "application": "Google Chrome"
      }
    },
    {
      "id": "app.whatsapp",
      "label": "WhatsApp",
      "icon": "message-circle",
      "kind": "open_app",
      "parameters": {
        "application": "WhatsApp"
      }
    },
    {
      "id": "media.spotify.playpause",
      "label": "Play/Pause",
      "icon": "play-pause",
      "kind": "spotify_control",
      "parameters": {
        "operation": "play_pause"
      }
    },
    {
      "id": "volume.up",
      "label": "Volume +",
      "icon": "volume-2",
      "kind": "volume_delta",
      "parameters": {
        "delta": 5
      }
    },
    {
      "id": "shortcut.work",
      "label": "Modo trabalho",
      "icon": "briefcase-business",
      "kind": "run_shortcut",
      "parameters": {
        "shortcut": "Modo Trabalho"
      }
    }
  ]
}
```

### 7.4 Tipos de ação permitidos no MVP

| `kind` | Implementação | Restrições |
|---|---|---|
| `open_app` | `/usr/bin/open -a <app>` | Aplicativo definido no Mac |
| `open_url` | `/usr/bin/open <url>` | URL fixa definida no Mac; somente `http` ou `https` |
| `run_shortcut` | `/usr/bin/shortcuts run <nome>` | Nome fixo definido no Mac |
| `keystroke` | AppleScript com template fixo | Somente teclas e modificadores permitidos |
| `volume_delta` | AppleScript com cálculo limitado | Resultado entre 0 e 100 |
| `volume_set` | AppleScript | Valor fixo entre 0 e 100 |
| `mute_toggle` | AppleScript | Sem parâmetros do cliente |
| `spotify_control` | AppleScript para Spotify | `play_pause`, `next` ou `previous` |
| `music_control` | AppleScript para Apple Music | `play_pause`, `next` ou `previous` |
| `screen_lock` | Atalho fixo do macOS | Exigir toque prolongado no Android |
| `display_sleep` | `/usr/bin/pmset displaysleepnow` | Exigir toque prolongado no Android |

Não implementar no MVP:

- `shell`
- `exec`
- `script` livre
- argumento enviado pelo cliente
- caminho de arquivo enviado pelo cliente
- URL enviada pelo cliente

Todo processo deve ser criado com `exec.CommandContext`. Não usar `sh -c`, `bash -c`, `zsh -c`, `eval` ou concatenação de comando.

### 7.5 Executor abstrato

Criar uma interface interna para permitir testes sem executar ações reais:

```go
type Executor interface {
    Execute(ctx context.Context, action Action) ExecutionResult
}
```

Implementações:

- `MacOSExecutor`: execução real.
- `FakeExecutor`: testes unitários e de integração.

Cada ação deve possuir:

- timeout máximo;
- código de resultado estável;
- duração em milissegundos;
- saída interna limitada;
- mensagem segura para o cliente;
- log técnico sem token e sem dados sensíveis.

Timeout padrão: 10 segundos. Nenhum processo filho pode sobreviver ao cancelamento do contexto.

---

## 8. API HTTP

Base URL:

```text
http://<ip-do-mac>:38121/api/v1
```

### 8.1 Endpoints

| Método | Endpoint | Autenticação | Finalidade |
|---|---|---|---|
| `GET` | `/health` | Não | Confirma que o agente está acessível; retorna dados mínimos |
| `POST` | `/pair` | Código temporário | Pareia um novo dispositivo |
| `GET` | `/actions` | Bearer token | Lista ações autorizadas e metadados visuais |
| `GET` | `/state` | Bearer token | Estado inicial do Mac e da mídia |
| `GET` | `/version` | Não | Versões do agente e protocolo |
| `GET` | `/ws` | Primeira mensagem WebSocket | Canal em tempo real |

Não criar endpoint HTTP genérico para executar comandos. A execução normal será feita pelo protocolo WebSocket.

### 8.2 Resposta de saúde

```json
{
  "status": "ok",
  "service": "deskpanel-agent",
  "protocolVersion": 1
}
```

Não incluir hostname, usuário, caminhos, IPs internos, tokens ou lista de ações no endpoint público.

### 8.3 Pareamento

O comando abaixo abre uma janela de pareamento de cinco minutos:

```bash
deskpanel-agent pair
```

Saída esperada:

```text
DeskPanel pairing enabled for 5 minutes.
Code: 483921
```

Requisição Android:

```http
POST /api/v1/pair
Content-Type: application/json
```

```json
{
  "code": "483921",
  "deviceId": "uuid-gerado-no-android",
  "deviceName": "Moto G60 do Ricardo",
  "appVersion": "0.1.0",
  "protocolVersion": 1
}
```

Resposta:

```json
{
  "deviceId": "uuid-gerado-no-android",
  "accessToken": "token-aleatorio-de-alta-entropia",
  "protocolVersion": 1
}
```

Regras:

- Código numérico de seis dígitos gerado criptograficamente.
- Expiração em cinco minutos.
- Uso único.
- No máximo cinco tentativas por código.
- Token gerado com pelo menos 32 bytes aleatórios.
- O Mac armazena apenas hash seguro do token.
- Uma entrada por dispositivo em `devices.json`.
- O Android armazena o token no Android Keystore por um plugin Capacitor local e mínimo.
- Nunca registrar código completo ou token nos logs.

---

## 9. Protocolo WebSocket

### 9.1 Conexão

```text
ws://<ip-do-mac>:38121/api/v1/ws
```

Após abrir o socket, o Android tem cinco segundos para enviar `auth.hello`. Até a autenticação ser concluída, nenhuma outra mensagem será aceita.

### 9.2 Envelope padrão

```json
{
  "type": "action.execute",
  "protocolVersion": 1,
  "requestId": "uuid",
  "timestamp": "2026-09-15T18:30:00Z",
  "payload": {}
}
```

### 9.3 Autenticação

Cliente:

```json
{
  "type": "auth.hello",
  "protocolVersion": 1,
  "requestId": "uuid",
  "timestamp": "2026-09-15T18:30:00Z",
  "payload": {
    "deviceId": "uuid-do-dispositivo",
    "accessToken": "token",
    "appVersion": "0.1.0"
  }
}
```

Servidor:

```json
{
  "type": "auth.accepted",
  "protocolVersion": 1,
  "requestId": "mesmo-uuid",
  "timestamp": "2026-09-15T18:30:00Z",
  "payload": {
    "agentVersion": "0.1.0",
    "macName": "MacBook de Ricardo"
  }
}
```

### 9.4 Executar ação

Cliente:

```json
{
  "type": "action.execute",
  "protocolVersion": 1,
  "requestId": "uuid",
  "timestamp": "2026-09-15T18:31:00Z",
  "payload": {
    "actionId": "app.chrome"
  }
}
```

Servidor durante a execução:

```json
{
  "type": "action.started",
  "protocolVersion": 1,
  "requestId": "mesmo-uuid",
  "timestamp": "2026-09-15T18:31:00Z",
  "payload": {
    "actionId": "app.chrome"
  }
}
```

Servidor ao terminar:

```json
{
  "type": "action.result",
  "protocolVersion": 1,
  "requestId": "mesmo-uuid",
  "timestamp": "2026-09-15T18:31:00Z",
  "payload": {
    "actionId": "app.chrome",
    "status": "success",
    "durationMs": 84,
    "errorCode": null,
    "message": null
  }
}
```

### 9.5 Estado e conectividade

Mensagens adicionais:

- `ping`
- `pong`
- `state.snapshot`
- `state.changed`
- `error`

Heartbeat:

- servidor envia `ping` a cada 20 segundos;
- cliente responde `pong` em até 10 segundos;
- após duas falhas, o socket é encerrado;
- Android inicia reconexão com backoff e jitter: 1 s, 2 s, 4 s, 8 s e máximo de 15 s;
- ao reconectar, requisitar um novo `state.snapshot`.

### 9.6 Códigos de erro

```text
AUTH_REQUIRED
AUTH_INVALID
AUTH_REVOKED
PAIRING_CLOSED
PAIRING_CODE_INVALID
PAIRING_CODE_EXPIRED
PROTOCOL_UNSUPPORTED
INVALID_MESSAGE
ACTION_NOT_FOUND
ACTION_NOT_ALLOWED
ACTION_TIMEOUT
ACTION_FAILED
RATE_LIMITED
INTERNAL_ERROR
```

As mensagens mostradas ao usuário devem ser amigáveis. Detalhes internos ficam somente no log local.

---

## 10. Aplicativo Android

### 10.1 Requisitos gerais

- Orientação bloqueada em paisagem.
- Interface utilizável a partir de Android 8, com testes obrigatórios no Moto G60.
- Tela cheia imersiva opcional.
- `keep awake` habilitado enquanto o painel estiver visível.
- Botões com área de toque mínima de 48 dp.
- Interface responsiva a diferentes proporções de tela.
- Tema escuro como padrão.
- Sem dependência de internet após a instalação.
- Ícones empacotados localmente.
- Nenhum tracker ou ferramenta de analytics.
- Vibração curta opcional em ação bem-sucedida.

### 10.2 Telas

#### A. Primeiro acesso

Campos:

- Nome do dispositivo, preenchido inicialmente como `Moto G60 do Ricardo`.
- IP ou hostname do Mac.
- Porta, preenchida com `38121`.
- Código de pareamento.

Ações:

- Testar conexão.
- Parear.
- Mostrar erro objetivo.

#### B. Painel principal

- Grade inicial: quatro colunas por duas linhas em paisagem.
- Suporte a várias páginas.
- Troca de página por gesto horizontal.
- Indicador discreto da página atual.
- Indicador de conexão no topo.
- Cada botão exibe ícone e nome.
- Estado visual: normal, pressionado, executando, sucesso e erro.
- Sucesso volta ao estado normal após aproximadamente 600 ms.
- Erro mostra mensagem curta e acessível.

#### C. Editor

Aberto por toque prolongado em área vazia ou opção nas configurações.

Permitir:

- adicionar botão;
- remover botão;
- escolher ação da lista recebida do Mac;
- trocar nome visual local;
- trocar ícone dentre os ícones empacotados;
- trocar cor dentre uma paleta limitada;
- reorganizar por arrastar e soltar;
- mover entre páginas;
- marcar ação como perigosa para exigir toque prolongado.

Não permitir definir comandos ou parâmetros executáveis.

#### D. Configurações

- Nome do dispositivo.
- IP/hostname do Mac.
- Porta.
- Reconectar.
- Refazer pareamento.
- Vibração ligada/desligada.
- Manter tela ligada.
- Modo imersivo.
- Brilho reduzido opcional enquanto o painel estiver ativo.
- Diagnóstico de conexão.
- Versões do app, agente e protocolo.

### 10.3 Modelo de dados do layout

```ts
type DashboardConfig = {
  schemaVersion: 1;
  activeProfileId: string;
  profiles: DashboardProfile[];
};

type DashboardProfile = {
  id: string;
  name: string;
  pages: DashboardPage[];
};

type DashboardPage = {
  id: string;
  name: string;
  columns: 4;
  rows: 2;
  buttons: DashboardButton[];
};

type DashboardButton = {
  id: string;
  actionId: string;
  position: number;
  labelOverride?: string;
  iconOverride?: string;
  color?: 'neutral' | 'blue' | 'green' | 'orange' | 'red' | 'purple';
  requireLongPress: boolean;
};
```

Persistir layout e preferências com Capacitor Preferences. Persistir token por meio de um plugin local que use Android Keystore; não salvar token em `localStorage`, arquivo JSON, log ou repositório.

### 10.4 Layout inicial

Página 1 — Aplicativos:

1. Chrome
2. WhatsApp
3. Finder
4. Terminal
5. Spotify
6. Visual Studio Code
7. Captura de tela
8. Modo trabalho

Página 2 — Mídia e sistema:

1. Música anterior
2. Play/Pause
3. Próxima música
4. Silenciar
5. Volume -
6. Volume +
7. Apagar monitor — toque prolongado
8. Bloquear Mac — toque prolongado

A instalação deverá tolerar aplicativos ausentes: a ação falha com mensagem clara, sem derrubar o agente.

---

## 11. Permissões

### 11.1 Android

- Acesso à internet/rede local.
- Vibração, se utilizada.
- Manter tela ativa.
- Armazenamento seguro via Android Keystore.

O APK poderá permitir tráfego HTTP claro apenas para a comunicação local do MVP. Isso deverá ser documentado em `docs/SECURITY.md`. Não usar essa exceção para comunicação com endereços públicos.

### 11.2 macOS

O instalador e `doctor` devem orientar o usuário sobre:

- Acessibilidade: necessária para teclas e alguns controles de sistema.
- Automação: necessária para Spotify, Apple Music e System Events.
- Rede local e firewall: necessária para o Android alcançar a porta `38121`.
- Gravação de tela: somente se uma função futura realmente capturar conteúdo.

Não solicitar Gravação de Tela no MVP apenas para abrir a ferramenta de captura.

---

## 12. Segurança obrigatória

1. Não aceitar comandos, scripts, caminhos ou URLs enviados pelo Android.
2. Não usar shell intermediário.
3. Validar todos os JSONs com limite de tamanho.
4. Limite sugerido por mensagem WebSocket: 16 KiB.
5. Limite sugerido por requisição HTTP: 64 KiB.
6. Usar comparação de token resistente a timing quando aplicável.
7. Armazenar somente hash do token no Mac.
8. Armazenar token no Android Keystore.
9. Pareamento fechado por padrão.
10. Aplicar rate limit por IP e por dispositivo.
11. Bloquear temporariamente tentativas repetidas de autenticação.
12. Não registrar tokens, códigos completos ou payloads sensíveis.
13. Validar `Origin` do WebSocket e aceitar somente as origens esperadas do Capacitor.
14. Rejeitar versões incompatíveis do protocolo.
15. Permitir revogação individual de dispositivos.
16. Não escutar em IPv6 público no MVP.
17. Não adicionar UPnP, port forwarding, Cloudflare Tunnel ou exposição pública.
18. Informar claramente no README que acesso remoto, se futuramente necessário, deve usar Tailscale.

O agente deve verificar se o endereço do cliente pertence a uma faixa privada ou link-local, salvo configuração futura explicitamente aprovada.

---

## 13. Logs e diagnóstico

Formato de log estruturado JSON Lines ou texto estruturado consistente.

Campos mínimos:

- timestamp UTC;
- nível;
- evento;
- request ID, quando houver;
- device ID abreviado ou hash;
- action ID;
- duração;
- resultado;
- código de erro.

Não registrar:

- access token;
- código de pareamento completo;
- conteúdo arbitrário do sistema;
- saída completa de processos;
- dados pessoais desnecessários.

Rotação:

- tamanho máximo aproximado de 5 MiB por arquivo;
- manter até três arquivos;
- falha de escrita do log não pode interromper o agente.

O comando `deskpanel-agent doctor` deve verificar:

- config válida;
- diretórios e permissões;
- porta livre ou ocupada pelo próprio agente;
- existência de `open`, `shortcuts`, `osascript` e `pmset`;
- status do LaunchAgent;
- conectividade local;
- ações inválidas ou duplicadas;
- versão do protocolo.

---

## 14. Instalação e execução

### 14.1 Desenvolvimento

O `Makefile` deverá expor, no mínimo:

```bash
make setup
make dev-agent
make dev-android
make test
make lint
make build-agent
make build-apk
make verify
```

`make verify` deverá executar formatação, lint, testes e builds sem instalar nada no sistema.

### 14.2 Instalação no Mac

O script `scripts/install-macos.sh` deve:

1. Verificar macOS e arquitetura.
2. Compilar ou receber o binário já compilado.
3. Criar os diretórios necessários.
4. Instalar o binário em `~/Library/Application Support/DeskPanel/bin/`.
5. Criar `config.json` a partir do exemplo somente quando não existir.
6. Criar o LaunchAgent sem inserir nome de usuário fixo.
7. Carregar o serviço com `launchctl` no domínio do usuário.
8. Exibir comandos de status, pareamento e diagnóstico.
9. Não sobrescrever configuração existente sem confirmação.

O script `uninstall-macos.sh` deve remover serviço e binário, mas preservar configuração e dispositivos por padrão. A remoção de dados deve exigir flag explícita, por exemplo `--purge`.

### 14.3 Instalação no Moto G60

O script `scripts/build-apk.sh` deve:

1. Instalar dependências apenas quando necessário.
2. Executar testes e build web.
3. Sincronizar o Capacitor.
4. Compilar APK de debug ou release conforme parâmetro.
5. Informar o caminho final do APK.

O README deve explicar instalação por ADB e por abertura manual do arquivo APK.

---

## 15. Testes

### 15.1 Go

Testes unitários obrigatórios:

- parsing e validação de configuração;
- IDs duplicados;
- tipos de ação desconhecidos;
- URLs inválidas;
- limites de volume;
- pareamento, expiração e uso único;
- hash e validação de token;
- dispositivo revogado;
- rate limit;
- protocolo incompatível;
- mensagem WebSocket inválida;
- ação inexistente;
- timeout;
- cancelamento de contexto;
- nenhuma chamada ao shell;
- persistência atômica.

Testes de integração com `FakeExecutor`:

- parear dispositivo;
- autenticar WebSocket;
- listar ações;
- executar ação;
- receber `started` e `result`;
- rejeitar ação inexistente;
- rejeitar token inválido;
- reconectar e obter estado.

Os testes automatizados não podem abrir aplicativos, bloquear o Mac, mudar volume ou controlar mídia real.

### 15.2 React/Android

Testes obrigatórios:

- tela de pareamento;
- validação de IP e porta;
- armazenamento de configuração sem token em texto simples;
- grade 4 × 2;
- troca de páginas;
- estados do botão;
- toque prolongado em ação perigosa;
- edição e persistência do layout;
- reconexão com backoff;
- tratamento dos códigos de erro;
- incompatibilidade de protocolo.

### 15.3 Teste E2E local

Usar o agente em modo de teste com `FakeExecutor`:

1. iniciar agente local;
2. parear cliente de teste;
3. abrir painel;
4. executar ações fictícias;
5. confirmar feedback visual;
6. derrubar conexão;
7. restaurar conexão;
8. confirmar reconexão e atualização do estado.

### 15.4 Teste físico no Moto G60

Checklist:

- APK instala e abre.
- Orientação permanece horizontal.
- Tela não apaga quando configurada para permanecer ativa.
- Pareamento funciona na mesma rede do Mac.
- Os oito botões da página inicial cabem sem rolagem.
- Toques não geram ações duplicadas.
- Reconexão ocorre após desligar e religar o Wi-Fi.
- Layout persiste após fechar e reabrir.
- Token permanece válido após reiniciar.
- Botões perigosos exigem toque prolongado.
- Interface permanece legível com brilho reduzido.
- Uso contínuo por duas horas não causa travamento ou aquecimento anormal.

---

## 16. Metas de desempenho

- Ação simples deve começar em até 500 ms na rede local em condições normais.
- Feedback de botão pressionado deve aparecer em menos de 100 ms.
- Aplicativo Android deve abrir em até 3 segundos no Moto G60.
- Reconexão deve começar em até 1 segundo após detectar queda.
- Agente ocioso deve consumir pouca CPU, com meta inferior a 1% em condições normais.
- Agente não deve crescer memória continuamente durante teste de duas horas.
- Ações idempotentes não devem ser executadas duas vezes pelo mesmo `requestId` dentro de uma janela curta.

O agente deve manter um cache limitado de `requestId` recentes para evitar execução duplicada durante reconexões.

---

## 17. Fases de implementação

### Fase 0 — Fundação

Entregas:

- estrutura do repositório;
- `README.md` inicial;
- `docs/DECISIONS.md` e `docs/STATUS.md`;
- Makefile;
- configuração de lint e formatação;
- esqueleto React/Capacitor;
- esqueleto Go;
- protocolo documentado;
- CI local executável por `make verify`.

Critério de saída:

- projetos compilam;
- testes mínimos executam;
- nenhuma funcionalidade real ainda necessária;
- working tree limpa após commit.

Commit sugerido:

```text
chore: bootstrap DeskPanel applications
```

### Fase 1 — Agente e ações locais

Entregas:

- carregamento e validação de `config.json`;
- catálogo de ações;
- `MacOSExecutor` e `FakeExecutor`;
- ações `open_app`, `open_url`, `run_shortcut`, volume e mídia;
- CLI `serve`, `status`, `doctor` e `version`;
- testes unitários sem ação real.

Critério de saída:

- todas as ações podem ser testadas pelo `FakeExecutor`;
- configuração insegura é rejeitada;
- nenhuma execução usa shell intermediário.

Commit sugerido:

```text
feat(agent): add validated action execution engine
```

### Fase 2 — API, WebSocket e segurança

Entregas:

- endpoints HTTP;
- WebSocket;
- protocolo v1;
- pareamento;
- tokens e revogação;
- rate limit;
- heartbeat;
- deduplicação por `requestId`;
- testes de integração.

Critério de saída:

- dispositivo não pareado não executa ação;
- token revogado deixa de funcionar imediatamente;
- testes cobrem fluxo completo usando `FakeExecutor`.

Commit sugerido:

```text
feat(agent): add secure local control protocol
```

### Fase 3 — Painel Android funcional

Entregas:

- tela de conexão e pareamento;
- armazenamento seguro do token;
- cliente HTTP e WebSocket;
- grade 4 × 2;
- duas páginas iniciais;
- estados visuais;
- reconexão;
- modo paisagem e tela ativa;
- testes de componentes.

Critério de saída:

- painel executa ações usando o agente de teste;
- APK de debug é gerado;
- nenhuma informação secreta aparece nos logs.

Commit sugerido:

```text
feat(android): add paired macro dashboard
```

### Fase 4 — Editor do painel

Entregas:

- criar e excluir páginas;
- adicionar, remover e reorganizar botões;
- escolher ação, nome, ícone e cor;
- toque prolongado;
- persistência e migração de schema;
- restauração do layout padrão.

Critério de saída:

- layout sobrevive a reinício;
- config corrompida usa backup ou padrão sem travar;
- editor nunca cria ação executável nova no Mac.

Commit sugerido:

```text
feat(android): add dashboard layout editor
```

### Fase 5 — Instalação real

Entregas:

- scripts de instalação e desinstalação do macOS;
- LaunchAgent;
- build APK;
- documentação de permissões;
- teste físico no Moto G60;
- teste real controlado no Mac.

Critério de saída:

- agente volta após logout/login;
- Moto G60 reconecta automaticamente;
- ações básicas funcionam no setup real;
- documentação permite reinstalação do zero.

Commit sugerido:

```text
build: add macOS and Android installation flows
```

### Fase 6 — Polimento do MVP

Entregas:

- correções do teste físico;
- feedback háptico;
- diagnóstico de conexão;
- rotação de logs;
- acessibilidade básica;
- ícones locais;
- revisão de segurança;
- versão `0.1.0`.

Critério de saída:

- todos os critérios da seção 18 aprovados;
- `make verify` limpo;
- repositório sem segredos e sem arquivos gerados indevidos;
- documentação atualizada;
- tag local `v0.1.0`, somente após autorização do proprietário.

Commit sugerido:

```text
release: prepare DeskPanel 0.1.0
```

---

## 18. Critérios de aceite do MVP

O MVP será considerado concluído somente quando:

- [ ] O APK estiver instalado no Moto G60.
- [ ] O aplicativo funcionar em modo paisagem.
- [ ] O Mac puder ser pareado por código temporário.
- [ ] O token estiver protegido pelo Android Keystore.
- [ ] O agente armazenar somente o hash do token.
- [ ] O agente iniciar automaticamente após login.
- [ ] O celular mostrar claramente online e offline.
- [ ] A reconexão funcionar sem refazer pareamento.
- [ ] Abrir Chrome, WhatsApp, Finder, Terminal, Spotify e VS Code funcionar quando instalados.
- [ ] Controle de volume funcionar.
- [ ] Controle do Spotify ou Apple Music funcionar conforme configurado.
- [ ] Atalhos do macOS poderem ser executados por nome autorizado.
- [ ] Bloqueio e apagamento do monitor exigirem toque prolongado.
- [ ] O usuário conseguir reorganizar e substituir botões.
- [ ] Layout persistir depois de reiniciar o aplicativo.
- [ ] Dispositivo revogado perder acesso.
- [ ] Action ID desconhecido ser rejeitado.
- [ ] Nenhum endpoint aceitar shell arbitrário.
- [ ] Nenhum token aparecer em logs.
- [ ] Testes automatizados passarem.
- [ ] Teste contínuo de duas horas não apresentar travamento.
- [ ] `make verify` terminar com código zero.
- [ ] `git diff --check` não apresentar erros.
- [ ] `docs/STATUS.md` refletir o estado real.

---

## 19. Diretrizes para a IA desenvolvedora

### 19.1 Antes de cada fase

1. Inspecionar arquivos e estado do Git.
2. Confirmar que não existem alterações do usuário que possam ser sobrescritas.
3. Criar ou atualizar uma tarefa em `docs/tasks/`.
4. Apresentar plano breve da fase.
5. Implementar somente o escopo da fase atual.

### 19.2 Durante a implementação

- Preferir código pequeno, explícito e testável.
- Não esconder erros com fallback silencioso.
- Não adicionar dependências sem necessidade demonstrável.
- Não alterar decisões deste documento apenas por conveniência.
- Não usar APIs privadas do macOS quando houver alternativa pública ou comando oficial.
- Não realizar chamadas a serviços de produção.
- Não adicionar analytics, anúncios ou SDKs externos.
- Não armazenar credenciais no Git.
- Não executar ações destrutivas reais durante testes.
- Não bloquear, suspender ou desligar o Mac automaticamente em testes.

### 19.3 Ao concluir cada fase

Executar, no mínimo:

```bash
make verify
git diff --check
git status --short
```

Atualizar:

- `docs/STATUS.md`;
- `docs/DECISIONS.md`, se houver nova decisão;
- tarefa correspondente em `docs/tasks/`;
- README, quando o uso mudar.

Relatar:

1. O que foi implementado.
2. Arquivos alterados.
3. Testes executados e seus resultados.
4. Riscos ou limitações restantes.
5. Próxima fase recomendada.

Não criar tag, publicar release, enviar ao GitHub, instalar no Mac real ou executar ação real sem autorização explícita do proprietário.

---

## 20. Primeira instrução de execução

Ao receber este arquivo pela primeira vez, a IA deve executar somente o seguinte:

1. Inspecionar o diretório atual e o estado do Git.
2. Caso o repositório ainda não exista, propor o diretório e a inicialização sem apagar conteúdo existente.
3. Comparar o ambiente disponível com os requisitos deste documento.
4. Apresentar o plano concreto da **Fase 0**.
5. Aguardar autorização antes de criar a estrutura do projeto.

Depois da autorização, implementar a Fase 0, validar e parar para revisão. Não avançar automaticamente para a Fase 1.

---

## 21. Evoluções futuras possíveis

Após o MVP estável, avaliar separadamente:

- mDNS/Bonjour com serviço `_deskpanel._tcp`;
- TLS local com estratégia confiável de certificados;
- aplicativo de menu em SwiftUI ou integração controlada com o agente Go;
- capa e metadados do Spotify;
- monitoramento de CPU, memória e bateria do Mac;
- detecção do aplicativo em primeiro plano;
- perfis automáticos;
- macros com múltiplas etapas e atrasos limitados;
- comandos para Home Assistant;
- widget de status do Proxmox, Uptime Kuma ou Grafana;
- acesso remoto via Tailscale;
- suporte a outros celulares Android;
- suporte opcional a Windows e Linux.

Cada evolução deverá ter ameaça, escopo, testes e critérios de aceite próprios. Nenhuma delas deve aumentar silenciosamente o poder de execução remota do MVP.

---

## 22. Resumo executivo

O DeskPanel será um painel Android dedicado, bonito e rápido, instalado no Moto G60 e conectado diretamente ao Mac pela rede local. O Android cuidará da experiência visual; o agente Go cuidará da autenticação e das ações do macOS. A divisão mantém o aplicativo simples, permite evolução futura e reduz o risco de execução remota indevida.

A primeira entrega prática será um painel 4 × 2 com duas páginas, pareamento seguro, reconexão automática e ações para aplicativos, volume, mídia, atalhos e bloqueio. A prioridade é confiabilidade no setup real, não distribuição pública.
