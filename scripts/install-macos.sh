#!/usr/bin/env bash
set -euo pipefail

# Instalador do DeskPanel Agent no macOS (PROJECT.md §14.2).
#
# Uso:
#   scripts/install-macos.sh [opções]
#
# Opções:
#   --binary <caminho>   usa um binário já compilado em vez de compilar com go build.
#   --force               sobrescreve config.json existente (copia de novo o exemplo).
#   --skip-launchagent    instala arquivos mas não cria/carrega o LaunchAgent.
#   --no-load             cria o LaunchAgent mas não chama launchctl bootstrap.
#   -h, --help            mostra esta ajuda.
#
# O script nunca aceita comando, caminho ou URL vindos de fora que não sejam
# os parâmetros acima — nada aqui é repassado para um shell intermediário.

APP_NAME="DeskPanel"
LAUNCH_LABEL="dev.ricardopena.deskpanel.agent"
AGENT_VERSION="0.1.0"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AGENT_DIR="$REPO_ROOT/apps/mac-agent"

BINARY_PATH=""
FORCE_CONFIG=0
SKIP_LAUNCHAGENT=0
NO_LOAD=0

print_help() {
	sed -n '3,17p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
}

while [ $# -gt 0 ]; do
	case "$1" in
	--binary)
		BINARY_PATH="${2:-}"
		if [ -z "$BINARY_PATH" ]; then
			echo "erro: --binary exige um caminho" >&2
			exit 1
		fi
		shift 2
		;;
	--force)
		FORCE_CONFIG=1
		shift
		;;
	--skip-launchagent)
		SKIP_LAUNCHAGENT=1
		shift
		;;
	--no-load)
		NO_LOAD=1
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

# 1. Verificar macOS e arquitetura -------------------------------------------
if [ "$(uname -s)" != "Darwin" ]; then
	echo "erro: este instalador só roda no macOS (uname -s retornou $(uname -s))" >&2
	exit 1
fi
ARCH="$(uname -m)"
echo "macOS detectado (arquitetura: $ARCH)"

# 2. Compilar ou receber o binário já compilado ------------------------------
BUILT_BINARY="$AGENT_DIR/bin/deskpanel-agent"
if [ -n "$BINARY_PATH" ]; then
	if [ ! -f "$BINARY_PATH" ]; then
		echo "erro: binário não encontrado em $BINARY_PATH" >&2
		exit 1
	fi
	echo "usando binário fornecido: $BINARY_PATH"
else
	if ! command -v go >/dev/null 2>&1; then
		echo "erro: 'go' não encontrado no PATH. Instale Go 1.23+ ou use --binary <caminho>." >&2
		exit 1
	fi
	echo "compilando deskpanel-agent com 'go build' (versão $AGENT_VERSION)..."
	(cd "$AGENT_DIR" && go build -ldflags "-X main.version=$AGENT_VERSION" -o bin/deskpanel-agent ./cmd/deskpanel-agent)
	BINARY_PATH="$BUILT_BINARY"
fi

# 3. Criar os diretórios necessários -----------------------------------------
APP_SUPPORT_DIR="$HOME/Library/Application Support/$APP_NAME"
BIN_DIR="$APP_SUPPORT_DIR/bin"
LOG_DIR="$HOME/Library/Logs/$APP_NAME"
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"

mkdir -p "$BIN_DIR"
chmod 700 "$APP_SUPPORT_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$LAUNCH_AGENTS_DIR"
echo "diretórios prontos em: $APP_SUPPORT_DIR"

# 4. Instalar o binário -------------------------------------------------------
INSTALLED_BINARY="$BIN_DIR/deskpanel-agent"
cp "$BINARY_PATH" "$INSTALLED_BINARY"
chmod 755 "$INSTALLED_BINARY"
echo "binário instalado em: $INSTALLED_BINARY"

# 5. Criar config.json a partir do exemplo, só se ainda não existir ---------
CONFIG_PATH="$APP_SUPPORT_DIR/config.json"
CONFIG_EXAMPLE="$REPO_ROOT/configs/config.example.json"
if [ -f "$CONFIG_PATH" ] && [ "$FORCE_CONFIG" -ne 1 ]; then
	echo "config.json já existe em $CONFIG_PATH — mantendo (use --force para sobrescrever)"
else
	if [ -f "$CONFIG_PATH" ] && [ "$FORCE_CONFIG" -eq 1 ]; then
		echo "aviso: --force informado, sobrescrevendo config.json existente"
	fi
	cp "$CONFIG_EXAMPLE" "$CONFIG_PATH"
	chmod 600 "$CONFIG_PATH"
	echo "config.json criado a partir de configs/config.example.json"
fi

if [ "$SKIP_LAUNCHAGENT" -eq 1 ]; then
	echo "arquivos instalados. LaunchAgent não foi criado (--skip-launchagent)."
	exit 0
fi

# 6. Criar o LaunchAgent, sem usuário fixo hardcoded -------------------------
PLIST_PATH="$LAUNCH_AGENTS_DIR/$LAUNCH_LABEL.plist"
STDOUT_LOG="$LOG_DIR/agent.log"
STDERR_LOG="$LOG_DIR/agent.err.log"

cat >"$PLIST_PATH" <<PLIST_EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$LAUNCH_LABEL</string>
	<key>ProgramArguments</key>
	<array>
		<string>$INSTALLED_BINARY</string>
		<string>serve</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>StandardOutPath</key>
	<string>$STDOUT_LOG</string>
	<key>StandardErrorPath</key>
	<string>$STDERR_LOG</string>
	<key>ProcessType</key>
	<string>Background</string>
</dict>
</plist>
PLIST_EOF
chmod 644 "$PLIST_PATH"
echo "LaunchAgent criado em: $PLIST_PATH"

if [ "$NO_LOAD" -eq 1 ]; then
	echo "LaunchAgent criado mas não carregado (--no-load). Carregue manualmente com:"
	echo "  launchctl bootstrap gui/\$(id -u) \"$PLIST_PATH\""
	exit 0
fi

# 7. Carregar o serviço com launchctl no domínio do usuário ------------------
UID_NUM="$(id -u)"
# bootout é idempotente: se o serviço não estiver carregado, apenas falha
# silenciosamente (ignoramos o código de saída aqui de propósito).
launchctl bootout "gui/$UID_NUM/$LAUNCH_LABEL" >/dev/null 2>&1 || true
if launchctl bootstrap "gui/$UID_NUM" "$PLIST_PATH"; then
	echo "serviço carregado via launchctl (gui/$UID_NUM)"
else
	echo "erro: falha ao carregar o LaunchAgent via launchctl bootstrap" >&2
	exit 1
fi

# 8. Exibir comandos de status, pareamento e diagnóstico ---------------------
cat <<NEXT_EOF

Instalação concluída.

Próximos passos:
  "$INSTALLED_BINARY" status    # ver porta, dispositivos e conexões
  "$INSTALLED_BINARY" pair      # abrir janela de pareamento (5 min)
  "$INSTALLED_BINARY" doctor    # diagnosticar permissões e configuração

Logs:
  $STDOUT_LOG
  $STDERR_LOG

Para desinstalar: scripts/uninstall-macos.sh
NEXT_EOF
