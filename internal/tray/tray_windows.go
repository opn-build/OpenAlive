//go:build windows

// Package tray wraps walk.NotifyIcon, a Go port of tray/tray_icon.py. The native
// NotifyIcon shares walk's message loop, so there is no separate tray thread and
// no main-thread ownership conflict.
package tray

import (
	"github.com/lxn/walk"

	"openalive/internal/i18n"
	"openalive/internal/icons"
)

// Tray is the system tray icon with its context menu.
type Tray struct {
	ni    *walk.NotifyIcon
	state map[string]*walk.Icon

	showA, toggleA, exitA *walk.Action
}

// New creates the tray icon attached to form. Callbacks fire on the GUI thread.
func New(form walk.Form, onShow, onToggle, onExit func()) (*Tray, error) {
	ni, err := walk.NewNotifyIcon(form)
	if err != nil {
		return nil, err
	}

	t := &Tray{ni: ni, state: map[string]*walk.Icon{}}
	for _, st := range []string{"active", "inactive", "paused"} {
		if ic, err := walk.NewIconFromImage(icons.Circle(st)); err == nil {
			t.state[st] = ic
		}
	}
	_ = ni.SetIcon(t.state["active"])
	_ = ni.SetToolTip(i18n.T("tray.active"))

	t.showA = walk.NewAction()
	_ = t.showA.SetText(i18n.T("tray.show"))
	t.showA.Triggered().Attach(onShow)

	t.toggleA = walk.NewAction()
	_ = t.toggleA.SetText(i18n.T("tray.toggle"))
	t.toggleA.Triggered().Attach(onToggle)

	t.exitA = walk.NewAction()
	_ = t.exitA.SetText(i18n.T("tray.exit"))
	t.exitA.Triggered().Attach(onExit)

	acts := ni.ContextMenu().Actions()
	_ = acts.Add(t.showA)
	_ = acts.Add(t.toggleA)
	_ = acts.Add(walk.NewSeparatorAction())
	_ = acts.Add(t.exitA)

	// Left click restores the window (the menu's default action in Python).
	ni.MouseDown().Attach(func(_, _ int, button walk.MouseButton) {
		if button == walk.LeftButton {
			onShow()
		}
	})

	_ = ni.SetVisible(true)
	return t, nil
}

// Update sets the icon color (state: "active" | "inactive" | "paused") and tooltip.
func (t *Tray) Update(state, tooltip string) {
	if ic, ok := t.state[state]; ok {
		_ = t.ni.SetIcon(ic)
	}
	if tooltip == "" {
		tooltip = i18n.T("tray.default")
	}
	_ = t.ni.SetToolTip(tooltip)
}

// Notify shows a balloon notification.
func (t *Tray) Notify(title, message string) {
	_ = t.ni.ShowInfo(title, message)
}

// Retranslate refreshes the menu labels after a language change.
func (t *Tray) Retranslate() {
	_ = t.showA.SetText(i18n.T("tray.show"))
	_ = t.toggleA.SetText(i18n.T("tray.toggle"))
	_ = t.exitA.SetText(i18n.T("tray.exit"))
}

// IsActive reports whether the tray icon exists and can restore the window.
func (t *Tray) IsActive() bool { return t != nil && t.ni != nil }

// Dispose removes the tray icon and releases the cached state icons.
func (t *Tray) Dispose() {
	if t.ni != nil {
		_ = t.ni.Dispose()
	}
	for _, ic := range t.state {
		ic.Dispose()
	}
}
