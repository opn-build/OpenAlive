//go:build windows

// Package ui builds the main window (4 tabs) with walk, a Go port of ui/*.py.
// Minimizing routes to the tray with no taskbar button; closing goes to the tray
// or quits depending on the minimize_on_close setting.
package ui

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"openalive/internal/config"
	"openalive/internal/i18n"
	"openalive/internal/icons"
	"openalive/internal/schedule"
)

const paypalURL = "https://paypal.me/fcapuz"

var keyOptions = []string{"shift", "ctrl", "alt", "f13", "f14", "f15", "f16", "scrolllock", "numlock", "capslock"}

type langOption struct{ label, code string }

var langOptions = []langOption{
	{"Español", "es"}, {"English", "en"}, {"Português (Brasil)", "pt-BR"}, {"Français", "fr"},
	{"日本語", "ja"}, {"中文 (简体)", "zh-CN"}, {"한국어", "ko"}, {"Kreyòl ayisyen", "ht"},
}

// Deps wires the window to the app without an import cycle.
type Deps struct {
	Cfg          *config.Store
	Sched        *schedule.Scheduler
	OnToggle     func() bool      // toggle manual; returns new active state
	OnQuit       func()           // real exit path
	OnTrayHint   func()           // show the tray hint balloon once
	TrayActive   func() bool      // whether the tray can restore the window
	SetAutostart func(bool) error // registry HKCU Run
	OnLangChange func()           // app refreshes tray labels/tooltip
}

// Window is the application's main window.
type Window struct {
	mw        *walk.MainWindow
	deps      Deps
	allowExit bool
	done      chan struct{}

	retranslators []func()

	// tab-change unsaved-changes guard
	tabWidget    *walk.TabWidget
	prevTabIndex int
	switching    bool

	// status tab — colored badge + terminal log
	statusBadge     *walk.CustomWidget
	badgeState      string // "active" | "schedule-paused" | "inactive"
	schedStatus     string // last scheduler status key (for badge text)
	badgeFont       *walk.Font
	nextEventLbl    *walk.Label
	scheduleSummLbl *walk.Label

	// terminal log widget (3 lines: prev / current / next)
	logWidget      *walk.CustomWidget
	logPrev        string
	logCurrent     string
	logNext        string
	logFont        *walk.Font // Consolas 9
	logHeaderFont  *walk.Font // Segoe UI 8

	// schedule tab
	schedEnable *walk.CheckBox
	workStart   *walk.LineEdit
	workEnd     *walk.LineEdit
	lunchEnable *walk.CheckBox
	lunchStart  *walk.LineEdit
	lunchEnd    *walk.LineEdit
	schedMsg    *walk.Label

	// settings tab
	pixels      *walk.NumberEdit
	interval    *walk.NumberEdit
	keyEnable   *walk.CheckBox
	keyCombo    *walk.ComboBox
	langCombo   *walk.ComboBox
	minimize    *walk.CheckBox
	autostart   *walk.CheckBox
	startActive *walk.CheckBox
	setMsg      *walk.Label

	// support tab
	paypalBitmap *walk.Bitmap
	paypalView   *walk.ImageView
}

// New builds and creates the main window (initially hidden).
func New(deps Deps) (*Window, error) {
	w := &Window{deps: deps, done: make(chan struct{})}

	if img := icons.PaypalDonate(); img != nil {
		if bmp, err := walk.NewBitmapFromImage(img); err == nil {
			w.paypalBitmap = bmp
		}
	}

	if err := (MainWindow{
		AssignTo: &w.mw,
		Title:    "OpenAlive " + version(),
		Size:     Size{Width: 440, Height: 320},
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				AssignTo:              &w.tabWidget,
				OnCurrentIndexChanged: w.onTabChanged,
				Pages: []TabPage{
					w.statusPage(),
					w.schedulePage(),
					w.settingsPage(),
					w.supportPage(),
				},
			},
		},
	}).Create(); err != nil {
		return nil, err
	}

	if ic, err := walk.NewIconFromImage(icons.AppIcon()); err == nil {
		_ = w.mw.SetIcon(ic)
	}

	w.badgeState = "inactive"
	w.schedStatus = schedule.StatusManualOff
	if f, err := walk.NewFont("Segoe UI", 13, walk.FontBold); err == nil {
		w.badgeFont = f
	}
	if f, err := walk.NewFont("Consolas", 10, 0); err == nil {
		w.logFont = f
	}
	if f, err := walk.NewFont("Segoe UI", 8, 0); err == nil {
		w.logHeaderFont = f
	}
	w.loadValues()
	w.wireEvents()
	w.refresh()
	w.startRefreshLoop()
	return w, nil
}

