//go:build windows

// Command openalive is the Windows entry point: single-instance guard, then it
// wires config, scheduler, engine, tray and window together and runs walk's
// message loop. Go port of main.py + app.py.
package main

import (
	"os"
	"runtime/debug"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"openalive/internal/config"
	"openalive/internal/engine"
	"openalive/internal/i18n"
	"openalive/internal/schedule"
	"openalive/internal/tray"
	"openalive/internal/ui"
)

// Version is the app version shown in the title bar and installer.
const Version = "1.3.0"

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

func main() {
	// The app allocates almost nothing in steady state (one action every ≥5 s
	// while idling in the tray); a low GC target keeps the heap tight for a
	// negligible CPU cost.
	debug.SetGCPercent(20)

	if alreadyRunning() {
		messageBox("OpenAlive", i18n.T("app.already_running"))
		return
	}

	cfg := config.Load(config.Path())
	i18n.SetLang(cfg.Snapshot().Language)
	ui.SetVersion(Version)

	app := &App{cfg: cfg, sched: schedule.New(cfg)}
	app.run()
}

// App owns the long-lived components and the callbacks that connect them.
type App struct {
	cfg    *config.Store
	sched  *schedule.Scheduler
	engine *engine.Engine
	tray   *tray.Tray
	win    *ui.Window
}

func (a *App) run() {
	win, err := ui.New(ui.Deps{
		Cfg:          a.cfg,
		Sched:        a.sched,
		OnToggle:     a.toggleActive,
		OnQuit:       a.quit,
		OnTrayHint:   a.notifyTrayHint,
		TrayActive:   func() bool { return a.tray.IsActive() },
		SetAutostart: a.setAutostart,
		OnLangChange: a.onLangChange,
	})
	if err != nil {
		messageBox("OpenAlive", err.Error())
		return
	}
	a.win = win

	a.tray, err = tray.New(win.Form(), a.showWindow, a.toggleFromTray, a.quit)
	if err != nil {
		messageBox("OpenAlive", err.Error())
		return
	}

	a.engine = engine.New(a.cfg, a.sched, a.onEngineAction, a.onEngineStatusChange)
	a.engine.Start()
	a.pushStatus()

	// Startup is the allocation peak (walk widget creation, embedded-asset
	// decoding); return that transient memory to the OS before settling in.
	go debug.FreeOSMemory()

	win.Show() // first launch shows the window (mirrors the Python app)
	win.Run()  // blocks on walk's message loop until Exit
}

// ── Callbacks ────────────────────────────────────────────────────────────────

func (a *App) showWindow()     { a.win.Show() }
func (a *App) toggleFromTray() { a.toggleActive() }

func (a *App) toggleActive() bool {
	active := a.sched.ToggleManual()
	a.pushStatus()
	return active
}

func (a *App) onEngineAction(_, msg string) {
	a.win.SafeUpdate(func() { a.win.AddLog(msg) })
}

func (a *App) onEngineStatusChange(_ string) { a.pushStatus() }

// pushStatus updates both the tray and the window for the current status. It may
// be called from the engine goroutine, so all GUI/tray work is marshalled onto
// the GUI thread (walk's NotifyIcon and widgets are not thread-safe).
func (a *App) pushStatus() {
	status := a.sched.Status()
	visual := "inactive"
	switch status {
	case schedule.StatusActive:
		visual = "active"
	case schedule.StatusLunch, schedule.StatusOutsideHours:
		visual = "paused"
	}
	tooltip := i18n.T("tray." + status)
	next := a.sched.NextEvent(status)

	update := func() {
		if a.tray != nil {
			a.tray.Update(visual, tooltip)
		}
		if a.win != nil {
			a.win.UpdateStatus(status, next)
		}
	}
	if a.win != nil {
		a.win.SafeUpdate(update)
	} else {
		update()
	}
}

// notifyTrayHint runs whenever the window hides to the tray. That is the start
// of the long idle stretch (and when users check the Task Manager), so shrink
// the working set here; off the GUI thread to avoid a paint hitch.
func (a *App) notifyTrayHint() {
	go debug.FreeOSMemory()

	if a.cfg.Snapshot().TrayHintShown {
		return
	}
	if a.sched.Status() != schedule.StatusActive {
		return
	}
	if a.tray != nil {
		a.tray.Notify("OpenAlive", i18n.T("tray.hint"))
	}
	a.cfg.Apply(func(c *config.Config) { c.TrayHintShown = true })
	_ = a.cfg.Save()
}

func (a *App) onLangChange() {
	if a.tray != nil {
		a.tray.Retranslate()
	}
	a.pushStatus()
}

func (a *App) quit() {
	if a.engine != nil {
		a.engine.Stop()
	}
	if a.tray != nil {
		a.tray.Dispose()
	}
	if a.win != nil {
		a.win.Exit()
	}
}

// ── Autostart (HKCU Run) ─────────────────────────────────────────────────────

func (a *App) setAutostart(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return key.SetStringValue("OpenAlive", `"`+exe+`"`)
	}
	err = key.DeleteValue("OpenAlive")
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// ── Single instance ──────────────────────────────────────────────────────────

func alreadyRunning() bool {
	name, _ := windows.UTF16PtrFromString("OpenAliveSingleInstance_v1")
	_, err := windows.CreateMutex(nil, false, name)
	return err == windows.ERROR_ALREADY_EXISTS
}

func messageBox(title, text string) {
	t, _ := windows.UTF16PtrFromString(text)
	c, _ := windows.UTF16PtrFromString(title)
	const mbIconInformation = 0x40
	_, _ = windows.MessageBox(0, t, c, mbIconInformation)
}
