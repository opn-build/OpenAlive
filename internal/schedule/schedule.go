// Package schedule decides whether the activity engine should act right now,
// a Go port of core/scheduler.py.
package schedule

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"openalive/internal/config"
	"openalive/internal/i18n"
)

// Status values, identical to the Python scheduler.
const (
	StatusActive       = "active"
	StatusManualOff    = "manual_off"
	StatusOutsideHours = "outside_hours"
	StatusLunch        = "lunch"
)

// nowFn is the clock; overridable in tests for deterministic time-of-day logic.
var nowFn = time.Now

// Scheduler resolves the current status from the config and the manual toggle.
type Scheduler struct {
	cfg          *config.Store
	mu           sync.Mutex
	manualActive bool
}

// New builds a Scheduler; the manual toggle starts from start_active.
func New(cfg *config.Store) *Scheduler {
	return &Scheduler{cfg: cfg, manualActive: cfg.Snapshot().StartActive}
}

// Status returns active | manual_off | outside_hours | lunch.
func (s *Scheduler) Status() string {
	s.mu.Lock()
	manual := s.manualActive
	s.mu.Unlock()

	c := s.cfg.Snapshot()
	if !manual {
		return StatusManualOff
	}
	if !c.ScheduleEnabled {
		return StatusActive
	}
	if !inWorkHours(c) {
		return StatusOutsideHours
	}
	if c.LunchEnabled && isLunchTime(c) {
		return StatusLunch
	}
	return StatusActive
}

// ToggleManual flips the manual on/off switch and returns the new state.
func (s *Scheduler) ToggleManual() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manualActive = !s.manualActive
	return s.manualActive
}

// ManualActive reports the current manual switch state.
func (s *Scheduler) ManualActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.manualActive
}

// NextEvent returns the localized "next event" line for the given pre-computed
// status. Accepting status avoids a second Status() call and the race condition
// that could make the result inconsistent with the badge state.
func (s *Scheduler) NextEvent(status string) string {
	c := s.cfg.Snapshot()
	now := nowFn()

	switch status {
	case StatusManualOff:
		return i18n.T("sched.no_event")

	case StatusOutsideHours:
		if ws, ok := parseTime(c.WorkStart); ok {
			nxt := atDate(now, ws)
			if !nxt.After(now) {
				nxt = nxt.AddDate(0, 0, 1)
			}
			mins := int(nxt.Sub(now).Minutes())
			h, m := mins/60, mins%60
			return i18n.T("sched.work_starts", "h", strconv.Itoa(h), "m", strconv.Itoa(m))
		}

	case StatusLunch:
		if le, ok := parseTime(c.LunchEnd); ok {
			m := minutesUntil(now, atDate(now, le))
			return i18n.T("sched.lunch_ends", "m", strconv.Itoa(m))
		}

	case StatusActive:
		if !c.ScheduleEnabled {
			return i18n.T("sched.no_schedule")
		}
		if c.LunchEnabled && beforeLunch(c) {
			if ls, ok := parseTime(c.LunchStart); ok {
				m := minutesUntil(now, atDate(now, ls))
				return i18n.T("sched.lunch_in", "m", strconv.Itoa(m))
			}
		}
		if we, ok := parseTime(c.WorkEnd); ok {
			mins := minutesUntil(now, atDate(now, we))
			h, m := mins/60, mins%60
			return i18n.T("sched.work_ends", "h", strconv.Itoa(h), "m", strconv.Itoa(m))
		}
	}
	return "—"
}

// ── helpers ────────────────────────────────────────────────────────────────

func inWorkHours(c config.Config) bool {
	ws, ok1 := parseTime(c.WorkStart)
	we, ok2 := parseTime(c.WorkEnd)
	if !ok1 || !ok2 {
		return true
	}
	now := nowMinutes()
	return ws <= now && now <= we
}

func isLunchTime(c config.Config) bool {
	ls, ok1 := parseTime(c.LunchStart)
	le, ok2 := parseTime(c.LunchEnd)
	if !ok1 || !ok2 {
		return false
	}
	now := nowMinutes()
	return ls <= now && now <= le
}

func beforeLunch(c config.Config) bool {
	ls, ok := parseTime(c.LunchStart)
	if !ok {
		return false
	}
	return nowMinutes() < ls
}

// parseTime parses "HH:MM" into minutes-since-midnight.
func parseTime(s string) (int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func nowMinutes() int {
	now := nowFn()
	return now.Hour()*60 + now.Minute()
}

// atDate returns today's date at the given minutes-since-midnight.
func atDate(now time.Time, minutes int) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), minutes/60, minutes%60, 0, 0, now.Location())
}

func minutesUntil(now, target time.Time) int {
	d := int(target.Sub(now).Minutes())
	if d < 0 {
		return 0
	}
	return d
}
