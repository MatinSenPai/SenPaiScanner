# SenPai Scanner — Release Notes

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
