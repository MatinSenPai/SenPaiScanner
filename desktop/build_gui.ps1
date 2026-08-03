# =============================================================================
#  SenPai Scanner — desktop GUI build script (Wails v2)
#
#  Usage:
#    .\build_gui.ps1                      # windows/amd64 only
#    .\build_gui.ps1 -All                 # all common platforms (non-Windows
#                                         # targets are best-effort: cross-
#                                         # compiling webview targets needs the
#                                         # target OS toolchain)
#    .\build_gui.ps1 -Platform windows/arm64
#    .\build_gui.ps1 -Version v1.2.3      # override version tag
#
#  The script must live in desktop/ next to wails.json; wails build is run
#  from here. Output is copied to ../dist next to the CLI binaries.
# =============================================================================

param(
    [string]$Version = "1.0.0",
    [switch]$All,
    [string]$Platform = ""
)

$ErrorActionPreference = "Continue"
Set-Location -Path $PSScriptRoot   # desktop/ — wails.json lives here

# --- locate wails -----------------------------------------------------------
$Wails = (Get-Command wails -ErrorAction SilentlyContinue).Source
if (!$Wails -and (Test-Path "$env:USERPROFILE\go\bin\wails.exe")) {
    $Wails = "$env:USERPROFILE\go\bin\wails.exe"
}
if (!$Wails) {
    Write-Host "wails CLI not found. Install it with:" -ForegroundColor Red
    Write-Host "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Yellow
    exit 1
}

# --- version info (same fields as the root build.ps1) -----------------------
if (!$Version) { $Version = "1.0.0" }
$Commit    = ((git rev-parse --short HEAD 2>$null) -replace "`n", "")
$BuildDate = (Get-Date -Format "yyyy-MM-dd")
$LdFlags   = "-s -w " +
    "-X github.com/matinsenpai/senpaiscanner/pkg/version.Version=$Version " +
    "-X github.com/matinsenpai/senpaiscanner/pkg/version.Commit=$Commit " +
    "-X github.com/matinsenpai/senpaiscanner/pkg/version.BuildDate=$BuildDate " +
    "-X github.com/matinsenpai/senpaiscanner/pkg/version.BuiltBy=wails-local"

# --- targets ----------------------------------------------------------------
if ($Platform) {
    $targets = @($Platform)
} elseif ($All) {
    $targets = @("windows/amd64", "linux/amd64", "darwin/amd64", "darwin/arm64")
} else {
    $targets = @("windows/amd64")
}

$OutDir = Join-Path $PSScriptRoot "..\dist"
if (!(Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

Write-Host ""
Write-Host "  SenPai Scanner — GUI build" -ForegroundColor Cyan
Write-Host "  version  : $Version" -ForegroundColor Cyan
Write-Host "  commit   : $Commit" -ForegroundColor Cyan
Write-Host "  targets  : $($targets -join ', ')" -ForegroundColor Cyan
Write-Host ""

$ok  = 0
$err = 0

foreach ($t in $targets) {
    $outName = "senpaiscanner-gui-" + ($t -replace "/", "-")
    if ($t -like "windows/*") { $outName += ".exe" }
    Write-Host -NoNewline "  building $($t.PadRight(20))  ->  dist/$outName  "

    # Wails only regenerates the native Windows icon when icon.ico is absent.
    # Treat it as generated output so build/appicon.png is always authoritative.
    if ($t -like "windows/*") {
        $nativeIcon = Join-Path $PSScriptRoot "build\windows\icon.ico"
        if (Test-Path -LiteralPath $nativeIcon) {
            Remove-Item -LiteralPath $nativeIcon -Force
        }
    }

    & $Wails build -clean -platform $t -ldflags $LdFlags -o $outName 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $bin = Get-ChildItem -Path (Join-Path $PSScriptRoot "build\bin") -Filter "$outName*" | Select-Object -First 1
        if ($bin) {
            Copy-Item $bin.FullName (Join-Path $OutDir $bin.Name) -Force
            $size = [math]::Round($bin.Length / 1MB, 1)
            Write-Host "OK  ($($size) MB)" -ForegroundColor Green
            $ok++
        } else {
            Write-Host "OK  (binary not found in build/bin)" -ForegroundColor Yellow
            $ok++
        }
    } else {
        Write-Host "FAILED" -ForegroundColor Red
        $err++
    }
}

Write-Host ""
if ($err -eq 0) {
    Write-Host "  All $ok GUI builds succeeded." -ForegroundColor Green
} else {
    Write-Host "  $ok succeeded, $err failed." -ForegroundColor Red
    exit 1
}
Write-Host ""
