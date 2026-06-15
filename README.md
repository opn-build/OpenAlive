# OpenAlive (Go) — Windows native

Native Windows port of OpenAlive in Go. Single `.exe`, no runtime, native Win32
widgets via [`lxn/walk`](https://github.com/lxn/walk). Replaces the Python build
to cut footprint and startup time.

| Metric        | Python (PyInstaller onedir) | Go (this build)        |
| ------------- | --------------------------- | ---------------------- |
| Idle RAM      | ~60–80 MB                   | ~5–15 MB (native Win32)|
| Startup       | ~2–4 s                      | ~instant               |
| Distribution  | ~16 MB installer            | ~6 MB single exe       |
| Runtime       | bundled Python              | none                   |

The mouse/keyboard action itself uses the same Win32 `SendInput`, so it is not
"faster" at simulating — the win is footprint, startup and distribution.

## Layout

```
cmd/openalive/         entry point: single-instance guard + app wiring
internal/winput/       SendInput (windows) + no-op stub (other OS)
internal/config/       thread-safe JSON store + defaults
internal/schedule/     status + next-event logic (clock-injectable for tests)
internal/engine/       activity loop (goroutine + context cancel)
internal/tray/         walk.NotifyIcon wrapper (icon by state, menu, balloon)
internal/ui/           main window + 4 tabs (status/schedule/settings/support)
internal/i18n/         8 languages, embedded from strings.json
internal/icons/        colored status circles (pure Go image)
assets/                icon.ico, app.manifest, versioninfo.json, generated .syso
```

## Build

`lxn/walk` is **cgo-free by default**, so the Windows exe cross-builds from
Linux/WSL — no C toolchain, no Go-on-Windows required.

### From WSL / Linux (recommended for this repo)

```bash
cd go
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" \
  -o build/OpenAlive.exe ./cmd/openalive
```

Then on Windows, compile the installer without rebuilding:

```powershell
cd go
.\build.ps1 -SkipBuild
```

### From Windows (if Go is installed there)

```powershell
cd go
.\build.ps1
```

`build.ps1` builds the exe, compiles `installer\setup.iss` with Inno Setup 6,
copies the installer to `Releases\<version>\`, and asks (default **No**) whether
to publish a GitHub release.

## Regenerating the icon/manifest resource

The committed `cmd/openalive/resource_windows_amd64.syso` embeds the icon and a
Common-Controls v6 + DPI manifest (native theming). Regenerate it only if the
icon or manifest changes:

```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
cd go/assets
goversioninfo -icon=icon.ico -manifest=app.manifest \
  -o ../cmd/openalive/resource_windows_amd64.syso versioninfo.json
```

## Tests

The platform-neutral packages are tested on any OS:

```bash
cd go
go test ./...        # config, schedule, i18n
go vet ./...         # add GOOS=windows to vet the Windows-only packages
```
