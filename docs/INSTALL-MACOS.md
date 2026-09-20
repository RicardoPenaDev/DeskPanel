# macOS installation

## Recommended: packaged DMG

1. Download `DeskPanel-0.1.1.dmg` from the release page.
2. Open the image and double-click **DeskPanel Installer.app**.
3. The installer places **DeskPanel.app** in `/Applications` and starts the user LaunchAgent.
4. Use the local setup page to generate a temporary pairing QR code.

The installer preserves an existing configuration and paired devices. Local data is stored under:

```text
~/Library/Application Support/DeskPanel/
~/Library/Logs/DeskPanel/
~/Library/LaunchAgents/dev.ricardopena.deskpanel.agent.plist
```

## Local development installation

```bash
scripts/install-macos.sh
```

Useful flags:

| Flag | Effect |
|---|---|
| `--binary <path>` | Use a prebuilt agent binary |
| `--force` | Replace an existing `config.json` |
| `--skip-launchagent` | Install files without loading launchd |
| `--no-load` | Write the plist without bootstrapping it |

## Verification

```bash
~/Library/Application\ Support/DeskPanel/bin/deskpanel-agent status
~/Library/Application\ Support/DeskPanel/bin/deskpanel-agent doctor
launchctl print gui/$(id -u)/dev.ricardopena.deskpanel.agent
```

The `doctor` command checks the configuration, local port, LaunchAgent, local connectivity and required macOS tools.

## Setup assistant

With the agent running, open **DeskPanel.app** from `/Applications`, or run:

```bash
~/Library/Application\ Support/DeskPanel/bin/deskpanel-agent setup
```

The assistant binds only to `127.0.0.1`, shows local Mac addresses and port `38121`, and generates a short-lived pairing code and QR code.

## Uninstall

```bash
scripts/uninstall-macos.sh
scripts/uninstall-macos.sh --purge
```

The normal uninstall removes the service and binary while preserving local data. `--purge` also removes configuration, paired devices and logs.

## Permissions

Some macOS actions require Accessibility or Automation permission. See [`SECURITY.md`](./SECURITY.md). Do not expose port `38121` directly to the public Internet.
