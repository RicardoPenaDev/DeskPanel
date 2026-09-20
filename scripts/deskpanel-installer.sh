#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_ROOT="$APP_ROOT/Resources"
INSTALLER="$PACKAGE_ROOT/install.sh"
APPLICATIONS_APP="$PACKAGE_ROOT/DeskPanel.app"

dialog() {
  /usr/bin/osascript -e 'on run argv' -e 'display dialog (item 1 of argv) with title "DeskPanel" buttons {"OK"} default button "OK"' -e 'end run' "$1" >/dev/null
}

if ! DESKPANEL_PACKAGE_ROOT="$PACKAGE_ROOT" "$INSTALLER"; then
  dialog "Não foi possível instalar o DeskPanel. Verifique os registros em ~/Library/Logs/DeskPanel."
  exit 1
fi

if ! /usr/bin/ditto "$APPLICATIONS_APP" "/Applications/DeskPanel.app"; then
  dialog "O agente foi instalado, mas não foi possível copiar o DeskPanel para /Applications."
  exit 1
fi

dialog "DeskPanel instalado com sucesso. A página de configuração será aberta agora."
mkdir -p "$HOME/Library/Logs/DeskPanel"
nohup "$HOME/Library/Application Support/DeskPanel/bin/deskpanel-agent" setup \
  >"$HOME/Library/Logs/DeskPanel/setup.log" 2>&1 </dev/null &
