# OpenAlive

> Keep your PC active. Automatically.

OpenAlive prevents your workstation from going idle by simulating subtle mouse movements and optional keystrokes at configurable intervals — so your status stays green and your screen stays on while you work.

No installation of runtimes required. Single `.exe`, ~6 MB, ~5–15 MB of RAM at idle.

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
  <img src="https://opn-build.github.io/images/01-status.jpg" width="320" alt="Status tab" />
  <img src="https://opn-build.github.io/images/02-schedule.jpg" width="320" alt="Schedule tab" />
  <img src="https://opn-build.github.io/images/03-settings.jpg" width="320" alt="Settings tab" />
</p>

---

## Download

Get the latest installer from the [Releases](https://github.com/opn-build/OpenAlive/releases) page.

Run the installer, launch OpenAlive, and it will appear in your system tray.

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

## License

[GNU General Public License v3.0](LICENSE)

---

## Donate ❤️

<p align="center">
  <a href="https://www.paypal.com/paypalme/fcapuz">
    <img src="assets/paypal_donate.png" alt="Donate via PayPal" />
  </a>
</p>
