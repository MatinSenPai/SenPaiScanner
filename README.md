# SenPai Scanner

[![CI](https://github.com/matinsenpai/senpaiscanner/actions/workflows/ci.yml/badge.svg)](https://github.com/matinsenpai/senpaiscanner/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/matinsenpai/senpaiscanner?style=flat-square)](https://github.com/matinsenpai/senpaiscanner/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/matinsenpai/senpaiscanner?style=flat-square)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-informational?style=flat-square)](#installation)

A Cloudflare IP finder with a terminal UI, built for networks where latency is unpredictable and connections drop without warning. Paste your VLESS or Trojan config, pick your settings, and let it find IPs that actually work through your proxy — no commands to memorize.

---

## How it works

Run `senpaiscanner` and you land in a short menu. Navigate with arrow keys and Enter — no flags, no subcommands.

```
┌────────────────────────────────────────────────────────────┐
│  ▶  Find Working IPs   paste a config and test CF IPs      │
│     About                                                │
│     Quit                                                 │
└────────────────────────────────────────────────────────────┘
```

**Find Working IPs** runs in two phases:

1. **Phase 1 — Connectivity scan** probes candidate Cloudflare IPs using settings derived from your config URL (SNI, host, WebSocket path, port). It checks trace reachability and, for WebSocket configs, whether a WS-style TLS connection survives DPI.
2. **Phase 2 — xray validation** launches an embedded xray instance and tests the best Phase 1 hits end-to-end through your actual VLESS/Trojan config. Results show endpoint, transport type, download speed, latency (TTFB), and pass/fail status.

When Phase 2 finishes, press **`c`** to copy working `IP:port` endpoints to the clipboard and save them to `ips.txt` next to the binary (or current working directory).

---

## Installation

### Pre-built binary

Download from the [releases page](https://github.com/matinsenpai/senpaiscanner/releases/latest).

| Platform | Architecture | File |
|---|---|---|
| Linux | x86_64 | `senpaiscanner_linux_x86_64.tar.gz` |
| Linux | ARM64 | `senpaiscanner_linux_arm64.tar.gz` |
| macOS | Intel | `senpaiscanner_darwin_x86_64.tar.gz` |
| macOS | Apple Silicon | `senpaiscanner_darwin_arm64.tar.gz` |
| Windows | x86_64 | `senpaiscanner_windows_x86_64.zip` |

**Linux / macOS:**

stable release:
```bash
curl -fsSL https://github.com/MatinSenPai/SenPaiScanner/raw/refs/heads/main/install.sh | bash
```

pre-release:
```bash
curl -fsSL https://github.com/MatinSenPai/SenPaiScanner/raw/refs/heads/main/install.sh | bash -s -- --prerelease
```

**Windows (PowerShell):**
```powershell
$r = Invoke-RestMethod https://api.github.com/repos/matinsenpai/senpaiscanner/releases/latest
$url = ($r.assets | Where-Object name -like "*windows*x86_64*").browser_download_url
Invoke-WebRequest $url -OutFile senpaiscanner.zip
Expand-Archive senpaiscanner.zip .
```

### From source

```bash
go install github.com/matinsenpai/senpaiscanner/cmd/senpaiscanner@latest
```

---

## Usage

```bash
senpaiscanner              # open the TUI
senpaiscanner --version    # print version and exit
senpaiscanner -v           # same
senpaiscanner version      # same
```

Everything else is inside the TUI — there are no scan-related CLI flags.

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | move between rows |
| `←` / `→` or `h` / `l` | move between options within a row |
| `Enter` | select / confirm / start |
| `Esc` | go back |
| `q` | quit from menu; during a scan, cancel or return to menu when finished |

On the **Config URL** row, `←` / `→` move the text cursor; `Ctrl+A` / `Ctrl+E` jump to start / end. Vim keys `h` / `j` / `k` / `l` type normally into the URL field on that row.

---

## Find Working IPs

Paste a **`vless://`** or **`trojan://`** share URL, adjust the setup rows, then start the scan.

### Setup rows

| Row | Options | Notes |
|---|---|---|
| **Config** | paste URL | your VLESS/Trojan share link |
| **Source** | Random / From File | random Cloudflare IPv4 ranges, or read candidates from `ips.txt` |
| **Count** | 1,000 / 5,000 / 20,000 / Custom | IPs to probe in Phase 1; **ignored when Source is From File** |
| **Workers** | 50 / 100 / 200 / Custom | parallel probers (default 50 — safe on restricted networks) |
| **Timeout** | 2s / 3s / 5s / Custom | per-probe deadline (default 5s) |
| **Ports** | Config, 443, 8443, 2053, 2083, 2087, 2096 | multi-select; each IP is tested on every selected port |

**Source → From File:** place `ips.txt` (one IP per line) next to the binary or in the directory you run from. All listed IPs are probed on the selected ports; the Count row is ignored.

**Ports row:** use `←` / `→` to focus a port pill, then **`Space`** or **`Enter`** to toggle it. Select **Config** alone to use the port from your URL. Selecting multiple ports multiplies Phase 1 work (IPs × ports).

**Starting the scan:** `Enter` on the URL row moves to the next row. On Count, Workers, or Timeout, `Enter` on **Custom** opens a text field. From any other row, `Enter` starts Phase 1 (after validating the URL).

### Phase 1 — Finding reachable IPs

Probes use HTTP validation against Cloudflare's edge, with SNI/host/path taken from your config URL. For WebSocket configs, a WS upgrade probe is required for an IP to count as healthy. TLS certificate verification is skipped in this phase — Phase 2 validates through xray properly.

Press `q` / `Esc` to cancel and return to the menu.

### Phase 2 — xray validation

The top Phase 1 candidates are tested through an embedded xray instance with your config. The live table shows:

| Column | Meaning |
|---|---|
| **ENDPOINT** | `IP:port` that was validated |
| **TYPE** | transport (`ws`, `grpc`, `xhttp`, …) |
| **SPEED** | measured download throughput in Mbps, or `n/a` if speed could not be measured |
| **LATENCY** | time to first byte through the proxy (TTFB) |
| **STATUS** | ✓ working / ✗ failed |

| Key | Action |
|-----|--------|
| `c` | copy working endpoints to clipboard **and** save to `ips.txt` |
| `q` / `Esc` | return to the main menu |

Exported lines look like `104.16.72.162:443` — ready to paste into client configs or DNS/IP lists.

### About

Version string and short project blurb; `Enter` / `q` / `Esc` back to the menu.

---

## Tips for restricted networks

**Start with defaults.** 5,000 random IPs, 50 workers, 5s timeout, and the config port are a good baseline on lossy or filtered lines.

**Use From File after a partial run.** Copy working endpoints with `c`, edit `ips.txt`, then re-run with **Source → From File** to validate only your shortlist on more ports.

**Try multiple ports.** Cloudflare CDN ports (443, 8443, 2053, …) behave differently under DPI. Multi-port selection lets Phase 1 find the best `IP:port` pair before xray validation.

**WebSocket configs need WS-friendly IPs.** Phase 1 runs an idle TLS hold plus a WebSocket upgrade check when your URL uses `type=ws`. An IP that passes trace but fails WS will not become a Phase 2 candidate.

**0% loss alone is not enough.** For HTTP-style probing, non-zero download throughput or a successful WS check is required for an IP to count as healthy.

**Speed in Phase 2 is best-effort.** Connectivity is confirmed via `cp.cloudflare.com/cdn-cgi/trace`. Download speed is measured when possible (Cloudflare speed endpoint, config host, or parallel trace sampling). If speed cannot be measured reliably, the endpoint can still show ✓ with SPEED `n/a`.

---

## FAQ

**Why doesn't it just run a ping?**
Cloudflare drops ICMP on their edge IPs. SenPai Scanner validates HTTP/TLS behaviour and, for proxy configs, runs traffic through xray — closer to real VLESS/Trojan usage than ping or bare TCP.

**How is this different from warp-plus?**
SenPai Scanner does not run a permanent proxy. It finds and validates Cloudflare IPs for **your** xray config and exports `IP:port` lists you can plug into Sing-Box, v2rayN, etc.

**Where do the IP ranges come from?**
Embedded from Cloudflare's official published lists (`cloudflare.com/ips-v4`, `cloudflare.com/ips-v6`). The binary ships with a snapshot; ranges rarely change.

**"ips.txt not found" when using From File**
Place `ips.txt` next to the executable or in your current working directory before starting. One IP or CIDR per line; `#` comments and CSV lines (IP/CIDR in the first column) are supported.

**The scan feels slow with many ports selected**
Each selected port is probed for every IP. Testing 5 ports on 5,000 IPs means 25,000 probes in Phase 1 — lower Count or narrow the port list if needed.

**What happened to Quick Scan, Custom Scan, Test IPs, and Discover Colos?**
Those separate menu flows were removed to focus on one workflow: paste your config, find working endpoints, export results. The core probe engine is still used internally by **Find Working IPs**.

---

## Building from source

```bash
git clone https://github.com/matinsenpai/senpaiscanner.git
cd senpaiscanner
make build          # current platform
make build-all      # all platforms → dist/
make test
make install        # to $GOPATH/bin
```

**Windows (cross-compile all platforms):**
```powershell
powershell -ExecutionPolicy Bypass -File build.ps1
# optional: -Version "0.4.0"
```
Binaries land in `dist/`.

---

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for project principles, development setup, and pull request guidelines.

Issues and PRs are welcome. For larger changes, open an issue first to discuss scope.

For bugs, include your OS/arch, version (`senpaiscanner --version`), the screen you were on, and what you expected vs what happened.

---

## Roadmap

- Configurable download/upload thresholds for final filtering
- Persistent settings saved between sessions
- `Watch` mode for continuous monitoring
- Export directly to xray/Sing-Box JSON from the results screen

---

## License

MIT — see [LICENSE](LICENSE).
