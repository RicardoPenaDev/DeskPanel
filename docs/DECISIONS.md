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
