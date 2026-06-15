package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	s := Load(filepath.Join(t.TempDir(), "nope.json"))
	got := s.Snapshot()
	if got != Defaults() {
		t.Fatalf("missing file should yield defaults, got %+v", got)
	}
}

func TestLoadMergesPartialOverDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	// Only two keys present; everything else must keep its default.
	if err := os.WriteFile(path, []byte(`{"mouse_move_pixels": 99, "language": "es"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Load(path).Snapshot()

	if got.MouseMovePixels != 99 {
		t.Errorf("MouseMovePixels = %d, want 99", got.MouseMovePixels)
	}
	if got.Language != "es" {
		t.Errorf("Language = %q, want es", got.Language)
	}
	// Untouched keys keep defaults.
	if got.WorkStart != "08:00" {
		t.Errorf("WorkStart = %q, want default 08:00", got.WorkStart)
	}
	if got.MouseIntervalSeconds != 10 {
		t.Errorf("MouseIntervalSeconds = %d, want default 10", got.MouseIntervalSeconds)
	}
}

func TestLoadMalformedFileFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Load(path).Snapshot(); got != Defaults() {
		t.Fatalf("malformed file should yield defaults, got %+v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	s := Load(path)
	s.Apply(func(c *Config) {
		c.MouseMovePixels = 42
		c.ScheduleEnabled = true
		c.KeystrokeKey = "f13"
	})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	got := Load(path).Snapshot()
	if got.MouseMovePixels != 42 || !got.ScheduleEnabled || got.KeystrokeKey != "f13" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}
