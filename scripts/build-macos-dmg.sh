#!/usr/bin/env bash
set -euo pipefail

[ "$(uname -s)" = "Darwin" ] || { echo "erro: este script só gera DMG no macOS" >&2; exit 1; }
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AGENT_DIR="$REPO_ROOT/apps/mac-agent"
VERSION="0.1.0"
OUTPUT_DIR="$REPO_ROOT/dist/macos"
OUTPUT_DMG="$OUTPUT_DIR/DeskPanel-$VERSION.dmg"

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="${2:-}"; shift 2; OUTPUT_DMG="$OUTPUT_DIR/DeskPanel-$VERSION.dmg";;
    --output) OUTPUT_DMG="${2:-}"; shift 2;;
    -h|--help) echo "Uso: scripts/build-macos-dmg.sh [--version VERSION] [--output CAMINHO]"; exit 0;;
    *) echo "erro: opção desconhecida: $1" >&2; exit 1;;
  esac
done

STAGING="$(mktemp -d "${TMPDIR:-/tmp}/deskpanel-dmg.XXXXXX")"
trap 'rm -rf "$STAGING"' EXIT
PACKAGE="$STAGING/DeskPanel"
mkdir -p "$PACKAGE/bin"
(cd "$AGENT_DIR" && go build -ldflags "-X main.version=$VERSION" -o "$PACKAGE/bin/deskpanel-agent" ./cmd/deskpanel-agent)
cp "$REPO_ROOT/configs/config.example.json" "$PACKAGE/config.example.json"
cp "$REPO_ROOT/scripts/install-packaged-macos.sh" "$PACKAGE/install.sh"
chmod 755 "$PACKAGE/install.sh"
cat >"$PACKAGE/LEIA-ME.txt" <<EOF
DeskPanel $VERSION

Abra o aplicativo "DeskPanel Installer" para instalar com duplo clique.
Configurações e pareamentos existentes são preservados.
EOF

INSTALLER_APP="$STAGING/DeskPanel Installer.app"
mkdir -p "$INSTALLER_APP/Contents/MacOS" "$INSTALLER_APP/Contents/Resources"
cp -R "$PACKAGE/bin" "$PACKAGE/config.example.json" "$PACKAGE/install.sh" "$INSTALLER_APP/Contents/Resources/"
cp "$REPO_ROOT/scripts/deskpanel-installer.sh" "$INSTALLER_APP/Contents/MacOS/DeskPanel Installer"
cp "$REPO_ROOT/assets/macos/DeskPanel.icns" "$INSTALLER_APP/Contents/Resources/DeskPanel.icns"
chmod 755 "$INSTALLER_APP/Contents/MacOS/DeskPanel Installer"

DESKPANEL_APP="$INSTALLER_APP/Contents/Resources/DeskPanel.app"
mkdir -p "$DESKPANEL_APP/Contents/MacOS" "$DESKPANEL_APP/Contents/Resources"
(
  cd "$AGENT_DIR" &&
  go build -ldflags "-X main.version=$VERSION" -o "$DESKPANEL_APP/Contents/MacOS/DeskPanel" ./cmd/deskpanel-launcher
)
cp "$REPO_ROOT/assets/macos/DeskPanel.icns" "$DESKPANEL_APP/Contents/Resources/DeskPanel.icns"
chmod 755 "$DESKPANEL_APP/Contents/MacOS/DeskPanel"
cat >"$DESKPANEL_APP/Contents/Info.plist" <<PLIST_EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleDisplayName</key><string>DeskPanel</string>
<key>CFBundleExecutable</key><string>DeskPanel</string>
<key>CFBundleIconFile</key><string>DeskPanel.icns</string>
<key>CFBundleIdentifier</key><string>dev.ricardopena.deskpanel</string>
<key>CFBundleName</key><string>DeskPanel</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>$VERSION</string>
<key>CFBundleVersion</key><string>$VERSION</string>
<key>LSUIElement</key><true/>
</dict></plist>
PLIST_EOF
cat >"$INSTALLER_APP/Contents/Info.plist" <<PLIST_EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleDisplayName</key><string>DeskPanel Installer</string>
<key>CFBundleExecutable</key><string>DeskPanel Installer</string>
<key>CFBundleIdentifier</key><string>dev.ricardopena.deskpanel.installer</string>
<key>CFBundleIconFile</key><string>DeskPanel.icns</string>
<key>CFBundleName</key><string>DeskPanel Installer</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>$VERSION</string>
<key>CFBundleVersion</key><string>$VERSION</string>
</dict></plist>
PLIST_EOF
/usr/bin/codesign --force --deep --sign - "$INSTALLER_APP" >/dev/null
mkdir -p "$(dirname "$OUTPUT_DMG")"
rm -f "$OUTPUT_DMG"
hdiutil create -volname "DeskPanel $VERSION" -srcfolder "$STAGING" -ov -format UDZO "$OUTPUT_DMG" >/dev/null
echo "DMG criado: $OUTPUT_DMG"
