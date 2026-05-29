# SenPai Scanner

[![CI](https://github.com/matinsenpai/senpaiscanner/actions/workflows/ci.yml/badge.svg)](https://github.com/matinsenpai/senpaiscanner/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/matinsenpai/senpaiscanner?style=flat-square)](https://github.com/matinsenpai/senpaiscanner/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/matinsenpai/senpaiscanner?style=flat-square)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-informational?style=flat-square)](#installation)

A Cloudflare IP scanner with a full terminal UI, built for networks where latency is unpredictable and connections drop without warning. Just run it — no commands to memorize.

It finds the Cloudflare edge nodes that actually work from your location, ranks them by real measured latency and packet loss, and lets you export the results for use in a v2ray/xray/Trojan proxy config.

---

## How it works

Run `senpaiscanner` and you land in a menu. From there you navigate everything with arrow keys and Enter — no flags, no subcommands.

```
┌──────────────────────────────────────┐
│  ▶  Quick Scan    scan random CF IPs │
│     Custom Scan   configure details  │
│     Test IPs      validate a list    │
│     Discover Colos find reachable DCs │
│     About                            │
│     Quit                             │
└──────────────────────────────────────┘
```

The scanner:
- Probes Cloudflare's IP ranges via **TCP**, **TLS handshake**, or **HTTP** validation (`http` / `https` in Custom Scan)
- In **TCP/TLS** mode with empty SNI, rotates through several well-known Cloudflare hostnames per try; **HTTP** mode uses `speed.cloudflare.com` for trace and download
- Measures latency, packet loss, jitter, composite score, and (in HTTP mode) a bounded multi-stream download sample
- Identifies the colo (Cloudflare PoP) behind each IP via `/cdn-cgi/trace` (with `CF-Ray` fallback)
- Shows live results in a color-coded table; after a scan, a **Results** screen lists the top healthy IPs by score
- Persists a SQLite leaderboard and warm-starts future scans with previously good IPs
- Writes **CSV, JSON, or TXT** only when you set an output path in **Custom Scan** (other modes are on-screen only)

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
senpaiscanner                      # open the TUI
senpaiscanner web                  # open the local Web UI on 127.0.0.1:8787
senpaiscanner web --web-port 8788  # change the port
senpaiscanner web --web-host 0.0.0.0 --web-port 8080  # expose on the LAN
senpaiscanner web --addr 127.0.0.1:9000               # full host:port override
senpaiscanner --version            # print version and exit
senpaiscanner -v           # same
senpaiscanner version      # same
```

Everything else is inside the TUI or the local Web UI.

### Web UI

Run `senpaiscanner web` and open `http://127.0.0.1:8787`. The web console exposes the same TCP/TLS/HTTP scanner with live counters, score-sorted results, cancel, copy, and the persisted leaderboard.

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | move between rows |
| `←` / `→` or `h` / `l` | move between options within a row |
| `Enter` | select / confirm / start scan |
| `Esc` | go back |
| `q` | quit from menu; during a live scan, cancel or go to Results when finished |

### Scan screens

#### Quick Scan
Opens a three-row setup screen. Use `↑/↓` to move between rows and `←/→` to pick a preset.

**Enter on a preset launches the scan immediately** (you do not need to visit every row). Workers and timeout use whatever is currently selected on those rows — on first open the defaults are **100** workers and **3s** timeout. Choose **Custom** on a row to type your own value; after the last row, Enter starts the scan.

**Count** — how many IPs to probe:

| Preset | Notes |
|--------|-------|
| **5,000** | A few minutes on most connections — good for a quick check |
| **20,000** | Covers the active slice of Cloudflare's edge well |
| **100,000** | Thorough sweep; expect 10–30 min depending on network |
| **Custom** | Type any number, e.g. `50000` |

**Workers** — parallel goroutines:

| Preset | Notes |
|--------|-------|
| **50** | Safe for slow or lossy lines; won't overwhelm limited bandwidth |
| **100** | Default — good balance on most connections |
| **200** | Faster on stable, low-latency networks |
| **Custom** | Any integer; values above 500 rarely help and may cause OS errors |

**Timeout** — per-probe deadline:

| Preset | Notes |
|--------|-------|
| **2s** | Aggressive — good for fast networks; drops slow IPs quickly |
| **3s** | Default — works well on most restricted networks |
| **5s** | Relaxed — use this if you see high loss counts or many timeouts |
| **Custom** | Any Go duration string: `4s`, `1500ms`, `8s` |

All other settings stay at defaults: HTTP validation mode, port 443, 4 tries per IP, and a small `speed.cloudflare.com` download sample for ranking real data transfer.

#### Custom Scan
A form where you configure:

| Field | Default | Notes |
|---|---|---|
| Count | 500 | IPs to probe; 0 = unlimited |
| Workers | 100 | parallel goroutines |
| Timeout | 3s | per-probe deadline |
| Tries | 4 | probes per IP (for loss/jitter) |
| Port | 443 | 443 or 80 |
| CIDR | (all CF) | e.g. `104.16.0.0/13` |
| Output | (none) | `.csv`, `.json`, or `.txt` |
| Colo filter | (all) | e.g. `FRA,AMS` |
| SNI | (empty) | override hostname; empty = rotate (TCP/TLS) or `speed.cloudflare.com` (HTTP) |
| Mode | HTTP | `http` (or `https`), `tls`, or `tcp` — cycle with **Ctrl+← / Ctrl+→** |
| IPv4 / IPv6 | v4 on | toggle with F2 / F3 |
| Profile | Turbo | toggle Turbo/Stealth jitter with F4 |

Navigate fields with Tab / Shift+Tab (or ↑/↓). Press Enter to start. Timeout accepts Go durations like `1500ms` or `5s`; a plain number is treated as seconds.

Set **Output** to a path ending in `.csv`, `.json`, or `.txt` to stream results to disk during the scan. Quick Scan, Test IPs, and Discover Colos do not write files.

#### Test IPs
Put IPs in `ips.txt` in the **current working directory** (one per line, or CSV with IP in the first column) **before** you select this menu item. The scan **starts as soon as you choose Test IPs** from the menu.

Settings: HTTP on port 443, 6 tries, 10s timeout, 20 workers, fixed SNI `speed.cloudflare.com`, 512 KiB download sample, colo via `/cdn-cgi/trace`. Good for confirming candidates before putting them in a proxy config.

#### Discover Colos
Probes **300** random **IPv4** addresses via HTTP (`/cdn-cgi/trace` only — no download sample, 2 tries, 5s timeout). When finished, shows a table grouped by colo with average and best latency among healthy results.

### Live scan keys

| Key | Action |
|-----|--------|
| `s` | cycle sort: avg → loss → jitter → colo → speed (download) |
| `q` / `Esc` | while running: cancel and return to menu; when done: open **Results** |
| `Enter` | when done: open **Results** |

### Results screen

After Quick Scan or Custom Scan finishes, shows up to **20** healthy IPs (sorted by composite score). `s` re-sorts; `w` saves results; `Enter` / `q` / `Esc` return to the main menu.

### About

Version string and short project blurb; `Enter` / `q` / `Esc` back to the menu.

---

## Output formats

Only **Custom Scan** writes a file, when **Output** is set. Rows are appended in real time as results arrive.

**CSV** (`.csv`):
```
ip,score,loss_pct,avg_ms,min_ms,max_ms,jitter_ms,download_kbps,speed_tested,colo,tls_ok,http_status
104.21.14.53,91.8,0.0,87.40,82.10,93.60,4.20,540.8,true,FRA,true,200
```

**JSONL** (`.json`):
```json
{"ip":"104.21.14.53","score":91.8,"loss_pct":0,"avg_ms":87.4,"download_kbps":540.8,"speed_tested":true,"colo":"FRA","tls_ok":true}
```

**TXT** (`.txt`):
```
104.21.14.53    loss=0.0%   avg=87.40ms   jitter=4.20ms   dl=540.8KB/s   colo=FRA
```

---

## Tips for restricted networks

**HTTP mode by default.** A clean TCP/TLS handshake can still fail once real data starts moving. HTTP mode confirms Cloudflare with `/cdn-cgi/trace` and takes a small download sample from `speed.cloudflare.com` through the candidate IP.

**More tries = better loss estimate.** The default of 4 gives a rough picture. If your connection is particularly flaky, try Custom Scan with tries set to 6 or 8.

**Turbo vs Stealth.** Turbo is the default and removes inter-try sleeps for much faster large scans. Stealth adds small jitter between tries for more cautious validation on DPI-sensitive links.

**Leaderboard warm-start.** Healthy results are saved to a local SQLite leaderboard. Later scans retest those known-good IPs first, then continue into fresh random coverage.

**0% loss is not enough.** For Iran, a "clean" IP should pass HTTP validation and show non-zero download throughput. When speed testing is active, `0 KB/s` is treated as unhealthy even if latency and loss look good.

**Filter by colo.** If your proxy server is physically in one city, restrict to that colo so you're not routing traffic around the world: set colo filter to `FRA` for Frankfurt, `AMS` for Amsterdam, and so on.

**SNI.** In Custom Scan, leave SNI empty for automatic behavior: **TCP/TLS** rotates hostnames per try; **HTTP** (Quick Scan default) always uses `speed.cloudflare.com`. Override SNI if you need a specific hostname (especially for TLS/TCP on filtered networks).

**Final proxy validation still matters.** The strongest test is replacing the address in your VLESS/VMess/Trojan config and testing through Xray/V2Ray, because DPI behavior can depend on transport, SNI, path, ALPN, and fragmentation settings.

---

## FAQ

**Why doesn't it just run a ping?**
Cloudflare drops ICMP on their IPs. TCP on port 80 works but does not tell you much about TLS or payload transfer. SenPai Scanner defaults to HTTP validation because it is closer to real proxy traffic than TCP or TLS alone.

**How is this different from warp-plus?**
SenPai Scanner doesn't run a proxy or manage connections. It gives you a sorted list of fast Cloudflare IPs, which you plug into your own VLESS/VMess/Trojan config. Think of it as a diagnostic tool.

**IPv6 support?**
Enable it in Custom Scan with the F3 toggle. IPv6 availability on Iranian ISPs is inconsistent but it's there if you need it.

**The scan is taking forever.**
Lower workers and raise timeout: set concurrency to 30 and timeout to 8s. The default 3s is aggressive for high-latency connections.

**Where do the IP ranges come from?**
Embedded directly from Cloudflare's official published lists (`cloudflare.com/ips-v4`, `cloudflare.com/ips-v6`). The binary ships with a snapshot; there's no auto-update mechanism by design, since the ranges rarely change.

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

---

## Contributing

Issues and PRs welcome. For bigger changes, open an issue first to discuss.

For bugs, include your OS/arch, the screen you were on, and what happened.

---

## Roadmap

- Real Xray/V2Ray config validation with user-provided configs
- Configurable download/upload thresholds for final filtering
- Persistent settings saved between sessions
- `Watch` mode for continuous monitoring
- Export to xray/v2ray JSON config format directly from the results screen

---

## License

MIT — see [LICENSE](LICENSE).
