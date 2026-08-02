#!/usr/bin/env pwsh
# Build the SenPaiScanner desktop GUI (Wails v2) for all supported platforms.
# Usage:  .\build_gui.ps1              # build for this machine only
#         .\build_gui.ps1 -All        # cross-compile all desktop targets
#         .\build_gui.ps1 -Platform windows/amd64

param(
    [switch]$All,
    [string]$Platform = ""
)

$ErrorActionPreference = "Continue"

$Wails = "C:\Users\khmja\go\bin\wails.exe"
if (!(Test-Path $Wails)) {
    $Wails = (Get-Command wails -ErrorAction SilentlyContinue).Source
}
if (!$Wails) {
    Write-Host "wails CLI not found. Install it with:  go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Red
    exit 1
}

$DesktopDir = Join-Path $PSScriptRoot "desktop"
$OutDir     = Join-Path $PSScriptRoot "dist"
if (!(Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

if ($Platform) {
    $targets = @($Platform)
} elseif ($All) {
    $targets = @(
        "windows/amd64",
        "linux/amd64",
        "linux/arm64",
        "darwin/amd64",
        "darwin/arm64"
    )
} else {
    $targets = @((& $Wails doctor 2>$null | Select-String -Pattern "OS\s*:\s*(.+)").Matches.Groups[1].Value.Trim())
    if (!$targets) { $targets = @("windows/amd64") }
}

Write-Host ""
Write-Host "  SenPaiScanner GUI  —  wails $(& $Wails version 2>$null | Select-String -Pattern "v?\d+\.\d+\.\d+" | Select-Object -First 1)" -ForegroundColor Cyan
Write-Host "  Building $($targets.Count) target(s) ..." -ForegroundColor Cyan
Write-Host ""

$ok  = 0
$err = 0

foreach ($t in $targets) {
    $outName = "senpaiscanner-gui-" + ($t -replace "/", "-")
    if ($t -like "windows/*" -and !$outName.EndsWith(".exe")) { $outName += ".exe" }
    Write-Host -NoNewline "  building $($t.PadRight(20))  ->  dist/$outName  "

    & $Wails build -clean -platform $t -o $outName 2>&1 | Out-Null
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
