#!/usr/bin/env pwsh
# Cross-compile SenPaiScanner for all supported platforms.
# Usage:  .\build.ps1
#         .\build.ps1 -Version "0.7.1"

param(
    [string]$Version = "0.7.1"
)

# Use Continue so that go's informational stderr lines don't abort the script.
$ErrorActionPreference = "Continue"

$Commit    = (git rev-parse --short HEAD 2>$null) -replace "`n",""
$BuildDate = (Get-Date -Format "yyyy-MM-dd")
$BuiltBy   = "goreleaser-local"
$MainPkg   = "./cmd/senpaiscanner"
$OutDir    = "dist"
$LdFlags   = "-s -w " +
             "-X github.com/matinsenpai/senpaiscanner/pkg/version.Version=$Version " +
             "-X github.com/matinsenpai/senpaiscanner/pkg/version.Commit=$Commit " +
             "-X github.com/matinsenpai/senpaiscanner/pkg/version.BuildDate=$BuildDate " +
             "-X github.com/matinsenpai/senpaiscanner/pkg/version.BuiltBy=$BuiltBy"

$Targets = @(
    @{ os="darwin";  arch="amd64"; out="senpaiscanner-darwin-amd64";    ext="" },
    @{ os="darwin";  arch="arm64"; out="senpaiscanner-darwin-arm64";    ext="" },
    @{ os="linux";   arch="386";   out="senpaiscanner-linux-386";       ext="" },
    @{ os="linux";   arch="amd64"; out="senpaiscanner-linux-amd64";     ext="" },
    @{ os="linux";   arch="arm64"; out="senpaiscanner-linux-arm64";     ext="" },
    @{ os="windows"; arch="386";   out="senpaiscanner-windows-386";     ext=".exe" },
    @{ os="windows"; arch="amd64"; out="senpaiscanner-windows-amd64";   ext=".exe" }
)

if (!(Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

Write-Host ""
Write-Host "  SenPaiScanner $Version  ($Commit  $BuildDate)" -ForegroundColor Cyan
Write-Host "  Building $($Targets.Count) binaries into /$OutDir ..." -ForegroundColor Cyan
Write-Host ""

$ok  = 0
$err = 0

foreach ($t in $Targets) {
    $bin = "$OutDir/$($t.out)$($t.ext)"
    $env:GOOS        = $t.os
    $env:GOARCH      = $t.arch
    $env:CGO_ENABLED = "0"

    $label = "$($t.os)/$($t.arch)".PadRight(16)
    Write-Host -NoNewline "  building $label  ->  $bin  "

    # Redirect stderr to a temp file so go's informational "downloading…"
    # lines don't produce PowerShell NativeCommandError records.
    $stderrFile = [System.IO.Path]::GetTempFileName()
    go build -trimpath -ldflags $LdFlags -o $bin $MainPkg 2>$stderrFile
    $buildCode = $LASTEXITCODE

    if ($buildCode -eq 0) {
        $size = [math]::Round((Get-Item $bin).Length / 1MB, 1)
        Write-Host "OK  ($($size) MB)" -ForegroundColor Green
        $ok++
    } else {
        Write-Host "FAILED" -ForegroundColor Red
        Get-Content $stderrFile | ForEach-Object { Write-Host "    $_" -ForegroundColor Red }
        $err++
    }
    Remove-Item $stderrFile -ErrorAction SilentlyContinue
}

# restore env
Remove-Item Env:\GOOS        -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH      -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host ""
if ($err -eq 0) {
    Write-Host "  All $ok builds succeeded." -ForegroundColor Green
} else {
    Write-Host "  $ok succeeded, $err failed." -ForegroundColor Red
    exit 1
}
Write-Host ""
