# Instalação do DeskPanel Android

Guia de build e instalação real (Fase 5 — `PROJECT.md` §14.3/§17).

## Gerar o APK

```bash
scripts/build-apk.sh              # debug
scripts/build-apk.sh --release    # release (precisa de assinatura configurada no Gradle para publicar)
```

O que o script faz, em ordem:

1. `pnpm install` (só se `node_modules` não existir; `--skip-install` pula
   sempre).
2. `pnpm test` (testes Vitest; `--skip-tests` pula).
3. `pnpm build` (build web de produção via Vite).
4. `npx cap sync android` (copia o build web para o projeto Android e
   sincroniza plugins nativos, incluindo o `SecureTokenStorage`).
5. `./gradlew assembleDebug` ou `assembleRelease`, conforme a flag.
6. Imprime o caminho final do `.apk` gerado.

### Flags

| Flag | Efeito |
|---|---|
| `--release` | compila `assembleRelease` em vez de `assembleDebug` |
| `--web-only` | roda só os passos 1–4 (sem Gradle) — útil em máquina sem Android SDK |
| `--skip-install` | pula `pnpm install` mesmo sem `node_modules` |
| `--skip-tests` | pula `pnpm test` |

Requer Android SDK + Gradle instalados (variável `ANDROID_HOME`/`ANDROID_SDK_ROOT`
configurada) para os passos 5–6. Sem isso, use `--web-only` para validar o
restante do pipeline e compile o APK numa máquina com o SDK.

## Instalação no Moto G60

### Via ADB (recomendado durante desenvolvimento)

1. Ative "Opções do desenvolvedor" → "Depuração USB" no Moto G60.
2. Conecte o celular ao Mac via USB e autorize a depuração quando solicitado.
3. Confirme que o `adb` enxerga o aparelho:

   ```bash
   adb devices
   ```

4. Instale o APK:

   ```bash
   adb install caminho/para/app-debug.apk
   ```

   Para reinstalar sobre uma versão já presente: `adb install -r caminho/para/app-debug.apk`.

### Manual (sem cabo/ADB)

1. Copie o `.apk` para o celular (e-mail, cabo, AirDrop-like via app de
   transferência, etc.).
2. No Moto G60, abra o arquivo `.apk` pelo gerenciador de arquivos.
3. Autorize "instalar apps de fontes desconhecidas" para o app usado na
   transferência, quando solicitado pelo Android.
4. Confirme a instalação.

## Após instalar

1. Abra o DeskPanel no celular.
2. Na tela de pareamento, informe o IP do Mac na rede local e a porta
   `38121`.
3. No Mac, rode `deskpanel-agent pair` para abrir a janela de pareamento
   (5 minutos, código de uso único) e digite o código no celular.
4. Confirme que as ações básicas do layout padrão funcionam (abrir app,
   volume, mídia).

## Desenvolvimento local (sem empacotar)

```bash
cd apps/android-panel
pnpm install
pnpm dev
```

Isso sobe o Vite dev server no navegador — útil para iterar na UI antes de
empacotar com Capacitor.
