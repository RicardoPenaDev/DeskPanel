# Project status

Last updated: 2026-09-20

## Current state

- Android dashboard tested on a physical Moto G60.
- macOS agent runs through a user LaunchAgent.
- DMG `v0.0.1` installs the agent and `DeskPanel.app` in `/Applications`.
- Local setup page generates a temporary QR pairing code.
- Android QR scanner pairs successfully with the Mac.
- Connection feedback, reconnection, rotation and editor UI are implemented.
- Keyboard injection actions were removed because macOS Accessibility/TCC blocks them reliably in the packaged agent.
- Release artifacts are stored in `releases/v0.0.1/`.

## Validation

- Android: 159 tests passing, TypeScript build passing, Capacitor sync passing.
- Android debug APK: built and installed on the connected Moto G60.
- macOS: agent doctor passes with the local LaunchAgent and port `38121`.
- DMG: verified and tested through a clean installation.

## Known limitations

- macOS Accessibility and Automation permissions may still be required for system-level actions.
- The project is local-network only; port `38121` must not be exposed to the public Internet.
- The release is not signed or notarized yet.
- The Android APK is a sideloaded debug artifact; Play Store publishing is not part of the MVP.

## Next steps

1. Sign and notarize the macOS application.
2. Add a native menu-bar experience.
3. Improve diagnostics and permission guidance.
4. Complete physical endurance and multi-device testing.
