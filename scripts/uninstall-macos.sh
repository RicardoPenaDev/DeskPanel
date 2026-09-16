#!/usr/bin/env bash
set -euo pipefail

# Desinstalador do DeskPanel Agent no macOS (PROJECT.md §14.2).
#
# Por padrão remove apenas o serviço (LaunchAgent) e o binário instalado,
# preservando config.json, devices.json, state.json e logs.
#
# Uso:
#   scripts/uninstall-macos.sh [opções]
#
# Opções:
#   --purge     também remove config.json, devices.json, state.json, o
#               socket administrativo e os logs. Requer confirmação
#               interativa, a menos que --yes também seja informado.
#   --yes, -y   não pedir confirmação (usar com --purge em automação).
#   -h, --help  mostra esta ajuda.

APP_NAME="DeskPanel"
LAUNCH_LABEL="dev.ricardopena.deskpanel.agent"

PURGE=0
ASSUME_YES=0

print_help() {
	sed -n '3,17p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
}

while [ $# -gt 0 ]; do
	case "$1" in
	--purge)
		PURGE=1
		shift
		;;
	--yes | -y)
		ASSUME_YES=1
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

if [ "$(uname -s)" != "Darwin" ]; then
	echo "erro: este desinstalador só roda no macOS (uname -s retornou $(uname -s))" >&2
	exit 1
fi

APP_SUPPORT_DIR="$HOME/Library/Application Support/$APP_NAME"
LOG_DIR="$HOME/Library/Logs/$APP_NAME"
PLIST_PATH="$HOME/Library/LaunchAgents/$LAUNCH_LABEL.plist"
UID_NUM="$(id -u)"

# 1. Descarregar e remover o LaunchAgent -------------------------------------
if launchctl bootout "gui/$UID_NUM/$LAUNCH_LABEL" >/dev/null 2>&1; then
	echo "serviço descarregado (launchctl bootout)"
else
	echo "serviço não estava carregado (ou já foi descarregado)"
fi

if [ -f "$PLIST_PATH" ]; then
	rm -f "$PLIST_PATH"
	echo "LaunchAgent removido: $PLIST_PATH"
else
	echo "LaunchAgent não encontrado em $PLIST_PATH (nada a remover)"
fi

# 2. Remover o binário, preservando dados por padrão -------------------------
BIN_PATH="$APP_SUPPORT_DIR/bin/deskpanel-agent"
if [ -f "$BIN_PATH" ]; then
	rm -f "$BIN_PATH"
	echo "binário removido: $BIN_PATH"
fi
rmdir "$APP_SUPPORT_DIR/bin" 2>/dev/null || true

if [ "$PURGE" -ne 1 ]; then
	echo
	echo "Desinstalação concluída. Preservados (use --purge para remover):"
	echo "  $APP_SUPPORT_DIR/config.json"
	echo "  $APP_SUPPORT_DIR/devices.json"
	echo "  $APP_SUPPORT_DIR/state.json"
	echo "  $APP_SUPPORT_DIR/agent.sock (se ainda existir)"
	echo "  $LOG_DIR"
	exit 0
fi

# 3. --purge: remover dados, com confirmação explícita ------------------------
echo
echo "aviso: --purge vai remover permanentemente:"
echo "  $APP_SUPPORT_DIR (config.json, devices.json, state.json, agent.sock)"
echo "  $LOG_DIR"

if [ "$ASSUME_YES" -ne 1 ]; then
	read -r -p "Confirma a remoção permanente desses dados? [y/N] " reply
	case "$reply" in
	[yY] | [yY][eE][sS]) ;;
	*)
		echo "purge cancelado — dados preservados."
		exit 0
		;;
	esac
fi

rm -rf "$APP_SUPPORT_DIR"
rm -rf "$LOG_DIR"
echo "dados removidos: $APP_SUPPORT_DIR e $LOG_DIR"
