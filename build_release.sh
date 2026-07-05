#!/usr/bin/env bash
# build_release.sh — Run from the project root:
#   bash build_release.sh
#
# Prerequisites (all in WSL):
#   goversioninfo  — go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
#   Inno Setup 6   — installed on Windows at "C:\Program Files (x86)\Inno Setup 6"
#   gh CLI         — authenticated as opn-build
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO=/home/fcapuz/go-sdk/go/bin/go
GVI=/home/fcapuz/go/bin/goversioninfo
ISCC="/mnt/c/Program Files (x86)/Inno Setup 6/ISCC.exe"

# 1. Read version from main.go
VERSION=$(grep -oP 'const Version = "\K[^"]+' "$ROOT/cmd/openalive/main.go")
echo ""
echo "==> Building OpenAlive v$VERSION"
echo ""

# 2. goversioninfo — regenerate .syso with updated version metadata
echo "[1/4] goversioninfo..."
"$GVI" \
    -manifest "$ROOT/assets/app.manifest" \
    -icon "$ROOT/assets/icon.ico" \
    -o "$ROOT/cmd/openalive/resource_windows_amd64.syso" \
    "$ROOT/assets/versioninfo.json"
echo "    resource_windows_amd64.syso OK"

# 3. go build (cross-compile for Windows from WSL)
echo "[2/4] go build..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "$GO" build \
    -ldflags="-H=windowsgui -s -w" -trimpath \
    -o "$ROOT/build/OpenAlive.exe" \
    "$ROOT/cmd/openalive/"
echo "    build/OpenAlive.exe OK"

# 4. Inno Setup (ISCC.exe called from WSL via Windows path)
echo "[3/5] Inno Setup..."
"$ISCC" "$(wslpath -w "$ROOT/installer/setup.iss")"
INSTALLER="$ROOT/installer/Output/OpenAlive_Setup_v$VERSION.exe"
[ -f "$INSTALLER" ] || { echo "ERROR: installer not found: $INSTALLER"; exit 1; }
echo "    installer/Output/OpenAlive_Setup_v${VERSION}.exe OK"

# 5. Portable zip (just the standalone exe — config.json is created next to it on first run)
echo "[4/5] Portable zip..."
PORTABLE_ZIP="$ROOT/installer/Output/OpenAlive_Portable_v${VERSION}.zip"
rm -f "$PORTABLE_ZIP"
(cd "$ROOT/build" && zip -j "$PORTABLE_ZIP" OpenAlive.exe)
echo "    installer/Output/OpenAlive_Portable_v${VERSION}.zip OK"

# 6. Copy to Releases/<version>/
echo "[5/5] Copying to Releases/$VERSION/..."
RELEASE_DIR="$ROOT/Releases/$VERSION"
mkdir -p "$RELEASE_DIR"
cp "$INSTALLER" "$PORTABLE_ZIP" "$RELEASE_DIR/"
echo "    Releases/$VERSION/OpenAlive_Setup_v${VERSION}.exe OK"
echo "    Releases/$VERSION/OpenAlive_Portable_v${VERSION}.zip OK"

# 7. GitHub release (optional)
echo ""
read -rp "Create GitHub release v$VERSION on opn-build/OpenAlive? (y/N) " answer
if [[ "$answer" =~ ^[Yy]$ ]]; then
    gh release create "v$VERSION" \
        "$RELEASE_DIR/OpenAlive_Setup_v$VERSION.exe" \
        "$RELEASE_DIR/OpenAlive_Portable_v$VERSION.zip" \
        --repo opn-build/OpenAlive \
        --title "OpenAlive v$VERSION" \
        --notes "See README.md Changelog for details."
    echo "    GitHub release v$VERSION created."
else
    echo "    Skipped GitHub release."
fi

echo ""
echo "==> Done! Installer at Releases/$VERSION/OpenAlive_Setup_v${VERSION}.exe"
echo "==> Done! Portable zip at Releases/$VERSION/OpenAlive_Portable_v${VERSION}.zip"
