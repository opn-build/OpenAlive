# OpenAlive

<p align="center">
  <a href="https://opn-build.github.io/">
    <img src="https://opn-build.github.io/og-image.png" alt="OpenAlive — Keep your PC alive" width="600" />
  </a>
</p>

> Keep your PC active. Automatically.

**[Website](https://opn-build.github.io/) · [Download](https://github.com/opn-build/OpenAlive/releases) · [License](LICENSE)**

OpenAlive prevents your workstation from going idle by simulating subtle mouse movements and optional keystrokes at configurable intervals — so your status stays green and your screen stays on while you work.

No installation of runtimes required. Single `.exe` — ~4 MB installer, ~6.5 MB installed, ~5–15 MB of RAM at idle.

---

## Why OpenAlive?

Working remotely or in a monitored environment, your PC going to sleep or showing "Away" on Teams/Slack breaks your workflow. OpenAlive keeps Windows 10 and Windows 11 awake without requiring admin rights, installing drivers, or using physical mouse jigglers. It runs silently in the system tray, respects your work schedule, and collects zero telemetry.

---

## Features

- **Mouse activity simulation** — moves the cursor by a configurable number of pixels and returns it to the exact position, imperceptibly
- **Keystroke simulation** — optionally sends a configurable key (e.g. Shift) alongside the mouse movement
- **Work schedule** — activates only during your configured work hours; pauses automatically outside them
- **Lunch break** — define a lunch window inside your work hours and OpenAlive pauses itself during it
- **Manual toggle** — enable or disable at any time regardless of the schedule
- **System tray** — runs silently in the background with a colored icon that reflects the current state (Active, Inactive, Outside hours, Lunch)
- **Start with Windows** — optional autostart at login, no admin rights required
- **Minimize to tray** — closing the window keeps OpenAlive running in the background
- **Start active on open** — skip the manual toggle on launch
- **8 languages** — English, Español, Português (BR), Français, 日本語, 中文 (简体), 한국어, Kreyòl ayisyen
- **Single instance guard** — launching a second copy focuses the existing window instead

---

## Screenshots

<p align="center">
  <img src="https://opn-build.github.io/images/01-status.jpg" width="320" alt="OpenAlive Status tab — activity toggle and live log" />
  <img src="https://opn-build.github.io/images/02-schedule.jpg" width="320" alt="OpenAlive Schedule tab — work hours and lunch break configuration" />
  <img src="https://opn-build.github.io/images/03-settings.jpg" width="320" alt="OpenAlive Settings tab — movement interval, keystroke and language options" />
</p>

---

## Download

Get the latest installer — or a portable `.zip` if you'd rather not install anything — from the [Releases](https://github.com/opn-build/OpenAlive/releases) page.

Run the installer, launch OpenAlive, and it will appear in your system tray. The portable build works the same way: unzip anywhere (including a USB drive), run `OpenAlive.exe`, and it will create its `config.json` right next to itself.

---

## How it works

OpenAlive has four tabs:

| Tab | Purpose |
|-----|---------|
| **Status** | Current state, manual on/off toggle, live activity log |
| **Schedule** | Work hours (start/end) and optional lunch break window |
| **Settings** | Movement pixels, interval, keystroke, language, startup behavior |
| ☕ | Support the project |

The tray icon changes color to show the current state at a glance. Right-clicking it gives you a quick toggle and exit option.

---

## Build

<details>
<summary>Build from source (Linux / WSL)</summary>

The exe cross-compiles from Linux/WSL — no C toolchain required.

```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" \
  -o build/OpenAlive.exe ./cmd/openalive
```

### Tests

```bash
go test ./...
go vet ./...
```

</details>

---

## Changelog

### v1.3.0
- Fix: program icon embedded again in the compiled binary (`build_release.sh` was missing the `goversioninfo -icon` flag)
- Fix: `app.manifest` version was stale since v1.1.0
- UX: "next event" countdown now humanized (e.g. "8h 21m" / "30 min" instead of raw minutes)
- UX: default `Action Interval` raised to 30s; first action now waits 5s instead of firing instantly on launch
- UX: `Simulate key press` now defaults to disabled
- Cleanup: GDI fonts/icons now released alongside brushes on exit; `main.go` uses typed `x/sys/windows` calls instead of manual syscalls

### v1.2.1
- Performance: GDI brushes cached as Window fields — eliminates per-paint kernel alloc/dispose
- Performance: lock-free i18n via `atomic.Value` (no mutex on the 26-call hot path)
- Performance: `strconv.Itoa` replaces `fmt.Sprintf` for int→string in the activity loop
- Code: `i18n.T()` now takes variadic string pairs instead of `map[string]string`

### v1.2.0
- Window no longer has a maximize button
- Common Controls v6 manifest correctly embedded in binary (fixes startup crash on clean installs)

### v1.1.0
- Initial public release

---

## Privacy

OpenAlive runs entirely on your machine. It makes no network requests, sends no telemetry, and stores no personal data. All configuration is saved locally in `config.json`, next to the executable — so a portable copy stays fully self-contained.

---

## License

[GNU General Public License v3.0](LICENSE)

---

## Donate ❤️

<p align="center">
  <a href="https://www.paypal.com/paypalme/fcapuz">
    <img src="assets/paypal_donate.png" alt="Donate via PayPal" />
  </a>
</p>
