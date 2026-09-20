# Future roadmap

The roadmap preserves the core security model: Android sends only an `actionId`, and the Mac owns the authoritative action catalog.

## Near term

1. Complete physical endurance and multi-device testing.
2. Improve action feedback, diagnostics and permission guidance.
3. Add signed and notarized macOS distribution.
4. Add a native menu-bar experience around `DeskPanel.app`.
5. Improve QR pairing and local-network discovery with Bonjour/mDNS.

## Product improvements

- Media center with playback metadata and controls.
- Custom icons and more visual catalog options.
- Profiles based on the active Mac application.
- Optional Home Assistant integration.
- Better device management, revocation and safe diagnostics.

## Platform expansion

- Evaluate a Windows agent with the same protocol and least-privilege model.
- Keep platform-specific executors separate from shared authentication and transport code.

## Explicitly out of scope for the MVP

- Public Internet exposure.
- Arbitrary shell execution.
- Screen streaming, audio streaming or mouse control.
- Multi-user SaaS infrastructure.
- App Store or Google Play publication.
