# MAGIK

> **English:** [README.md](README.md)

[![CI](https://github.com/ehsatn/MAGIK/actions/workflows/ci.yml/badge.svg)](https://github.com/ehsatn/MAGIK/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ehsatn/MAGIK?style=flat-square)](https://github.com/ehsatn/MAGIK/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ehsatn/MAGIK?style=flat-square)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Platforms](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows%20%7C%20android%20%7C%20termux-informational?style=flat-square)](#نصب)

---

MAGIK یک فورک از [MatinSenPai/SenPaiScanner](https://github.com/MatinSenPai/SenPaiScanner) است که با نام جدید، رابط گرافیکی ساده‌تر و ظاهر مرتب‌تر ادامه داده شده.

اگر روی شبکه‌ای هستید که تأخیر بالاست، اتصال بی‌خبر قطع می‌شود و پیدا کردن یک IP کارآمد Cloudflare تبدیل به کار روزانه شده — MAGIK برای همین ساخته شده.

این ابزار IPهای لبه Cloudflare را پروب می‌کند، بهترین‌ها را با xray داخلی و کانفیگ VLESS یا Trojan خودتان تست می‌کند، و نتیجه را آماده‌ی paste کردن تحویل می‌دهد. بدون حفظ کردن فلگ‌های عجیب، بدون تنظیمات پیچیده — فقط ابزار را اجرا کنید و بگذارید کارش را بکند.

نسخه‌ی ترمینالی با `magik` ساخته می‌شود. برای رابط گرافیکی دسکتاپ، باینری `magik-gui` یک UI محلی دارک با فونت‌های کدنویسی باز می‌کند.

---

## چطور کار می‌کند

`magik` را اجرا کنید تا وارد یک منوی ساده شوید. با کلیدهای جهت‌نما و Enter جابه‌جا می‌شوید — هیچ فلگ CLI برای اسکن وجود ندارد.

```
┌────────────────────────────────────────────────────────────┐
│  ▶  Find Working IPs   scan Cloudflare IPs — config optional │
│     Retry Last Scan    retry last scan with previous config  │
│     About                                                    │
│     Quit                                                     │
└────────────────────────────────────────────────────────────┘
```

**Find Working IPs** در یک یا دو فاز کار می‌کند:

**فاز ۱ — اسکن اتصال:** کاندیداهای Cloudflare را پروب می‌کند. اگر کانفیگ URL ندهید، از HTTP استاندارد استفاده می‌شود؛ اگر بدهید، SNI، host، مسیر WebSocket و پورت از لینک شما گرفته می‌شود. در حالت **Random**، هر IP سالم به‌طور خودکار **اسکن همسایه** را فعال می‌کند — آدرس‌های نزدیک در همان بلوک Cloudflare هم بررسی می‌شوند.

**فاز ۲ — اعتبارسنجی xray** *(اختیاری):* یک نمونه xray داخلی راه‌اندازی می‌کند و بهترین نتایج فاز ۱ را با کانفیگ واقعی VLESS یا Trojan شما end-to-end تست می‌کند. خروجی شامل endpoint، نوع transport، سرعت دانلود، تأخیر (TTFB) و وضعیت pass/fail است.

بعد از اتمام، **`c`** را بزنید تا endpointهای سالم در کلیپ‌بورد کپی و در فایل `ips.txt` ذخیره شوند.

تنظیمات آخرین اسکن به‌صورت خودکار ذخیره می‌شود. **Retry Last Scan** همان اسکن را بدون وارد کردن مجدد تنظیمات تکرار می‌کند.

---

## نصب

### دسکتاپ — باینری آماده

از [صفحه Releases](https://github.com/ehsatn/MAGIK/releases/latest) دانلود کنید:

| پلتفرم | معماری | فایل |
|---|---|---|
| Linux | x86_64 | `magik-linux-amd64` |
| Linux | ARM64 | `magik-linux-arm64` |
| Linux | 32-bit x86 | `magik-linux-386` |
| macOS | Intel | `magik-darwin-amd64` |
| macOS | Apple Silicon | `magik-darwin-arm64` |
| Windows | x86_64 | `magik-windows-amd64.exe` |
| Windows | 32-bit x86 | `magik-windows-386.exe` |

**Linux / macOS — نصب با یک دستور:**

```bash
# نسخه پایدار
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash

# پیش‌انتشار
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash -s -- --prerelease
```

**Windows (PowerShell):**

```powershell
$r = Invoke-RestMethod https://api.github.com/repos/ehsatn/MAGIK/releases/latest
$url = ($r.assets | Where-Object name -eq "magik-windows-amd64.exe").browser_download_url
Invoke-WebRequest $url -OutFile magik.exe
```

### رابط گرافیکی دسکتاپ

برای ساخت نسخه GUI:

```bash
make build-gui
```

روی ویندوز، این هدف باینری GUI را بدون باز شدن پنجره CMD می‌سازد:

```bash
make build-gui-windows-amd64
```

بعد از اجرا، `magik-gui` رابط گرافیکی دارک را به‌صورت محلی در مرورگر باز می‌کند.

![رابط گرافیکی MAGIK](docs/assets/magik-gui.png)

### اندروید — APK آماده

APKهای امضاشده به هر Release ضمیمه می‌شوند:

| فایل | توضیح |
|---|---|
| `MAGIK-{version}-universal-release.apk` | همه ABIها — پیشنهادی |
| `MAGIK-{version}-arm64-v8a-release.apk` | فقط ARM 64-bit |
| `MAGIK-{version}-armeabi-v7a-release.apk` | فقط ARM 32-bit |

APK را نصب کنید، «نصب از منابع ناشناس» را در صورت نیاز فعال کنید، دسترسی شبکه بدهید و روی **START SCAN** بزنید.

### Termux

اگر می‌خواهید همان TUI کامل دسکتاپ را روی اندروید داشته باشید — از جمله اعتبارسنجی xray، فایل نتایج زنده و اسکن همسایه — Termux گزینه مناسبی است.

**۱. نصب Termux** از [F-Droid](https://f-droid.org/en/packages/com.termux/) (نسخه Play Store توصیه نمی‌شود):

```bash
pkg update && pkg upgrade -y
pkg install curl tar -y
```

**۲. نصب MAGIK:**

```bash
curl -fsSL https://github.com/ehsatn/MAGIK/raw/refs/heads/main/install.sh | bash
```

اسکریپت Termux را تشخیص می‌دهد و در `$PREFIX/bin` نصب می‌کند. روی گوشی ۶۴-bit، فایل `magik-linux-arm64` دانلود می‌شود.

**۳. اجرا:**

```bash
magik
```

**چند نکته برای Termux:**

| موضوع | توضیح |
|---|---|
| **ناوبری** | کلیدهای جهت‌دار یا کیبورد بلوتوث؛ در منوها `k` / `j` / `h` / `l` هم کار می‌کند |
| **Paste کردن URL** | لانگ‌پرس در Termux ← Paste |
| **کلید `c` (کلیپ‌بورد)** | ممکن است در همه setupها کار نکند؛ نتایج همیشه در `ips.txt` هستند |
| **فایل‌های نتایج** | قبل از اجرا `cd ~` کنید تا همه فایل‌ها در یک‌جا بمانند |
| **اسکن طولانی** | `termux-wake-lock` نصب کنید تا صفحه وسط اسکن خاموش نشود |
| **به‌روزرسانی** | همان one-liner را دوباره اجرا کنید؛ در صورت وجود نسخه جدید آپدیت می‌شود |

**نصب دستی (بدون اسکریپت):**

```bash
curl -fsSL -o "$PREFIX/bin/magik" \
  https://github.com/ehsatn/MAGIK/releases/latest/download/magik-linux-arm64
chmod +x "$PREFIX/bin/magik"
magik
```

### از سورس

```bash
go install github.com/ehsatn/MAGIK/cmd/magik@latest
```

---

## راهنمای استفاده

```bash
magik              # باز کردن TUI
magik --version    # نمایش نسخه
magik -v           # همان
magik version      # همان
```

بقیه کارها داخل TUI یا اپ اندروید است.

### کلیدهای ناوبری (TUI)

| کلید | عملکرد |
|---|---|
| `↑` / `↓` یا `k` / `j` | جابه‌جایی بین ردیف‌ها |
| `←` / `→` یا `h` / `l` | جابه‌جایی بین گزینه‌های یک ردیف |
| `Enter` | انتخاب / تأیید / شروع اسکن |
| `Esc` | برگشت |
| `q` | خروج از منو؛ حین اسکن: لغو یا بازگشت بعد از اتمام |

در ردیف **Config URL** کلیدهای `←` / `→` مکان‌نما را جابه‌جا می‌کنند؛ `Ctrl+A` / `Ctrl+E` به ابتدا و انتها می‌روند.

---

## تنظیمات اسکن (Find Working IPs)

### مرحله ۱ — پارامترهای اصلی

| ردیف | گزینه‌ها | توضیح |
|---|---|---|
| **Source** | Random / From File | محدوده تصادفی IPv4 Cloudflare، یا لیست از `ips.txt` |
| **Count** | 1,000 / 5,000 / 20,000 / Custom | تعداد IP در فاز ۱ |
| **Workers** | 50 / 100 / 200 / Custom | پروبرهای موازی — پیش‌فرض ۵۰ مناسب شبکه‌های ضعیف‌تر است |
| **Timeout** | 2s / 3s / 5s / Custom | مهلت هر پروب |
| **Ports** | Config, 443, 8443, 2053, 2083, 2087, 2096 | چندانتخابی؛ هر IP روی تمام پورت‌های انتخاب‌شده تست می‌شود |

روی **Ports** Enter بزنید تا به مرحله بعد بروید. برای toggle هر پورت، روی آن فوکوس کنید و `Space` یا `Enter` بزنید.

### مرحله ۲ — کانفیگ اختیاری

| ردیف | گزینه‌ها | توضیح |
|---|---|---|
| **Config** | paste URL یا خالی | خالی = فقط فاز ۱؛ با URL = فاز ۱ + فاز ۲ |
| **Top N** | 10 / 25 / 50 / 100 / All / Custom | تعداد نتایج فاز ۱ که در فاز ۲ اعتبارسنجی می‌شوند |

لینک‌های پشتیبانی‌شده: **`vless://`** و **`trojan://`**. پارسر لینک‌های واقعی را هم می‌پذیرد.

### تنظیمات پایدار

هر بار که اسکن را شروع می‌کنید، تنظیمات در این مسیرها ذخیره می‌شوند:

| پلتفرم | مسیر |
|---|---|
| Windows | `%AppData%\magik\config.json` |
| macOS | `~/Library/Application Support/magik/config.json` |
| Linux / Termux | `~/.config/magik/config.json` |

**Retry Last Scan** در صفحه اصلی همین مقادیر را بارگذاری و فوراً اسکن را شروع می‌کند.

---

## فایل نتایج زنده

حین اسکن، یک فایل با نام `MAGIKResult-YYYYMMDD-HHMMSS.txt` کنار باینری ساخته می‌شود. این فایل:

- تا رسیدن اولین نتیجه سالم ساخته نمی‌شود (فایل خالی placeholder نمی‌سازد)
- با هر به‌روزرسانی بازنویسی می‌شود — می‌توانید آن را در ادیتور tail کنید
- شامل طرح اسکن، جدول فاز ۱، و جدول فاز ۲ (در صورت وجود) است

---

## جزئیات فاز ۱ و ۲

### فاز ۱ — پیدا کردن IPهای در دسترس

| حالت | رفتار |
|---|---|
| بدون URL کانفیگ | پروب HTTP استاندارد (`speed.cloudflare.com`، نمونه ۶۴ KiB) |
| با URL کانفیگ | SNI / host / path از لینک؛ برای `type=ws` ارتقای WebSocket انجام می‌شود |

در حالت Random، با پیدا شدن هر IP سالم، آدرس‌های نزدیک در همان بلوک هم به‌صورت خودکار پروب می‌شوند (تا شعاع ۳۲، حداکثر ۱۲ همسایه به ازای هر hit).

جدول زنده ۲۰ نتیجه برتر را نشان می‌دهد: **ENDPOINT**، **LOSS**، **AVG(ms)**، **COLO**، **STATUS**.

### فاز ۲ — اعتبارسنجی xray

بهترین کاندیداهای فاز ۱ با xray داخلی تست می‌شوند:

| ستون | معنی |
|---|---|
| ENDPOINT | `IP:port` اعتبارسنجی‌شده |
| TYPE | نوع transport (ws, grpc, xhttp, …) |
| SPEED | throughput دانلود بر حسب Mbps، یا `n/a` |
| LATENCY | زمان تا اولین بایت (TTFB) |
| STATUS | ✓ موفق / ✗ ناموفق |

اتصال با `/cdn-cgi/trace` بررسی می‌شود. endpoint می‌تواند با SPEED مقدار `n/a` هم ✓ باشد. هر کاندید یک retry خودکار روی شکست دارد.

بعد از فاز ۲، **`c`** را بزنید تا endpointهای سالم (مثل `104.16.72.162:443`) در کلیپ‌بورد کپی و در `ips.txt` ذخیره شوند.

---

## فرمت `ips.txt` (From File)

فایل را کنار executable یا در پوشه کاری فعلی قرار دهید:

| نوع خط | مثال | رفتار |
|---|---|---|
| IPv4 ساده | `104.16.72.162` | بارگذاری |
| CSV | `104.16.72.162,note` | ستون اول خوانده می‌شود |
| کامنت / خالی | `# my list` | نادیده گرفته می‌شود |
| CIDR کوچک (≤۲۵۶ host) | `104.16.72.160/29` | expand کامل |
| CIDR بزرگ | `104.16.0.0/16` | نمونه تصادفی تا ۲۵۶ IP |
| CIDR نامعتبر | `not-a-cidr/99` | خطا و توقف اسکن |

خطوط IPv6 نادیده گرفته می‌شوند.

**یک workflow مفید:** اسکن Random بزنید ← با `c` نتایج را در `ips.txt` ذخیره کنید ← همان لیست را با **From File** روی پورت‌های بیشتر تست کنید.

---

## اپ اندروید

موتور پروب و منطق xray از طریق پل Go mobile (`gomobile bind`) روی اندروید اجرا می‌شود.

### بخش‌های اصلی UI

| بخش | امکانات |
|---|---|
| **Home** | کارت آمار (Tested، In-Flight، Healthy، Failed)؛ لیست IP؛ کپی تکی و گروهی |
| **Settings** | Source، Count، Workers، Timeout، Ports، Config URL، Top N |
| **FAB** | START SCAN / STOP SCAN |
| **Info** | معرفی، نسخه، لینک GitHub و تلگرام |

### تفاوت‌های اندروید با دسکتاپ

| قابلیت | TUI دسکتاپ | اندروید |
|---|---|---|
| تنظیمات پایدار + Retry Last Scan | ✓ | — (فقط در همان session) |
| فایل نتایج زنده | ✓ | — |
| فقط فاز ۱ (بدون URL) | ✓ | ✓ |
| اسکن همسایه (Random) | ✓ | ✓ |
| CIDR در `ips.txt` | ✓ | فقط IP ساده |
| ذخیره در `ips.txt` | ✓ | فقط کلیپ‌بورد |

---

## نکاتی برای شبکه‌های محدود

**با پیش‌فرض‌ها شروع کنید.** ۵,۰۰۰ IP، ۵۰ worker، timeout ۵ ثانیه و پورت ۴۴۳ baseline خوبی روی خطوط lossy است.

**بعد از اسکن اول از From File استفاده کنید.** با `c` نتایج را بگیرید، `ips.txt` را ویرایش کنید، و shortlist را با پورت‌های بیشتر تست کنید.

**چند پورت امتحان کنید.** پورت‌های CDN Cloudflare زیر DPI رفتار متفاوتی دارند.

**به اسکن همسایه اعتماد کنید.** در حالت Random نیازی به بالا بردن Count نیست — هر hit سالم به‌صورت خودکار همسایه‌هایش را در صف قرار می‌دهد.

**کانفیگ WebSocket به IP سازگار با WS نیاز دارد.** برای `type=ws`، hold TLS + upgrade WS اجرا می‌شود. IPای که trace را رد کند ولی WS را نه، کاندید فاز ۲ نمی‌شود.

**۰٪ loss به تنهایی کافی نیست.** throughput غیرصفر یا WS موفق هم لازم است.

---

## سوالات متداول

**چرا ping نمی‌زند؟**
Cloudflare ICMP را روی edge IPها drop می‌کند. MAGIK رفتار HTTP/TLS واقعی را تست می‌کند که به استفاده عملی VLESS/Trojan نزدیک‌تر است.

**تفاوتش با warp-plus چیست؟**
MAGIK پروکسی دائمی راه نمی‌اندازد. IPهای Cloudflare را برای **کانفیگ xray شما** پیدا و اعتبارسنجی می‌کند و لیست `IP:port` export می‌دهد.

**محدوده IPها از کجاست؟**
از لیست‌های رسمی Cloudflare — `cloudflare.com/ips-v4` و `cloudflare.com/ips-v6` — که داخل باینری embed شده‌اند.

**«ips.txt not found» در From File چیست؟**
فایل را کنار executable یا در پوشه کاری فعلی بگذارید.

**اسکن با پورت‌های زیاد کند است.**
هر پورت برای هر IP پروب می‌شود. ۵ پورت × ۵,۰۰۰ IP برابر است با ۲۵,۰۰۰ پروب در فاز ۱. Count یا تعداد پورت‌ها را کاهش دهید.

**چرا tested از Count بیشتر است؟**
در حالت Random، اسکن همسایه تا ۴۰۰ پروب اضافه می‌کند. این رفتار عمدی است.

---

## ساخت از سورس

### دسکتاپ

```bash
git clone https://github.com/ehsatn/MAGIK.git
cd magik
make build          # پلتفرم فعلی
make build-all      # همه پلتفرم‌ها ← dist/
make test
make install        # در $GOPATH/bin
```

**Windows (cross-compile):**

```powershell
powershell -ExecutionPolicy Bypass -File build.ps1
# اختیاری: -Version "0.6.0"
```

تگ `v*` با GitHub Actions و GoReleaser آرشیو چندپلتفرمی و checksum منتشر می‌کند.

### اندروید

```bash
# ساخت کتابخانه Go mobile
cd android
./build_go_mobile.sh        # Linux / macOS
build_go_mobile.bat         # Windows

# ساخت APK
./gradlew :app:assembleDebug
./gradlew :app:assembleRelease   # نیاز به keystore
```

CI با push تگ، APK امضاشده را به Release ضمیمه می‌کند.

---

## مشارکت

[CONTRIBUTING.md](CONTRIBUTING.md) را ببینید.

Issue و PR خوش‌آمد است. برای تغییرات بزرگ اول یک issue باز کنید تا درباره جهت پروژه صحبت کنیم.

برای گزارش باگ لطفاً موارد زیر را ذکر کنید: OS/arch، نسخه (`magik --version`)، صفحه‌ای که در آن بودید، رفتار موردانتظار در مقابل آنچه دیدید.

---

## نقشه راه

- آستانه قابل تنظیم download/upload برای فیلتر نهایی
- تنظیمات پایدار روی اندروید
- حالت `Watch` برای مانیتورینگ مداوم
- export مستقیم به فرمت JSON برای xray / Sing-Box از صفحه نتایج

---

## مجوز

MIT — [LICENSE](LICENSE)
