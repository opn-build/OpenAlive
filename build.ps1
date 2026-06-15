# build.ps1 — Build the Go (Windows-only) OpenAlive and its installer.
# Run from the go\ directory:  .\build.ps1
#
#   .\build.ps1               Build exe (needs Go on PATH) + installer
#   .\build.ps1 -SkipBuild    Skip go build (use an exe already built from WSL)
#
# Cross-building from WSL (no Go on Windows needed) — run in WSL, then -SkipBuild:
#   cd go && GOOS=windows GOARCH=amd64 \
#     ~/go-sdk/go/bin/go build -ldflags "-H windowsgui -s -w" -o build/OpenAlive.exe ./cmd/openalive

param(
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot

# ── Read version from the Go source (single source of truth) ────────────────
$mainGo = Get-Content (Join-Path $Root "cmd\openalive\main.go") -Raw
if ($mainGo -match 'const Version = "([^"]+)"') {
    $Version = $Matches[1]
} else {
    Write-Error "Could not read Version from cmd\openalive\main.go"; exit 1
}
Write-Host "`n==> Building OpenAlive (Go) v$Version`n" -ForegroundColor Cyan

$BuildDir = Join-Path $Root "build"
$Exe      = Join-Path $BuildDir "OpenAlive.exe"
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

# ── 1. go build ──────────────────────────────────────────────────────────────
if ($SkipBuild) {
    Write-Host "[1/4] go build SKIPPED (-SkipBuild)" -ForegroundColor Yellow
    if (-not (Test-Path $Exe)) { Write-Error "build\OpenAlive.exe not found. Build it from WSL first."; exit 1 }
} else {
    Write-Host "[1/4] go build..." -ForegroundColor Yellow
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) {
        Write-Error "Go not found on PATH. Install Go, or build the exe from WSL and re-run with -SkipBuild."
        exit 1
    }
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    Push-Location $Root
    & go build -ldflags "-H windowsgui -s -w" -o $Exe ./cmd/openalive
    $code = $LASTEXITCODE
    Pop-Location
    if ($code -ne 0) { Write-Error "go build failed"; exit 1 }
}
Write-Host "    build\OpenAlive.exe OK ($([math]::Round((Get-Item $Exe).Length/1MB,1)) MB)" -ForegroundColor Green

# ── 2. Inno Setup ────────────────────────────────────────────────────────────
Write-Host "[2/4] Inno Setup..." -ForegroundColor Yellow
$Iscc = @(
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $Iscc) { Write-Error "Inno Setup 6 not found. Install it or add ISCC.exe to PATH"; exit 1 }

& $Iscc (Join-Path $Root "installer\setup.iss")
if ($LASTEXITCODE -ne 0) { Write-Error "Inno Setup compilation failed"; exit 1 }

$Installer = Join-Path $Root "installer\Output\OpenAlive_Setup_v$Version.exe"
if (-not (Test-Path $Installer)) { Write-Error "Expected installer not found: $Installer"; exit 1 }
Write-Host "    installer\Output\OpenAlive_Setup_v$Version.exe OK" -ForegroundColor Green

# ── 3. Copy to Releases\<version>\ ──────────────────────────────────────────
Write-Host "[3/4] Copying to Releases\$Version\..." -ForegroundColor Yellow
$ReleaseDir = Join-Path $Root "Releases\$Version"
New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null
Copy-Item $Installer $ReleaseDir -Force
Write-Host "    Releases\$Version\OpenAlive_Setup_v$Version.exe OK" -ForegroundColor Green

# ── 4. Optional GitHub release (asks; default No) ───────────────────────────
Write-Host "[4/4] GitHub release (optional)..." -ForegroundColor Yellow
$ReleaseRepo = "opn-build/OpenAlive"
$answer = Read-Host "    Create GitHub release v$Version on $ReleaseRepo and upload the installer? (y/N)"
if ($answer -match '^(y|Y)') {
    $gh = Get-Command gh -ErrorAction SilentlyContinue
    if (-not $gh) {
        Write-Host "    gh CLI not found - skipped. Install GitHub CLI or upload manually." -ForegroundColor Yellow
    } else {
        & gh release create "v$Version" $Installer --repo $ReleaseRepo --title "OpenAlive v$Version" --notes "See CHANGELOG.md for details."
        if ($LASTEXITCODE -ne 0) { Write-Error "GitHub release failed"; exit 1 }
        Write-Host "    GitHub release v$Version created." -ForegroundColor Green
    }
} else {
    Write-Host "    Skipped GitHub release." -ForegroundColor Cyan
}

Write-Host ""
Write-Host "==> Done! Installer at Releases\$Version\OpenAlive_Setup_v$Version.exe" -ForegroundColor Cyan
