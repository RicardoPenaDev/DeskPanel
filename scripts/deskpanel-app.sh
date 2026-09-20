#!/usr/bin/env bash
set -euo pipefail

AGENT="$HOME/Library/Application Support/DeskPanel/bin/deskpanel-agent"
LOG_DIR="$HOME/Library/Logs/DeskPanel"
mkdir -p "$LOG_DIR"

if [ ! -x "$AGENT" ]; then
  /usr/bin/osascript -e 'display dialog "O DeskPanel ainda não está instalado. Abra o instalador do DMG primeiro." with title "DeskPanel" buttons {"OK"} default button "OK"' >/dev/null
  exit 1
fi

nohup "$AGENT" setup >"$LOG_DIR/setup.log" 2>&1 </dev/null &
