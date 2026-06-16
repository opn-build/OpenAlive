//go:build windows

package ui

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"openalive/internal/config"
	"openalive/internal/i18n"
)

// labelTr returns a declarative Label whose text is re-applied on language change.
func (w *Window) labelTr(key string) Label {
	var lbl *walk.Label
	w.addRetranslator(func() {
		if lbl != nil {
			_ = lbl.SetText(i18n.T(key))
		}
	})
	return Label{AssignTo: &lbl, Text: i18n.T(key)}
}

// groupTr returns a GroupBox whose title is re-applied on language change.
func (w *Window) groupTr(key string, layout Layout, children ...Widget) GroupBox {
	var gb *walk.GroupBox
	w.addRetranslator(func() {
		if gb != nil {
			_ = gb.SetTitle(i18n.T(key))
		}
	})
	return GroupBox{AssignTo: &gb, Title: i18n.T(key), Layout: layout, Children: children}
}

// checkTr registers a retranslator for a checkbox text and returns the CheckBox.
func (w *Window) checkTr(assign **walk.CheckBox, key string, onClick walk.EventHandler) CheckBox {
	w.addRetranslator(func() {
		if *assign != nil {
			_ = (*assign).SetText(i18n.T(key))
		}
	})
	return CheckBox{AssignTo: assign, Text: i18n.T(key), OnClicked: onClick}
}

func (w *Window) buttonTr(assign **walk.PushButton, key string, onClick walk.EventHandler) PushButton {
	w.addRetranslator(func() {
		if *assign != nil {
			_ = (*assign).SetText(i18n.T(key))
		}
	})
	return PushButton{AssignTo: assign, Text: i18n.T(key), OnClicked: onClick}
}

// ── Pages ────────────────────────────────────────────────────────────────────

func (w *Window) statusPage() TabPage {
	var page *walk.TabPage
	w.addRetranslator(func() {
		if page != nil {
			_ = page.SetTitle(i18n.T("tab.status"))
		}
	})
	return TabPage{
		AssignTo: &page,
		Title:    i18n.T("tab.status"),
		Layout:   VBox{},
		Children: []Widget{
			CustomWidget{AssignTo: &w.statusBadge, MinSize: Size{Height: 66}, MaxSize: Size{Height: 66}, Paint: w.paintBadge},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				Label{AssignTo: &w.nextEventLbl, Text: ""},
				HSpacer{},
			}},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				Label{AssignTo: &w.scheduleSummLbl, Text: ""},
				HSpacer{},
			}},
			VSpacer{},
			CustomWidget{AssignTo: &w.logWidget, MinSize: Size{Height: 90}, MaxSize: Size{Height: 90}, Paint: w.paintLog},
		},
	}
}

func (w *Window) schedulePage() TabPage {
	var page *walk.TabPage
	w.addRetranslator(func() {
		if page != nil {
			_ = page.SetTitle(i18n.T("tab.schedule"))
		}
	})
	return TabPage{
		AssignTo: &page,
		Title:    i18n.T("tab.schedule"),
		Layout:   VBox{},
		Children: []Widget{
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				w.checkTr(&w.schedEnable, "schedule.master_toggle", nil),
				HSpacer{},
			}},
			w.groupTr("schedule.work_hours", HBox{},
				w.labelTr("schedule.start"),
				LineEdit{AssignTo: &w.workStart, MaxLength: 5, MinSize: Size{Width: 50}, MaxSize: Size{Width: 70}},
				w.labelTr("schedule.end"),
				LineEdit{AssignTo: &w.workEnd, MaxLength: 5, MinSize: Size{Width: 50}, MaxSize: Size{Width: 70}},
			),
			w.groupTr("schedule.lunch_section", VBox{},
				Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
					w.checkTr(&w.lunchEnable, "schedule.lunch_toggle", nil),
					HSpacer{},
				}},
				Composite{
					Layout: HBox{},
					Children: []Widget{
						w.labelTr("schedule.start"),
						LineEdit{AssignTo: &w.lunchStart, MaxLength: 5, MinSize: Size{Width: 50}, MaxSize: Size{Width: 70}},
						w.labelTr("schedule.end"),
						LineEdit{AssignTo: &w.lunchEnd, MaxLength: 5, MinSize: Size{Width: 50}, MaxSize: Size{Width: 70}},
					},
				},
			),
			w.buttonTr(new(*walk.PushButton), "schedule.save_btn", func() { w.saveSchedule() }),
			Label{AssignTo: &w.schedMsg, Text: ""},
			VSpacer{},
		},
	}
}

