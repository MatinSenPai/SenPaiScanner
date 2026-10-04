# SenPai Scanner — Release Notes

## v1.1.1

- **Android now has every desktop feature**: Anti-DPI settings (editable recipe, validation, *Suggested values*), *IP source: Paste list* with a live address count, *Skip the reachability scan*, the *Fast / Gentle* scan profile, and *Resume scan* after an interruption. The new black / white / red look with square corners is applied too
- The Android engine now runs on the same shared scan runner as the desktop app and the plain text mode, so scans, targets, Gentle, Anti-DPI and resume behave identically on all platforms
- Fix: resuming a scan counted the healthy addresses found earlier but did not show them in the results list (desktop and Android)
- Fix: the Android app did not send options left at their default (WebSocket requirement, Anti-DPI on), so they could be silently off
- Tested on an Android 16 emulator: pasted targets, direct speed test, Gentle scan, Stop, saved state, Resume

## v1.1.0

- **Anti-DPI**: the TLS ClientHello of every probe (and of the xray tunnel validation) can be cut into small pieces so DPI cannot match the SNI. Defaults are the published values from t.me/MatinSenPaii/5469 (`tlshello` 0/104/1 + first-packet 114/1, delay 1 ms, maxSplit 11, fingerprint `unsafe`, ALPN `http/1.1`); every field is editable in the desktop app, with a "Suggested values" button
- Anti-DPI switch in the TUI config screen and in the Android scan config; values persist as `anti_dpi` in the shared config file
- New black / white / red look (square corners, numbered sections, Vazirmatn + JetBrains Mono) across the desktop GUI, the terminal UI and the Android app, with a light theme in the desktop app
- New logo, banner and social preview; all desktop and Android icons regenerated (`python gen_icons.py`)
- **Plain text mode for screen readers** (`senpaiscanner --plain`, `scan ...`, `help`): one sentence per line, no colours or redraws, throttled progress announcements, final list and `-output` file (#115 #124 #125)
- **Paste your own targets**: IPs, CIDRs, ranges and domains in the desktop app and the CLI; domains are resolved; invalid entries are reported (#113)
- **Skip the reachability scan** and test the pasted list directly, through your config or with a direct download sample (#113)
- **Gentle mode** (desktop profile switch, terminal UI row, `-gentle`): 25 workers max, 6 s timeout min, 40 probes/s, for ISPs that cut the connection during scans (#25 #56 #62 #96 #75)
- **Save and resume scans**: progress is saved every 20 s and on stop, the desktop app offers *Resume scan*, the CLI has `scan -resume` (#130 #64)
- 32-bit ARM CLI build (`cli-linux-armv7`) for 32-bit Termux (#111)
- Fix: the direct download speed test ("Speed test green results" without a config, and Android) judged a single-try probe with `IsHealthy`, which needs two tries, so every sample was reported as failed (#82)
- Fix: cancelling a Phase 2 scan on Android did not stop it; the Phase 2 table of the terminal UI dropped the failure reason; the live results file was not written for runs without healthy Phase 1 rows; CSV header and flush errors were ignored; the active-scan cancel function is now mutex-guarded (#133 #134 #107)
- Tests: byte-for-byte check of the Anti-DPI recipe, a probe test that proves the ClientHello arrives fragmented, and xray outbound tests

## v1.0.0

- Unified Signal Desk experience across Windows, Linux, macOS, and Android GUI builds
- Transparent SenPai logo applied to native desktop and Android application icons
- Android rebuilt with separate Scan, Results, and Export tabs plus a persistent session rail
- Android live copy actions for all green endpoints and the top 20 while scans are running
- Android post-stop speed testing, optional neighbor scanning, tunnel validation, and client exports
- GitHub release builds for all desktop GUIs, CLIs, and Android APK artifacts

## v0.8.0

### What's New

#### Desktop GUI (Wails v2)
- New cross-platform desktop application (`desktop/`) built with Wails v2 — Windows, macOS, and Linux
- Rebuilt “Signal Desk” UI with real Scan, Results, and Export tabs plus a persistent live session rail
- Copy every green endpoint or the top 20 unique IPs at any time, including while a scan is still running
- Stop a scan and speed-test all green results collected so far; uses xray when a share URL is present and direct Cloudflare samples otherwise
- Results and client exports now have separate workspaces, clear empty states, live filtering, and keyboard-accessible controls
- Full scan control: IP count (1k / 5k / 20k / custom), workers, timeout, ports, optional neighbor/WebSocket scanning, config URL, top-N, and speed thresholds
- Live progress events (`scan:stats`, `scan:result`, `scan:done`) streamed from Go to the UI
- Phase 2 xray validation of the top N healthy IPs inside the GUI (`validate:result`, `validate:done`)
- Config export built in: paste one `vless://` / `trojan://` / `vmess://` config URL, generate a full pack (subscription URLs, sing-box JSON, Clash YAML) for all working IPs, copy or save via native dialogs
- Build script `desktop/build_gui.ps1` — single-platform or cross-compile (`-All`)

#### Scan Control & Network Metadata
- Neighbor scanning is now opt-in in both the CLI and GUI and persists with the last scan settings
- ISP detection now merges Cloudflare metadata with parallel IPWhois/IPinfo HTTPS fallbacks and a Team Cymru DNS ASN fallback

#### MahsaNG-Style IP Sampling
- Cloudflare range pool (`internal/ipsrc/ranges_v4.txt`, 628 ranges) now sampled with the MahsaNG weighted-random algorithm: larger ranges are picked more often, every IP is unique, and small pools enumerate + shuffle instead of capping at 1024
- Original selectable counts (1000 / 5000 / 20000 / custom) are preserved — no hard limit
- Phase 1 and the Android app both use the same weighted sampling via `MahsaNGV4Stream`

#### Shared Export Package
- New `internal/export` package: `Generate`, `ParseEndpoints`, sing-box and Clash builders, shared by the TUI (`exportAllConfigs`) and the desktop GUI

### Other
- Removed the dead, unbuildable `pkg/scanner` package that blocked `go mod tidy` and full-module builds
- `go test ./...` and `go vet ./...` pass for the whole module

---

## v0.7.0

**Release date:** June 13, 2026
**Previous release:** v0.5.0 (May 30, 2026)
**Commits:** 20 | **Files changed:** 81 | **Lines:** +7,133 / -540

---

### What's New

#### Android Application
- Complete Android app built with Kotlin and Jetpack Compose (`android/`)
- MVVM architecture with `MainScreenViewModel` and `MainViewModel`
- Full UI implementation with theme support (Color, Theme, Type)
- Go Mobile bindings via `mobile/mobile.go` and `mobile/validate.go`
- CI/CD pipeline: debug and release APK builds with signing support
- Launcher icons for all Android densities (mdpi through xxxhdpi)
- Build scripts: `build_go_mobile.sh` and `build_go_mobile.bat`

#### New Proxy Protocols
- **Shadowsocks (SS)** support — parse `ss://` share URLs with base64-encoded payloads, method/password detection, and Xray outbound configuration
- **VMess** support — parse `vmess://` base64 JSON share URLs, build Xray outbound with `alterId: 0`
- Both protocols work through Phase 1 connectivity scan and Phase 2 xray validation

#### ISP Detection
- Fetches connection metadata from `speed.cloudflare.com/meta` on startup
- Falls back to `ip-api.com/json` when Cloudflare is unreachable (e.g. restricted networks)
- Displays ISP name and Cloudflare colo in the TUI header and Phase 1/Phase 2 pages

#### Multi-Format Export
Press `e` after Phase 2 to export working endpoints to:
- **Subscription URLs** — `senpaiscanner-sub.txt` (one `vless://` or `trojan://` URL per line)
- **Sing-Box JSON** — `senpaiscanner-singbox.json` with outbound array
- **Clash YAML** — `senpaiscanner-clash.yaml` with proxy list

#### Speed Settings
- **Min Speed filter** — discard Phase 2 results below a configurable Mbps threshold (None / 1 / 2 / 5 Mbps / Custom)
- **Custom Speed URL** — override the default `speed.cloudflare.com/__down` endpoint
- **Speed Size** — configurable download sample size (128 KB / 512 KB / 1 MB / 5 MB / Custom)
- **Upload Test** — optional Phase 2 setting that measures upload throughput by POSTing data through the proxy (toggle with space on the optional config page)

#### Persistent Configuration
- Scan settings saved to `~/.config/senpaiscanner/config.json` (Windows: `%AppData%/senpaiscanner/config.json`)
- **Retry Last Scan** menu option re-runs the previous scan with saved settings
- Saves: IP mode, count, workers, timeout, ports, config URL, top N, min speed, speed URL, speed size, upload test, RequireWS

#### Subnet & CIDR Support
- `ips.txt` now accepts CIDR notation (e.g. `104.16.0.0/13`) in addition to plain IPs
- For subnets larger than /24, randomly samples 256 IPs using O(log N) weighted selection
- Subnets <= /24 are expanded fully

#### Fallback SNI & Trace Handling
- Phase 1 probes fall back through multiple SNI hostnames (`speed.cloudflare.com`, `www.cloudflare.com`, `cloudflare.com`, etc.) when the primary is blocked
- Phase 2 connectivity check tries domain-first, then IP-based, then tunnel path, then data-path fallback
- Fallback trace URLs (`cp.cloudflare.com`, `cloudflare.com`) for restricted networks

---

### Bug Fixes

- **Phase 1 scan freezing** — workers could block while enqueueing neighbor probes, freezing progress at arbitrary tested counts. Routing now goes through a single queue owner with bounded WebSocket handshakes
- **Config URL input swallowing keys** — the character `l` and arrow keys were eaten by the text input handler when typing in the config URL field
- **Xray stdio restoration** — `os.Stdout`/`os.Stderr` redirection to `/dev/null` could leak file descriptors when `xcore.New()` failed, eventually exhausting FDs. Now uses `withSuppressedXrayOutput()` with proper defer cleanup
- **Share URL parsing** — hardened against missing `?` separator (common with terminal paste), truncated WS paths, and `worers.dev` typos (auto-corrected to `workers.dev`)
- **Empty live result file** — the live result file is no longer created until at least one healthy Phase 1 result arrives
- **Phase 1 results table alignment** — columns now align correctly across different terminal widths
- **IPv6 support** — dial addresses and URLs now use `net.JoinHostPort` for proper bracket notation
- **Concurrency & cancellation** — replaced blocking semaphore acquisition with nested select for immediate TUI cancel response
- **Dynamic timeouts** — Phase 2 xray validation timeout now scales with user-configured timeout and speed budget instead of using a fixed 30s
- **Input/output separation** — Phase 2 outputs write to `working_ips.txt`; custom IP inputs read strictly from `ips.txt`
- **Mobile context leak** — `mobile/mobile.go` now captures and defers context cancellations

---

### Testing

- VMess share URL parsing unit test
- Subnet parsing and CIDR expansion tests
- Persistent config and "Retry Last Scan" tests
- Phase 1 results table alignment tests
- Probe WebSocket host/path verification tests
- Xray stdio restoration test
- Builder `verifyPeerCertByName` test
- Parser tests for host normalization, path handling, and edge cases
- Concurrency and queue completion tests for Phase 1

---

### Documentation

- Updated `README.md` with new features, Android instructions, and troubleshooting
- Added `README.fa.md` — full Persian/Farsi translation
- Updated Termux tips, installation instructions, and FAQ

---

### Contributors

- **Matin SenPai** — fallback SNI, trace handling, xray improvements, docs
- **Mehdi (NaxonM)** — scan reliability, concurrency, persistent config, subnet parsing, UI
- **aradava** — VMess support, ISP info, multi-format export, speed settings, input fixes
- **Hidden-Node** — Android app, Go Mobile bindings, mobile connectivity improvements
- **Shayan SalehiRad** — share URL hardening, xray stdio fix, fallback trace targets
- **SofiaPetronelle (MercilessMarcel)** — Phase 1 freeze fix, table alignment
- **Erfan-0** — empty live result file fix
- **EDR Labs** — Shadowsocks support, validator package

---

### Upgrade Notes

- The config file location has changed from a flat `senpaiscanner-config.json` to `~/.config/senpaiscanner/config.json`. Old configs are not migrated automatically.
- `ips.txt` now supports CIDR notation — existing files with plain IPs continue to work unchanged.
- The default speed sample size increased from 64 KB to 128 KB for more reliable DPI detection on restricted networks.
- Phase 1 now requires 4 tries (up from 3) and at least 2 successful attempts for an IP to be marked healthy.
