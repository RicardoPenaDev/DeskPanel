#!/usr/bin/env bash
set -euo pipefail

AGENT="$HOME/Library/Application Support/DeskPanel/bin/deskpanel-agent"
LOG_DIR="$HOME/Library/Logs/DeskPanel"
mkdir -p "$LOG_DIR"

if [ ! -x "$AGENT" ]; then
  /usr/bin/osascript -e 'display dialog "O DeskPanel ainda não está instalado. Abra o instalador do DMG primeiro." with title "DeskPanel" buttons {"OK"} default button "OK"' >/dev/null
  exit 1
fi

SETUP_LOG="$LOG_DIR/setup.log"
: >"$SETUP_LOG"
nohup "$AGENT" setup >"$SETUP_LOG" 2>&1 </dev/null &

for _ in 1 2 3 4 5 6 7 8 9 10; do
  SETUP_URL="$(sed -n 's/^Assistente local aberto em //p' "$SETUP_LOG" | tail -1)"
  if [ -n "$SETUP_URL" ]; then
    /usr/bin/open "$SETUP_URL" >/dev/null 2>&1 || true
    break
  fi
  sleep 0.2
done
