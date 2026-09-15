# DeskPanel

Painel de atalhos tipo Stream Deck rodando em um Android dedicado (inicialmente um Motorola Moto G60), controlando um Mac pela rede local.

A especificação completa do produto — visão, arquitetura, protocolo, segurança, fases de implementação e critérios de aceite — vive em [`PROJECT.md`](./PROJECT.md). Este README é só o ponto de entrada prático; **qualquer decisão de arquitetura ou escopo segue o que está em `PROJECT.md`**, registrando mudanças em [`docs/DECISIONS.md`](./docs/DECISIONS.md).

## Componentes

| App | Linguagem | Caminho | Função |
|---|---|---|---|
| DeskPanel Android | React + TypeScript + Vite + Capacitor | `apps/android-panel` | Interface do painel no celular |
| DeskPanel Agent | Go | `apps/mac-agent` | Servidor local no Mac que valida e executa ações |

Comunicação: HTTP (pareamento, `/health`, `/version`) + WebSocket (execução de ações em tempo real) na porta `38121`, só em rede local. Nenhum comando arbitrário trafega do celular para o Mac — apenas um `actionId` pré-cadastrado. Detalhes do protocolo em [`docs/PROTOCOL.md`](./docs/PROTOCOL.md).

## Pré-requisitos de desenvolvimento

- Go >= 1.23 (`apps/mac-agent`)
- Node.js >= 20 e pnpm (`apps/android-panel`)
- Um Mac para rodar o agente de verdade (Accessibility/Automation permissions — ver [`docs/SECURITY.md`](./docs/SECURITY.md))
- Um dispositivo Android (ou emulador) para o app — ver [`docs/INSTALL-ANDROID.md`](./docs/INSTALL-ANDROID.md)

## Comandos

```bash
make setup        # instala dependências dos dois apps
make dev-agent     # roda o agente Go localmente
make dev-android    # roda o app Android em modo dev (Vite)
make test          # testes Go + testes React
make lint           # lint Go + lint TS
make build-agent    # compila o binário do agente
make build-apk       # gera o APK do painel Android
make verify           # formatação + lint + testes + builds, sem instalar nada no sistema
```

## Status

Veja [`docs/STATUS.md`](./docs/STATUS.md) para o estado atual do projeto por fase.

## Segurança

Modelo de segurança resumido em [`docs/SECURITY.md`](./docs/SECURITY.md): pareamento por código temporário, token armazenado como hash no Mac e no Android Keystore no celular, nenhum shell arbitrário, sem exposição à internet pública no MVP. Acesso remoto futuro, se necessário, será só via Tailscale — nunca porta pública.

## Licença

Projeto pessoal de Ricardo Pena. Sem licença de distribuição definida ainda.