func (w *Window) settingsPage() TabPage {
	var page *walk.TabPage
	w.addRetranslator(func() {
		if page != nil {
			_ = page.SetTitle(i18n.T("tab.settings"))
		}
	})

	langLabels := make([]string, len(langOptions))
	for i, o := range langOptions {
		langLabels[i] = o.label
	}

	return TabPage{
		AssignTo: &page,
		Title:    i18n.T("tab.settings"),
		Layout:   VBox{},
		Children: []Widget{
			w.groupTr("settings.mouse", HBox{},
				w.labelTr("settings.pixels"),
				NumberEdit{AssignTo: &w.pixels, Decimals: 0, MinValue: 1, MaxValue: 500, MaxSize: Size{Width: 60}},
				w.labelTr("settings.interval"),
				NumberEdit{AssignTo: &w.interval, Decimals: 0, MinValue: 5, MaxValue: 3600, MaxSize: Size{Width: 70}},
			),
			w.groupTr("settings.keyboard", HBox{},
				w.checkTr(&w.keyEnable, "settings.key_toggle", nil),
				w.labelTr("settings.key_label"),
				ComboBox{AssignTo: &w.keyCombo, Model: keyOptions},
			),
			w.groupTr("settings.language", Grid{Columns: 2},
				w.labelTr("settings.lang_label"), ComboBox{AssignTo: &w.langCombo, Model: langLabels},
			),
			w.groupTr("settings.behavior", VBox{},
				w.checkTr(&w.minimize, "settings.minimize", nil),
				Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
					w.checkTr(&w.autostart, "settings.autostart", nil),
					w.checkTr(&w.startActive, "settings.start_active", nil),
					HSpacer{},
				}},
			),
			w.buttonTr(new(*walk.PushButton), "settings.save_btn", func() { w.saveSettings() }),
			Label{AssignTo: &w.setMsg, Text: ""},
			VSpacer{},
		},
	}
}

func (w *Window) supportPage() TabPage {
	var page *walk.TabPage
	var titleLbl *walk.Label
	var bodyLnk *walk.LinkLabel
	w.addRetranslator(func() {
		if page != nil {
			_ = page.SetTitle(i18n.T("tab.support"))
		}
		if titleLbl != nil {
			_ = titleLbl.SetText(i18n.T("support.title"))
		}
		if bodyLnk != nil {
			_ = bodyLnk.SetText(i18n.T("support.body"))
		}
	})
	return TabPage{
		AssignTo: &page,
		Title:    i18n.T("tab.support"),
		Layout:   VBox{},
		Children: []Widget{
			VSpacer{},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				Label{AssignTo: &titleLbl, Text: i18n.T("support.title")},
				HSpacer{},
			}},
			LinkLabel{
				AssignTo: &bodyLnk,
				Text:     i18n.T("support.body"),
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					openURL(link.URL())
				},
			},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				ImageView{
					AssignTo: &w.paypalView,
					Image:    w.paypalBitmap,
					MinSize:  Size{Width: 96, Height: 20},
					MaxSize:  Size{Width: 96, Height: 20},
					OnMouseUp: func(x, y int, button walk.MouseButton) {
						if button == walk.LeftButton {
							openURL(paypalURL)
						}
					},
				},
				HSpacer{},
			}},
			VSpacer{},
		},
	}
}

// ── Load / Save ──────────────────────────────────────────────────────────────

func (w *Window) loadValues() {
	c := w.deps.Cfg.Snapshot()

	w.schedEnable.SetChecked(c.ScheduleEnabled)
	_ = w.workStart.SetText(c.WorkStart)
	_ = w.workEnd.SetText(c.WorkEnd)
	w.lunchEnable.SetChecked(c.LunchEnabled)
	_ = w.lunchStart.SetText(c.LunchStart)
	_ = w.lunchEnd.SetText(c.LunchEnd)

	_ = w.pixels.SetValue(float64(c.MouseMovePixels))
	_ = w.interval.SetValue(float64(c.MouseIntervalSeconds))
	w.keyEnable.SetChecked(c.KeystrokeEnabled)
	_ = w.keyCombo.SetCurrentIndex(indexOf(keyOptions, c.KeystrokeKey, 0))

	saved := c.Language
	if saved == "" {
		saved = i18n.Lang()
	}
	_ = w.langCombo.SetCurrentIndex(langIndex(saved))

	w.minimize.SetChecked(c.MinimizeOnClose)
	w.autostart.SetChecked(c.StartWithWindows)
	w.startActive.SetChecked(c.StartActive)
}

// saveSchedule validates and persists the schedule tab; returns false (and shows
// a message) if validation fails, so callers can keep the user on this tab.
func (w *Window) saveSchedule() bool {
	ws, ok1 := parseHHMM(w.workStart.Text())
	we, ok2 := parseHHMM(w.workEnd.Text())
	ls, ok3 := parseHHMM(w.lunchStart.Text())
	le, ok4 := parseHHMM(w.lunchEnd.Text())

	if !ok1 || !ok2 {
		w.showSchedMsg(i18n.T("schedule.invalid_time", map[string]string{"key": "work"}), false)
		return false
	}
	if we <= ws {
		w.showSchedMsg(i18n.T("schedule.err_order_work"), false)
		return false
	}
	if w.lunchEnable.Checked() {
		if !ok3 || !ok4 {
			w.showSchedMsg(i18n.T("schedule.invalid_time", map[string]string{"key": "lunch"}), false)
			return false
		}
		if le <= ls {
			w.showSchedMsg(i18n.T("schedule.err_order_lunch"), false)
			return false
		}
		if ls < ws || le > we {
			w.showSchedMsg(i18n.T("schedule.err_lunch_in_work"), false)
			return false
		}
	}

	w.deps.Cfg.Apply(func(c *config.Config) {
		c.ScheduleEnabled = w.schedEnable.Checked()
		c.WorkStart = w.workStart.Text()
		c.WorkEnd = w.workEnd.Text()
		c.LunchEnabled = w.lunchEnable.Checked()
		c.LunchStart = w.lunchStart.Text()
		c.LunchEnd = w.lunchEnd.Text()
	})
	_ = w.deps.Cfg.Save()
	w.showSchedMsg(i18n.T("schedule.saved"), true)
	w.refresh()
	return true
}

