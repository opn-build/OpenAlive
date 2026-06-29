# build_release.ps1 — Run from the project root:
#   cd C:\dev\activity-keeper\go
#   .\build_release.ps1
#
# Prerequisites:
#   goversioninfo  — go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
#   Inno Setup 6   — https://jrsoftware.org/isinfo.php
#   gh CLI         — https://cli.github.com  (optional, for GitHub release)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot

# ── 1. Read version from main.go ─────────────────────────────────────────────
$match = Select-String -Path "$Root\cmd\openalive\main.go" -Pattern 'const Version = "(.+)"'
$Version = $match.Matches[0].Groups[1].Value
if (-not $Version) { Write-Error "Could not read Version from cmd\openalive\main.go"; exit 1 }
Write-Host "`n==> Building OpenAlive v$Version`n" -ForegroundColor Cyan

# ── 2. goversioninfo — regenerate .syso with updated version metadata ─────────
Write-Host "[1/4] goversioninfo..." -ForegroundColor Yellow
$gvi = Get-Command goversioninfo -ErrorAction SilentlyContinue
if (-not $gvi) {
    Write-Error "goversioninfo not found. Install: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest"
    exit 1
}
& goversioninfo `
    -manifest "$Root\assets\app.manifest" `
    -o "$Root\cmd\openalive\resource_windows_amd64.syso" `
    "$Root\assets\versioninfo.json"
if ($LASTEXITCODE -ne 0) { Write-Error "goversioninfo failed"; exit 1 }
Write-Host "    resource_windows_amd64.syso OK" -ForegroundColor Green

# ── 3. go build ──────────────────────────────────────────────────────────────
Write-Host "[2/4] go build..." -ForegroundColor Yellow
$proc = Get-Process -Name "OpenAlive" -ErrorAction SilentlyContinue
if ($proc) {
    Write-Host "    Stopping running OpenAlive..." -ForegroundColor Yellow
    $proc | Stop-Process -Force
    Start-Sleep -Seconds 1
}
& go build -ldflags "-H windowsgui -s -w" -trimpath `
    -o "$Root\build\OpenAlive.exe" `
    "$Root\cmd\openalive"
if ($LASTEXITCODE -ne 0) { Write-Error "go build failed"; exit 1 }
Write-Host "    build\OpenAlive.exe OK" -ForegroundColor Green

# ── 4. Inno Setup ────────────────────────────────────────────────────────────
Write-Host "[3/4] Inno Setup..." -ForegroundColor Yellow
$Iscc = @(
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe"
) | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $Iscc) {
    Write-Error "Inno Setup 6 not found. Install from https://jrsoftware.org/isinfo.php"
    exit 1
}
& $Iscc "$Root\installer\setup.iss"
if ($LASTEXITCODE -ne 0) { Write-Error "Inno Setup compilation failed"; exit 1 }
$Installer = "$Root\installer\Output\OpenAlive_Setup_v$Version.exe"
if (-not (Test-Path $Installer)) {
    Write-Error "Expected installer not found: $Installer"
    exit 1
}
Write-Host "    installer\Output\OpenAlive_Setup_v$Version.exe OK" -ForegroundColor Green

# ── 5. Copy to Releases\<version>\ ──────────────────────────────────────────
Write-Host "[4/4] Copying to Releases\$Version\..." -ForegroundColor Yellow
$ReleaseDir = "$Root\Releases\$Version"
New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null
Copy-Item $Installer $ReleaseDir -Force
Write-Host "    Releases\$Version\OpenAlive_Setup_v$Version.exe OK" -ForegroundColor Green

# ── 6. GitHub release (optional) ─────────────────────────────────────────────
$answer = Read-Host "`nCreate GitHub release v$Version on opn-build/OpenAlive? (y/N)"
if ($answer -match '^(y|Y)') {
    $gh = Get-Command gh -ErrorAction SilentlyContinue
    if (-not $gh) {
        Write-Host "    gh CLI not found — upload manually from Releases\$Version\" -ForegroundColor Yellow
    } else {
        Write-Host "    Creating GitHub release v$Version..." -ForegroundColor Yellow
        & gh release create "v$Version" "$ReleaseDir\OpenAlive_Setup_v$Version.exe" `
            --repo opn-build/OpenAlive `
            --title "OpenAlive v$Version" `
            --notes "See README.md Changelog for details."
        if ($LASTEXITCODE -ne 0) { Write-Error "GitHub release failed"; exit 1 }
        Write-Host "    GitHub release v$Version created." -ForegroundColor Green
    }
} else {
    Write-Host "    Skipped GitHub release." -ForegroundColor Cyan
}

Write-Host ""
Write-Host "==> Done! Installer at Releases\$Version\OpenAlive_Setup_v$Version.exe" -ForegroundColor Cyan