func version() string { return appVersion }

// appVersion is set by the main package at startup.
var appVersion = "1.1.0"

// SetVersion overrides the version shown in the title bar.
func SetVersion(v string) {
	if v != "" {
		appVersion = v
	}
}

// ── Window lifecycle ─────────────────────────────────────────────────────────

// Run shows nothing by itself; it enters walk's message loop. The window starts
// hidden (lives in the tray) unless Show is called first.
func (w *Window) Run() int {
	return w.mw.Run()
}

// Show restores and focuses the window (also un-minimizes if needed).
func (w *Window) Show() {
	h := w.mw.Handle()
	win.ShowWindow(h, win.SW_RESTORE)
	w.mw.Show()
	_ = w.mw.Activate()
	win.SetForegroundWindow(h)
}

// Hide removes the window (and its taskbar button).
func (w *Window) Hide() { w.mw.Hide() }

// Exit allows the next close to proceed and terminates the message loop.
func (w *Window) Exit() {
	w.allowExit = true
	select {
	case <-w.done:
	default:
		close(w.done)
	}
	walk.App().Exit(0)
}

// SafeUpdate marshals fn onto the GUI thread (mirror of Tk after(0, fn)).
func (w *Window) SafeUpdate(fn func()) { w.mw.Synchronize(fn) }

// ── Event wiring ─────────────────────────────────────────────────────────────

func (w *Window) wireEvents() {
	// Minimize -> tray (no taskbar button). A withdrawn window is hidden, so this
	// only fires on a real minimize. Guarded by tray availability so the window
	// never becomes unrecoverable if the tray failed to start.
	w.mw.SizeChanged().Attach(func() {
		if win.IsIconic(w.mw.Handle()) && w.deps.TrayActive != nil && w.deps.TrayActive() {
			w.mw.Hide()
			w.deps.OnTrayHint()
		}
	})

	// Close -> tray or quit, per minimize_on_close.
	w.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if w.allowExit {
			return
		}
		if w.deps.Cfg.Snapshot().MinimizeOnClose {
			*canceled = true
			w.mw.Hide()
			w.deps.OnTrayHint()
		} else {
			w.deps.OnQuit()
		}
	})

	w.statusBadge.SetCursor(walk.CursorHand())
	if w.paypalView != nil {
		w.paypalView.SetCursor(walk.CursorHand())
	}

	w.statusBadge.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}
		_ = w.deps.OnToggle()
		w.refresh() // updates w.schedStatus to actual scheduler state
		w.AddLog(i18n.T("status.toggle_log", map[string]string{
			"label": i18n.T("status.label." + w.schedStatus),
		}))
	})
}

// Tab indices (declaration order in New).
const (
	tabStatus = iota
	tabSchedule
	tabSettings
	tabSupport
)

