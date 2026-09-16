#!/usr/bin/env bash
set -euo pipefail

# Build do APK do DeskPanel Android Panel (PROJECT.md §14.3).
#
# Uso:
#   scripts/build-apk.sh [--release] [--web-only] [--skip-install] [--skip-tests]
#
# Opções:
#   --release       compila APK de release (assembleRelease). Padrão: debug.
#   --web-only      roda só install/test/build web + cap sync, sem invocar o
#                   Gradle (útil em ambientes sem Android SDK instalado).
#   --skip-install  pula 'pnpm install' mesmo se node_modules não existir.
#   --skip-tests    pula 'pnpm test' (não recomendado fora de CI local).
#   -h, --help      mostra esta ajuda.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PANEL_DIR="$REPO_ROOT/apps/android-panel"

VARIANT="debug"
WEB_ONLY=0
SKIP_INSTALL=0
SKIP_TESTS=0

print_help() {
	sed -n '3,15p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
}

while [ $# -gt 0 ]; do
	case "$1" in
	--release)
		VARIANT="release"
		shift
		;;
	--web-only)
		WEB_ONLY=1
		shift
		;;
	--skip-install)
		SKIP_INSTALL=1
		shift
		;;
	--skip-tests)
		SKIP_TESTS=1
		shift
		;;
	-h | --help)
		print_help
		exit 0
		;;
	*)
		echo "erro: opção desconhecida: $1" >&2
		print_help
		exit 1
		;;
	esac
done

if ! command -v pnpm >/dev/null 2>&1; then
	echo "erro: 'pnpm' não encontrado no PATH. Instale com 'corepack enable' (Node 16.9+)." >&2
	exit 1
fi

cd "$PANEL_DIR"

# 1. Instalar dependências apenas quando necessário --------------------------
if [ "$SKIP_INSTALL" -eq 1 ]; then
	echo "pulando 'pnpm install' (--skip-install)"
elif [ -d node_modules ]; then
	echo "node_modules já existe — pulando 'pnpm install' (use sem --skip-install e apague node_modules para forçar)"
else
	echo "instalando dependências..."
	pnpm install
fi

# 2. Testes e build web -------------------------------------------------------
if [ "$SKIP_TESTS" -eq 1 ]; then
	echo "pulando 'pnpm test' (--skip-tests)"
else
	echo "rodando testes..."
	pnpm test
fi

echo "compilando build web..."
pnpm build

# 3. Sincronizar o Capacitor ---------------------------------------------------
echo "sincronizando Capacitor (cap sync android)..."
pnpm exec cap sync android

if [ "$WEB_ONLY" -eq 1 ]; then
	echo
	echo "--web-only informado: build web + cap sync concluídos, Gradle não foi executado."
	echo "Para gerar o APK, rode este script sem --web-only em uma máquina com Android SDK/Gradle."
	exit 0
fi

# 4. Compilar o APK (debug ou release) ----------------------------------------
ANDROID_DIR="$PANEL_DIR/android"
if [ ! -x "$ANDROID_DIR/gradlew" ]; then
	echo "erro: $ANDROID_DIR/gradlew não encontrado ou não executável." >&2
	echo "Instale o Android SDK/Gradle ou rode com --web-only para pular esta etapa." >&2
	exit 1
fi

if [ "$VARIANT" = "release" ]; then
	GRADLE_TASK="assembleRelease"
else
	GRADLE_TASK="assembleDebug"
fi

echo "compilando APK ($GRADLE_TASK)..."
(cd "$ANDROID_DIR" && ./gradlew "$GRADLE_TASK")

# 5. Informar o caminho final do APK ------------------------------------------
APK_PATH=$(find "$ANDROID_DIR/app/build/outputs/apk/$VARIANT" -name "*.apk" -print -quit 2>/dev/null || true)
if [ -z "$APK_PATH" ]; then
	echo "erro: build do Gradle terminou mas nenhum .apk foi encontrado em app/build/outputs/apk/$VARIANT" >&2
	exit 1
fi

echo
echo "APK gerado: $APK_PATH"
echo
echo "Instalar via ADB (celular em modo depuração USB):"
echo "  adb install \"$APK_PATH\""
echo
echo "Ou copie o arquivo para o celular e abra-o (permitir instalação de fontes desconhecidas)."
