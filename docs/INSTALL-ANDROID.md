# Instalação do DeskPanel Android

> Status: **ainda não implementado** (previsto para a Fase 5 — Instalação real, `PROJECT.md` §17). Este documento descreve o fluxo alvo; será atualizado com passos reais quando `scripts/build-apk.sh` existir.

## Alvo (Fase 5)

`scripts/build-apk.sh` deverá:

1. Instalar dependências (`pnpm install`) só quando necessário.
2. Rodar testes e build web (`pnpm build`).
3. Sincronizar o Capacitor (`npx cap sync android`).
4. Compilar o APK (debug ou release, conforme parâmetro).
5. Informar o caminho final do `.apk` gerado.

## Instalação no Moto G60

Duas formas, a documentar em detalhe na Fase 5:

- **Via ADB**: `adb install caminho/para/app-debug.apk` com o celular em modo depuração USB.
- **Manual**: copiar o `.apk` para o celular e abrir o arquivo (exige permitir instalação de fontes desconhecidas).

## Desenvolvimento local (antes da Fase 5)

```bash
cd apps/android-panel
pnpm install
pnpm dev
```

Isso sobe o Vite dev server no navegador — útil para iterar na UI antes de empacotar com Capacitor.