// onTabChanged prompts to save/discard unsaved edits before leaving a dirty tab,
// a port of MainWindow._on_tab_change in ui/main_window.py.
func (w *Window) onTabChanged() {
	if w.switching {
		return
	}
	newIdx := w.tabWidget.CurrentIndex()
	prev := w.prevTabIndex
	if newIdx == prev {
		return
	}
	if !w.dirtyAt(prev) {
		w.prevTabIndex = newIdx
		w.clearTabMsgs()
		return
	}

	// Revert to the dirty tab so the dialog shows in its context.
	w.setIndexGuarded(prev)

	title := i18n.T("dialog.unsaved_title")
	msg := i18n.T("dialog.unsaved_msg", map[string]string{"tab": tabTitle(prev)})
	if walk.MsgBox(w.mw, title, msg, walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == win.IDYES {
		if !w.saveAt(prev) {
			return // validation failed -> stay on the dirty tab
		}
	} else {
		w.loadValues() // discard edits
	}

	w.prevTabIndex = newIdx
	w.setIndexGuarded(newIdx)
	w.clearTabMsgs()
}

func (w *Window) setIndexGuarded(i int) {
	w.switching = true
	_ = w.tabWidget.SetCurrentIndex(i)
	w.switching = false
}

func (w *Window) dirtyAt(idx int) bool {
	switch idx {
	case tabSchedule:
		return w.scheduleDirty()
	case tabSettings:
		return w.settingsDirty()
	}
	return false
}

func (w *Window) saveAt(idx int) bool {
	switch idx {
	case tabSchedule:
		return w.saveSchedule()
	case tabSettings:
		return w.saveSettings()
	}
	return true
}

func tabTitle(idx int) string {
	switch idx {
	case tabSchedule:
		return i18n.T("tab.schedule")
	case tabSettings:
		return i18n.T("tab.settings")
	}
	return ""
}

// ── Periodic refresh ─────────────────────────────────────────────────────────


func (w *Window) startRefreshLoop() {
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-w.done:
				return
			case <-t.C:
				w.SafeUpdate(w.refresh)
			}
		}
	}()
}

func (w *Window) refresh() {
	status := w.deps.Sched.Status()
	w.UpdateStatus(status, w.deps.Sched.NextEvent(status))
}

// UpdateStatus refreshes the status tab. Must run on the GUI thread.
func (w *Window) UpdateStatus(status, nextEvent string) {
	w.refreshBadge(status)
	w.logNext = nextEvent
	_ = w.nextEventLbl.SetText(i18n.T("status.next_event", map[string]string{"event": nextEvent}))
	w.updateScheduleSummary()
	if w.logWidget != nil {
		w.logWidget.Invalidate()
	}
}

func (w *Window) refreshBadge(status string) {
	w.schedStatus = status
	switch status {
	case schedule.StatusActive:
		w.badgeState = "active"
	case schedule.StatusLunch, schedule.StatusOutsideHours:
		w.badgeState = "schedule-paused"
	default:
		w.badgeState = "inactive"
	}
	if w.statusBadge != nil {
		w.statusBadge.Invalidate()
	}
}

