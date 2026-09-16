.PHONY: setup dev-agent dev-android test test-agent test-android \
        lint lint-agent lint-android build-agent build-apk verify

AGENT_DIR := apps/mac-agent
PANEL_DIR := apps/android-panel
VERSION := 0.1.0

## setup: instala as dependências dos dois apps.
setup:
	cd $(AGENT_DIR) && go mod download
	cd $(PANEL_DIR) && pnpm install

## dev-agent: roda o agente Go localmente (sem instalar LaunchAgent).
dev-agent:
	cd $(AGENT_DIR) && go run ./cmd/deskpanel-agent serve

## dev-android: sobe o Vite dev server do painel.
dev-android:
	cd $(PANEL_DIR) && pnpm dev

## test: testes Go + testes React.
test: test-agent test-android

test-agent:
	cd $(AGENT_DIR) && go test ./...

test-android:
	cd $(PANEL_DIR) && pnpm test

## lint: lint/format Go + lint/format TS.
lint: lint-agent lint-android

lint-agent:
	cd $(AGENT_DIR) && gofmt -l . | (! grep .) && go vet ./...

lint-android:
	cd $(PANEL_DIR) && pnpm lint && pnpm format

## build-agent: compila o binário do agente.
build-agent:
	cd $(AGENT_DIR) && go build -ldflags "-X main.version=$(VERSION)" -o bin/deskpanel-agent ./cmd/deskpanel-agent

## build-apk: gera o APK do painel Android (Fase 5).
build-apk:
	./scripts/build-apk.sh

## verify: formatação + lint + testes + builds, sem instalar nada no sistema.
verify: lint test build-agent
	cd $(PANEL_DIR) && pnpm build
