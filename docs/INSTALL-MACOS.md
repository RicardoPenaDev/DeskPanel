# Instalação do DeskPanel Agent no macOS

> Status: **ainda não implementado** (previsto para a Fase 5 — Instalação real, `PROJECT.md` §17). Este documento descreve o fluxo alvo; será atualizado com passos reais quando `scripts/install-macos.sh` existir.

## Alvo (Fase 5)

`scripts/install-macos.sh` deverá:

1. Verificar macOS e arquitetura.
2. Compilar ou receber o binário já compilado (`apps/mac-agent`).
3. Criar `~/Library/Application Support/DeskPanel/` (`0700`) e `~/Library/Logs/DeskPanel/`.
4. Instalar o binário em `~/Library/Application Support/DeskPanel/bin/`.
5. Criar `config.json` a partir de `configs/config.example.json` só se ainda não existir.
6. Criar o LaunchAgent (`~/Library/LaunchAgents/dev.ricardopena.deskpanel.agent.plist`) sem usuário fixo hardcoded.
7. Carregar o serviço com `launchctl` no domínio do usuário.
8. Mostrar os próximos comandos (`status`, `pair`, `doctor`).
9. Nunca sobrescrever configuração existente sem confirmação.

`scripts/uninstall-macos.sh` remove serviço e binário, preserva `config.json`/`devices.json` por padrão; apagar dados exige `--purge` explícito.

## Permissões necessárias

Ver `docs/SECURITY.md` — Acessibilidade, Automação e Rede local/Firewall. `deskpanel-agent doctor` deverá diagnosticar o que falta.

## Desenvolvimento local (antes da Fase 5)

Enquanto o instalador não existe, para rodar o agente manualmente durante o desenvolvimento:

```bash
cd apps/mac-agent
go run ./cmd/deskpanel-agent serve
```

Isso não instala LaunchAgent nem persiste nada fora do diretório de trabalho — é só para iterar localmente.
