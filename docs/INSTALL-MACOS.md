# Instalação do DeskPanel Agent no macOS

Guia de instalação real (Fase 5 — `PROJECT.md` §14.2/§17). Os scripts abaixo
não aceitam comando, caminho ou URL arbitrários vindos de fora — só as flags
documentadas.

## Instalar

```bash
scripts/install-macos.sh
```

O que o script faz, em ordem:

1. Verifica que o sistema é macOS (`uname -s == Darwin`); aborta em qualquer
   outro sistema.
2. Compila `deskpanel-agent` com `go build` (ou usa `--binary <caminho>` para
   um binário já compilado, ex. cross-compilado em CI).
3. Cria `~/Library/Application Support/DeskPanel/` (`0700`) e
   `~/Library/Logs/DeskPanel/`.
4. Instala o binário em `~/Library/Application Support/DeskPanel/bin/deskpanel-agent`
   (`0755`).
5. Copia `configs/config.example.json` para `config.json` **só se ainda não
   existir** (`0600`). Use `--force` para sobrescrever de propósito.
6. Gera o LaunchAgent em
   `~/Library/LaunchAgents/dev.ricardopena.deskpanel.agent.plist`, sem nenhum
   nome de usuário fixo (usa o binário instalado e `$HOME` resolvidos em
   tempo de instalação).
7. Carrega o serviço com `launchctl bootstrap gui/$(id -u)` (com `bootout`
   antes, para reinstalação idempotente).
8. Imprime os próximos comandos (`status`, `pair`, `doctor`).

### Flags

| Flag | Efeito |
|---|---|
| `--binary <caminho>` | usa um binário já compilado em vez de `go build` |
| `--force` | sobrescreve `config.json` existente |
| `--skip-launchagent` | instala arquivos mas não cria/carrega o LaunchAgent |
| `--no-load` | cria o `.plist` mas não chama `launchctl bootstrap` |

## Verificar

```bash
~/Library/Application\ Support/DeskPanel/bin/deskpanel-agent status
~/Library/Application\ Support/DeskPanel/bin/deskpanel-agent doctor
launchctl print gui/$(id -u)/dev.ricardopena.deskpanel.agent
```

`doctor` verifica config, devices.json, permissões de Acessibilidade/Automação,
porta e comandos do sistema — ver `docs/SECURITY.md`.

Para confirmar que o agente volta sozinho após logout/login (critério de
saída da Fase 5): faça logout e login de novo, depois rode `status` — o
LaunchAgent tem `RunAtLoad=true` e `KeepAlive.SuccessfulExit=false`, então o
launchd sobe o processo automaticamente e o reinicia se ele cair.

## Desinstalar

```bash
scripts/uninstall-macos.sh            # remove serviço + binário, preserva dados
scripts/uninstall-macos.sh --purge    # também remove config.json/devices.json/state.json/logs (pede confirmação)
scripts/uninstall-macos.sh --purge --yes   # purge sem prompt (uso em automação)
```

## Permissões necessárias

Ver `docs/SECURITY.md` — Acessibilidade, Automação e Rede local/Firewall.
`deskpanel-agent doctor` diagnostica o que falta.

## Reinstalação do zero

```bash
scripts/uninstall-macos.sh --purge --yes
scripts/install-macos.sh
```

## Desenvolvimento local (sem instalar)

```bash
cd apps/mac-agent
go run ./cmd/deskpanel-agent serve
```

Isso não instala LaunchAgent nem persiste nada fora do diretório de
trabalho — é só para iterar localmente.
