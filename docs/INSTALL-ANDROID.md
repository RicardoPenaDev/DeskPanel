# Android installation

## Install the released APK

1. Download `DeskPanel-0.1.1.apk`.
2. Copy it to the Android phone.
3. Open it with the file manager and allow installation from that source when Android asks.
4. Launch DeskPanel.

## Pair with macOS

1. Open **Settings > Connect to Mac** in DeskPanel.
2. Tap **Scan QR Code**.
3. Scan the QR code shown by the macOS setup page.

Manual IP, port and temporary code entry remains available as a fallback.

## Build the APK

```bash
scripts/build-apk.sh
scripts/build-apk.sh --release
```

The script installs dependencies when needed, runs tests, builds the web bundle, synchronizes Capacitor and invokes Gradle.

## Install with ADB

Enable Developer options and USB debugging, connect the device, authorize the computer and run:

```bash
adb devices
adb install releases/v0.1.1/DeskPanel-0.1.1.apk
```

To update an existing installation while preserving app data:

```bash
adb install -r releases/v0.1.1/DeskPanel-0.1.1.apk
```
