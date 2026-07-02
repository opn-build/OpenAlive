package schedule

import (
	"path/filepath"
	"testing"
	"time"

	"openalive/internal/config"
)

// newStore builds a config store with the given overrides applied on top of
// defaults, persisted to a temp path.
func newStore(t *testing.T, apply func(*config.Config)) *config.Store {
	t.Helper()
	s := config.Load(filepath.Join(t.TempDir(), "config.json"))
	s.Apply(apply)
	return s
}

// fixedClock pins nowFn to a given HH:MM today and restores it after the test.
func fixedClock(t *testing.T, hh, mm int) {
	t.Helper()
	orig := nowFn
	t.Cleanup(func() { nowFn = orig })
	n := time.Now()
	nowFn = func() time.Time {
		return time.Date(n.Year(), n.Month(), n.Day(), hh, mm, 0, 0, n.Location())
	}
}

func TestStatusManualOff(t *testing.T) {
	s := New(newStore(t, func(c *config.Config) { c.StartActive = false }))
	if got := s.Status(); got != StatusManualOff {
		t.Fatalf("Status = %q, want manual_off", got)
	}
}

func TestStatusActiveWhenScheduleDisabled(t *testing.T) {
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = false
	}))
	if got := s.Status(); got != StatusActive {
		t.Fatalf("Status = %q, want active", got)
	}
}

func TestStatusActiveInsideWorkHours(t *testing.T) {
	fixedClock(t, 10, 0) // 10:00, within 08:00–18:00, before lunch
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
	}))
	if got := s.Status(); got != StatusActive {
		t.Fatalf("Status = %q, want active", got)
	}
}

func TestStatusOutsideWorkHoursBefore(t *testing.T) {
	fixedClock(t, 6, 30) // before 08:00
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
	}))
	if got := s.Status(); got != StatusOutsideHours {
		t.Fatalf("Status = %q, want outside_hours", got)
	}
}

func TestStatusOutsideWorkHoursAfter(t *testing.T) {
	fixedClock(t, 19, 0) // after 18:00
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
	}))
	if got := s.Status(); got != StatusOutsideHours {
		t.Fatalf("Status = %q, want outside_hours", got)
	}
}

func TestStatusLunch(t *testing.T) {
	fixedClock(t, 13, 30) // within 13:00–14:00 lunch
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
		c.LunchEnabled = true
	}))
	if got := s.Status(); got != StatusLunch {
		t.Fatalf("Status = %q, want lunch", got)
	}
}

func TestLunchDisabledStaysActive(t *testing.T) {
	fixedClock(t, 13, 30) // lunch time, but lunch disabled
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
		c.LunchEnabled = false
	}))
	if got := s.Status(); got != StatusActive {
		t.Fatalf("Status = %q, want active (lunch disabled)", got)
	}
}

// Boundary: exactly at work_start and work_end are inclusive (ws <= now <= we).
func TestWorkHourBoundariesInclusive(t *testing.T) {
	for _, hm := range [][2]int{{8, 0}, {18, 0}} {
		fixedClock(t, hm[0], hm[1])
		s := New(newStore(t, func(c *config.Config) {
			c.StartActive = true
			c.ScheduleEnabled = true
			c.LunchEnabled = false
		}))
		if got := s.Status(); got != StatusActive {
			t.Fatalf("at %02d:%02d Status = %q, want active (inclusive boundary)", hm[0], hm[1], got)
		}
	}
}

func TestToggleManual(t *testing.T) {
	s := New(newStore(t, func(c *config.Config) { c.StartActive = false }))
	if s.ToggleManual() != true || s.Status() != StatusActive {
		t.Fatal("toggle should turn manual on -> active")
	}
	if s.ToggleManual() != false || s.Status() != StatusManualOff {
		t.Fatal("toggle should turn manual off -> manual_off")
	}
}

func TestNextEventLunchIn(t *testing.T) {
	fixedClock(t, 12, 30) // 30 min before lunch at 13:00
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
		c.LunchEnabled = true
	}))
	if got, want := s.NextEvent(s.Status()), "Lunch in 30 min"; got != want {
		t.Fatalf("NextEvent = %q, want %q", got, want)
	}
}

func TestNextEventLunchInOverAnHour(t *testing.T) {
	fixedClock(t, 8, 39) // 4h21m before lunch at 13:00, within work hours (08:00-18:00)
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
		c.LunchEnabled = true
	}))
	if got, want := s.NextEvent(s.Status()), "Lunch in 4h 21m"; got != want {
		t.Fatalf("NextEvent = %q, want %q", got, want)
	}
}

func TestNextEventLunchEnds(t *testing.T) {
	fixedClock(t, 13, 30) // 30 min before lunch ends at 14:00
	s := New(newStore(t, func(c *config.Config) {
		c.StartActive = true
		c.ScheduleEnabled = true
		c.LunchEnabled = true
	}))
	if got, want := s.NextEvent(s.Status()), "Lunch ends in 30 min"; got != want {
		t.Fatalf("NextEvent = %q, want %q", got, want)
	}
}

func TestParseTimeBoundaries(t *testing.T) {
	cases := map[string]bool{
		"08:00": true, "23:59": true, "00:00": true,
		"24:00": false, "08:60": false, "bad": false, "8": false,
	}
	for in, ok := range cases {
		if _, got := parseTime(in); got != ok {
			t.Errorf("parseTime(%q) ok=%v, want %v", in, got, ok)
		}
	}
}
