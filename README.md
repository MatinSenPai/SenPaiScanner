# MAGIK

> **Persian / فارسی:** [README.fa.md](README.fa.md)

[![CI](https://github.com/ehsatn/MAGIK/actions/workflows/ci.yml/badge.svg)](https://github.com/ehsatn/MAGIK/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ehsatn/MAGIK?style=flat-square)](https://github.com/ehsatn/MAGIK/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ehsatn/MAGIK?style=flat-square)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows%20%7C%20android%20%7C%20termux-informational?style=flat-square)](#installation)

MAGIK is a fork of [MatinSenPai/SenPaiScanner](https://github.com/MatinSenPai/SenPaiScanner), renamed and polished with a simpler desktop GUI.

A Cloudflare IP finder with a terminal UI and an Android app, built for networks where latency is unpredictable and connections drop without warning. Probe Cloudflare edge IPs, optionally validate them through your VLESS or Trojan config with embedded xray — no commands to memorize.

---

## Features

*   **Cloudflare IP Scanning**: Quickly finds working Cloudflare IPs.
*   **Terminal UI (TUI)**: Interactive menu-driven interface, no complex CLI flags.
*   **Desktop GUI**: A clean local web GUI with dark mode and coding fonts.
*   **Multi-Platform**: Linux, macOS, Windows, Android (APK & Termux).
*   **Proxy Validation**: End-to-end testing of IPs using VLESS/Trojan configurations (via embedded Xray).
*   **Neighbor Scan**: Explores nearby IPs in the same Cloudflare block for more hits.
*   **Persistency**: Saves last scan settings automatically.
*   **Clipboard & File Output**: Copy working IPs to clipboard and save to `ips.txt`.

## How it works

Run `magik` and you land in a short menu. Navigate with arrow keys and Enter — no scan-related CLI flags.

```
┌────────────────────────────────────────────────────────────┐
│  ▶  Find Working IPs   scan Cloudflare IPs — config optional │
│     Retry Last Scan    retry last scan with previous config  │
│     About                                                │
│     Quit                                                 │
└────────────────────────────────────────────────────────────┘
```

**Find Working IPs** can run in one or two phases:

1.  **Phase 1 — Connectivity scan** probes candidate Cloudflare IPs. Without a config URL it uses a standard HTTP probe; with a URL it derives SNI, host, WebSocket path, and port from your link. MAGIK intelligently parses your VLESS or Trojan configuration URL. In **Random** mode, healthy hits also trigger a **neighbor scan** — nearby addresses in the same Cloudflare block are explored automatically.
2.  **Phase 2 — xray validation** (optional) launches an embedded xray instance and tests the best Phase 1 hits end-to-end through your actual VLESS/Trojan config. Results show endpoint, transport type, download speed, latency (TTFB), and pass/fail status.

Press **`c`** when a scan finishes to copy working `IP:port` endpoints to the clipboard and save them to `ips.txt` next to the binary (or current working directory).

Your last scan settings are saved automatically. Use **Retry Last Scan** on the home screen to repeat the previous run without re-entering anything.

---

## Installation

### Desktop — pre-built binary

Download from the [releases page](https://github.com/ehsatn/MAGIK/releases/latest).

| Platform | Architecture | File |
|---|---|---|
| Linux | x86_64 | `magik-linux-amd64` |
| Linux | ARM64 | `magik-linux-arm64` |
| Linux | 32-bit x86 | `magik-linux-386` |
| macOS | Intel | `magik-darwin-amd64` |
| macOS | Apple Silicon | `magik-darwin-arm64` |
| Windows | x86_64 | `magik-windows-amd64.exe` |
| Windows | 32-bit x86 | `magik-windows-386.exe` |

**Linux / macOS:**

```bash
# stable release
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash

# pre-release
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash -s -- --prerelease
```

**Windows (PowerShell):**

```powershell
$r = Invoke-RestMethod https://api.github.com/repos/ehsatn/MAGIK/releases/latest
$url = ($r.assets | Where-Object name -eq "magik-windows-amd64.exe").browser_download_url
Invoke-WebRequest $url -OutFile magik.exe
```

### Desktop GUI

The classic terminal UI is now built as `magik`. To build the graphical desktop launcher:

```bash
make build-gui
```

On Windows, build the GUI binary without a console window:

```bash
make build-gui-windows-amd64
```

Run `magik-gui` and it opens the dark browser-based interface locally.

![MAGIK desktop GUI](docs/assets/magik-gui.png)

### Android — pre-built APK

Signed release APKs are attached to each GitHub release:

| File pattern | Description |
|---|---|
| `MAGIK-{version}-universal-release.apk` | All ABIs (recommended) |
| `MAGIK-{version}-arm64-v8a-release.apk` | 64-bit ARM only |
| `MAGIK-{version}-armeabi-v7a-release.apk` | 32-bit ARM only |

Install the APK on your device (enable “Install from unknown sources” if needed), grant network permission, and tap **START SCAN** on the home screen.

### Termux (Android terminal)

Run the full desktop TUI inside [Termux](https://termux.dev/) — same workflow as Linux, including Phase 2 xray validation, persistent config, live results, and neighbor scan.

**1. Install Termux** from [F-Droid](https://f-droid.org/en/packages/com.termux/) (not the Play Store build). Open the app and run:

```bash
pkg update && pkg upgrade -y
pkg install curl tar -y
```

**2. Install MAGIK** (auto-detects Termux and installs to `$PREFIX/bin`):

```bash
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash
```

Pre-release channel:

```bash
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash -s -- --prerelease
```

The installer downloads `magik-linux-arm64` on 64-bit phones. (32-bit ARM devices are uncommon; use the native APK if the Linux binary is unavailable.)

**3. Run:**

```bash
magik
```

**Termux tips**

| Topic | Notes |
|---|---|
| **Navigation** | Arrow keys on the on-screen keyboard, or a Bluetooth keyboard. `k` / `j` / `h` / `l` also work in menus. |
| **Paste config URL** | Long-press in Termux → Paste, or `termux-clipboard-get` if `termux-api` is installed. *For reliable clipboard access, ensure `pkg install termux-api` is run and permissions are granted.* |
| **Clipboard (`c` key)** | May not work in all Termux setups by default. Results are always saved to `ips.txt` in the current directory when copy runs — use that file if clipboard fails. |
| **`ips.txt` / live results** | Keep files in `~/` (e.g. `cd ~` before starting). Paths shown in the TUI are relative to the working directory. |
| **Config file** | `~/.config/magik/config.json` — powers **Retry Last Scan**. |
| **Long scans** | Optional: `termux-wake-lock` (from `pkg install termux-api`) to reduce the screen turning off mid-scan. |
| **Update** | Re-run the `install.sh` one-liner; it upgrades to the latest stable release. |

---

## Troubleshooting / FAQ

*   **"Invalid URL error" or scan failures with a valid config**: Ensure your VLESS/Trojan configuration URL is correctly formatted and accessible. Check for typos. If the issue persists, the target server or Xray setup might be rejecting the connection. Consider testing the URL with a standalone Xray client first.
*   **Scanner gets stuck or crashes**: This can be due to high network latency, an unstable internet connection, or an issue with the Cloudflare IPs being probed. Try restarting the scan, or if persistent, check your system's resource usage. If it's a bug, please [open an issue](https://github.com/ehsatn/MAGIK/issues/new/choose) with details.
*   **No IPs found**: If the scan completes but finds no working IPs, it's possible that all probed IPs are blocked or unavailable in your region, or your network conditions are too poor for successful probes. Try scanning at a different time or from a different network.
*   **Clipboard not working in Termux**: As noted in the Termux tips, you might need to install `termux-api` (`pkg install termux-api`) and grant necessary permissions. If it still fails, rely on `ips.txt` for your results.
*   **Slow download speeds/high latency from found IPs**: The scanner validates connectivity and basic speed, but real-world performance can vary greatly based on network congestion, server load, and geographical distance.

---

## Contributing

We welcome contributions! If you're interested in helping develop MAGIK, please see our [CONTRIBUTING.md](CONTRIBUTING.md) guide for details on how to set up your development environment, propose changes, report bugs, or suggest new features.

Ideas for contributions include:
*   Adding support for more proxy protocols (e.g., Shadowsocks, WireGuard).
*   Improving the scanning algorithms or UI.
*   Enhancing the Android app experience.

---

## License

MAGIK is released under the MIT License. See [LICENSE](LICENSE) for more details.
