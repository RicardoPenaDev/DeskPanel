#!/usr/bin/env bash
set -euo pipefail

PACKAGE_ROOT="${DESKPANEL_PACKAGE_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
APP_NAME="DeskPanel"
LAUNCH_LABEL="dev.ricardopena.deskpanel.agent"
CONFIG_DIR="$HOME/Library/Application Support/$APP_NAME"
INSTALLED_BINARY="$CONFIG_DIR/bin/deskpanel-agent"
CONFIG_PATH="$CONFIG_DIR/config.json"
CONFIG_EXAMPLE="$PACKAGE_ROOT/config.example.json"
LOG_DIR="$HOME/Library/Logs/$APP_NAME"
PLIST_PATH="$HOME/Library/LaunchAgents/$LAUNCH_LABEL.plist"

[ "$(uname -s)" = "Darwin" ] || { echo "erro: este instalador só funciona no macOS" >&2; exit 1; }
[ -f "$PACKAGE_ROOT/bin/deskpanel-agent" ] && [ -f "$CONFIG_EXAMPLE" ] || { echo "erro: pacote DeskPanel incompleto" >&2; exit 1; }

mkdir -p "$CONFIG_DIR/bin" "$LOG_DIR" "$HOME/Library/LaunchAgents"
chmod 700 "$CONFIG_DIR"
cp "$PACKAGE_ROOT/bin/deskpanel-agent" "$INSTALLED_BINARY"
chmod 755 "$INSTALLED_BINARY"
if [ ! -f "$CONFIG_PATH" ]; then
  cp "$CONFIG_EXAMPLE" "$CONFIG_PATH"
  chmod 600 "$CONFIG_PATH"
fi

cat >"$PLIST_PATH" <<PLIST_EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>$LAUNCH_LABEL</string>
<key>ProgramArguments</key><array><string>$INSTALLED_BINARY</string><string>serve</string></array>
<key>RunAtLoad</key><true/>
<key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict>
<key>StandardOutPath</key><string>$LOG_DIR/agent.log</string>
<key>StandardErrorPath</key><string>$LOG_DIR/agent.err.log</string>
<key>ProcessType</key><string>Background</string>
</dict></plist>
PLIST_EOF

UID_NUM="$(id -u)"
launchctl bootout "gui/$UID_NUM/$LAUNCH_LABEL" >/dev/null 2>&1 || true
launchctl bootstrap "gui/$UID_NUM" "$PLIST_PATH"
echo "DeskPanel instalado. Próximo passo: conceda Acessibilidade/Automação ao agente."