// scheduleDirty reports whether the schedule tab has unsaved edits.
func (w *Window) scheduleDirty() bool {
	c := w.deps.Cfg.Snapshot()
	return w.schedEnable.Checked() != c.ScheduleEnabled ||
		w.workStart.Text() != c.WorkStart ||
		w.workEnd.Text() != c.WorkEnd ||
		w.lunchEnable.Checked() != c.LunchEnabled ||
		w.lunchStart.Text() != c.LunchStart ||
		w.lunchEnd.Text() != c.LunchEnd
}

// saveSettings validates and persists the settings tab; returns false (and shows
// a message) if validation fails.
func (w *Window) saveSettings() bool {
	px := int(w.pixels.Value())
	iv := int(w.interval.Value())
	if px < 1 {
		w.showSetMsg(i18n.T("settings.err_pixels"), false)
		return false
	}
	if px > 500 {
		w.showSetMsg(i18n.T("settings.err_pixels_max"), false)
		return false
	}
	if iv < 5 {
		w.showSetMsg(i18n.T("settings.err_interval"), false)
		return false
	}
	if iv > 3600 {
		w.showSetMsg(i18n.T("settings.err_interval_max"), false)
		return false
	}
	keyIdx := w.keyCombo.CurrentIndex()
	if keyIdx < 0 || keyIdx >= len(keyOptions) {
		w.showSetMsg(i18n.T("settings.err_invalid_key"), false)
		return false
	}

	langIdx := w.langCombo.CurrentIndex()
	newLang := ""
	if langIdx >= 0 && langIdx < len(langOptions) {
		newLang = langOptions[langIdx].code
	}
	langChanged := newLang != "" && newLang != w.deps.Cfg.Snapshot().Language

	w.deps.Cfg.Apply(func(c *config.Config) {
		c.MouseMovePixels = px
		c.MouseIntervalSeconds = iv
		c.KeystrokeEnabled = w.keyEnable.Checked()
		c.KeystrokeKey = keyOptions[keyIdx]
		c.MinimizeOnClose = w.minimize.Checked()
		c.StartWithWindows = w.autostart.Checked()
		c.StartActive = w.startActive.Checked()
		c.Language = newLang
	})

	if w.deps.SetAutostart != nil {
		_ = w.deps.SetAutostart(w.autostart.Checked())
	}
	_ = w.deps.Cfg.Save()

	if langChanged {
		i18n.SetLang(newLang)
		if w.deps.OnLangChange != nil {
			w.deps.OnLangChange()
		}
		w.Retranslate()
		w.showSetMsg(i18n.T("settings.lang_applied"), true)
		return true
	}
	w.showSetMsg(i18n.T("settings.saved"), true)
	return true
}

// settingsDirty reports whether the settings tab has unsaved edits.
func (w *Window) settingsDirty() bool {
	c := w.deps.Cfg.Snapshot()
	if int(w.pixels.Value()) != c.MouseMovePixels ||
		int(w.interval.Value()) != c.MouseIntervalSeconds ||
		w.keyEnable.Checked() != c.KeystrokeEnabled ||
		w.minimize.Checked() != c.MinimizeOnClose ||
		w.autostart.Checked() != c.StartWithWindows ||
		w.startActive.Checked() != c.StartActive {
		return true
	}
	if i := w.keyCombo.CurrentIndex(); i >= 0 && i < len(keyOptions) && keyOptions[i] != c.KeystrokeKey {
		return true
	}
	// Language: compare the selected code against the saved one (empty == not set).
	if i := w.langCombo.CurrentIndex(); i >= 0 && i < len(langOptions) {
		saved := c.Language
		if saved == "" {
			saved = i18n.Lang()
		}
		if langOptions[i].code != saved {
			return true
		}
	}
	return false
}

// ── small helpers ────────────────────────────────────────────────────────────

func parseHHMM(s string) (int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func indexOf(list []string, v string, def int) int {
	for i, s := range list {
		if s == v {
			return i
		}
	}
	return def
}

func langIndex(code string) int {
	for i, o := range langOptions {
		if o.code == code {
			return i
		}
	}
	return 0
}