func (w *Window) paintBadge(canvas *walk.Canvas, _ walk.Rectangle) error {
	bounds := w.statusBadge.ClientBounds()
	var bg walk.Color
	switch w.badgeState {
	case "active":
		bg = walk.RGB(39, 174, 96)
	case "schedule-paused":
		bg = walk.RGB(230, 126, 34)
	default:
		bg = walk.RGB(192, 57, 43)
	}
	brush, err := walk.NewSolidColorBrush(bg)
	if err != nil {
		return err
	}
	defer brush.Dispose()
	if err := canvas.FillRectangle(brush, bounds); err != nil {
		return err
	}
	if w.badgeFont == nil || w.schedStatus == "" {
		return nil
	}
	text := i18n.T("status.btn." + w.schedStatus)
	return canvas.DrawText(text, w.badgeFont, walk.RGB(255, 255, 255), bounds, walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
}

func (w *Window) paintLog(canvas *walk.Canvas, _ walk.Rectangle) error {
	if w.logWidget == nil {
		return nil
	}
	b := w.logWidget.ClientBounds()

	bg, err := walk.NewSolidColorBrush(walk.RGB(10, 14, 10))
	if err != nil {
		return err
	}
	defer bg.Dispose()
	_ = canvas.FillRectangle(bg, b)

	if w.logHeaderFont == nil || w.logFont == nil {
		return nil
	}

	const lineH = 19
	const pad = 8

	_ = canvas.DrawText("activity", w.logHeaderFont, walk.RGB(77, 176, 77),
		walk.Rectangle{X: pad, Y: 5, Width: b.Width - pad*2, Height: 14},
		walk.TextSingleLine)

	// Prev (dim green)
	if w.logPrev != "" {
		_ = canvas.DrawText(w.logPrev, w.logFont, walk.RGB(40, 100, 40),
			walk.Rectangle{X: pad, Y: 22, Width: b.Width - pad*2, Height: lineH},
			walk.TextSingleLine|walk.TextVCenter)
	}

	// Current (highlight row — phosphor green)
	hlBrush, err := walk.NewSolidColorBrush(walk.RGB(15, 45, 15))
	if err == nil {
		_ = canvas.FillRectangle(hlBrush, walk.Rectangle{X: 0, Y: 41, Width: b.Width, Height: lineH})
		hlBrush.Dispose()
	}
	if cur := w.logCurrent; cur != "" {
		_ = canvas.DrawText(cur, w.logFont, walk.RGB(57, 255, 20),
			walk.Rectangle{X: pad, Y: 41, Width: b.Width - pad*2, Height: lineH},
			walk.TextSingleLine|walk.TextVCenter)
	}

	// Next event (medium green)
	if w.logNext != "" {
		_ = canvas.DrawText("→ "+w.logNext, w.logFont, walk.RGB(77, 176, 77),
			walk.Rectangle{X: pad, Y: 60, Width: b.Width - pad*2, Height: lineH},
			walk.TextSingleLine|walk.TextVCenter)
	}
	return nil
}

func (w *Window) showSchedMsg(text string, ok bool) {
	prefix, color := "✓ ", walk.RGB(30, 120, 30)
	if !ok {
		prefix, color = "✕ ", walk.RGB(160, 30, 30)
	}
	w.schedMsg.SetTextColor(color)
	_ = w.schedMsg.SetText(prefix + text)
}

func (w *Window) showSetMsg(text string, ok bool) {
	prefix, color := "✓ ", walk.RGB(30, 120, 30)
	if !ok {
		prefix, color = "✕ ", walk.RGB(160, 30, 30)
	}
	w.setMsg.SetTextColor(color)
	_ = w.setMsg.SetText(prefix + text)
}

func (w *Window) clearTabMsgs() {
	_ = w.schedMsg.SetText("")
	_ = w.setMsg.SetText("")
}

// AddLog shifts the terminal log and triggers a repaint. Must run on the GUI thread.
func (w *Window) AddLog(message string) {
	ts := time.Now().Format("15:04:05")
	w.logPrev = w.logCurrent
	w.logCurrent = fmt.Sprintf("[%s] %s", ts, message)
	if w.logWidget != nil {
		w.logWidget.Invalidate()
	}
}

func (w *Window) updateScheduleSummary() {
	c := w.deps.Cfg.Snapshot()
	if !c.ScheduleEnabled {
		_ = w.scheduleSummLbl.SetText("")
		return
	}
	summary := fmt.Sprintf("Work %s–%s", c.WorkStart, c.WorkEnd)
	if c.LunchEnabled {
		summary += fmt.Sprintf(" · Lunch %s–%s", c.LunchStart, c.LunchEnd)
	}
	_ = w.scheduleSummLbl.SetText(summary)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (w *Window) addRetranslator(fn func()) { w.retranslators = append(w.retranslators, fn) }

// Retranslate updates every text element after a language change.
func (w *Window) Retranslate() {
	for _, fn := range w.retranslators {
		fn()
	}
	_ = w.mw.SetTitle("OpenAlive " + appVersion)
	w.refresh()
	w.refreshBadge(w.schedStatus)
}

func openURL(url string) {
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// Form returns the underlying walk form (used to attach the tray icon).
func (w *Window) Form() walk.Form { return w.mw }
