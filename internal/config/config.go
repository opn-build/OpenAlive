// Package config is the thread-safe settings store, a Go port of
// config/config_manager.py + config/defaults.py.
//
// The activity engine reads settings from a worker goroutine while the UI
// writes them from the GUI thread, so every access goes through an RWMutex.
// Save() snapshots under the lock and writes the file outside it, never holding
// the lock during disk I/O (same strategy as the Python version).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config mirrors DEFAULT_CONFIG. JSON tags match the keys written by the Python
// build so a config.json from either version is interchangeable.
type Config struct {
	WorkStart            string `json:"work_start"`
	WorkEnd              string `json:"work_end"`
	LunchStart           string `json:"lunch_start"`
	LunchEnd             string `json:"lunch_end"`
	LunchEnabled         bool   `json:"lunch_enabled"`
	MouseMovePixels      int    `json:"mouse_move_pixels"`
	MouseIntervalSeconds int    `json:"mouse_interval_seconds"`
	KeystrokeEnabled     bool   `json:"keystroke_enabled"`
	KeystrokeKey         string `json:"keystroke_key"`
	MinimizeOnClose      bool   `json:"minimize_on_close"`
	StartActive          bool   `json:"start_active"`
	StartWithWindows     bool   `json:"start_with_windows"`
	ScheduleEnabled      bool   `json:"schedule_enabled"`
	Language             string `json:"language"`
	TrayHintShown        bool   `json:"tray_hint_shown"`
}

// Defaults returns the baseline config, mirroring config/defaults.py.
func Defaults() Config {
	return Config{
		WorkStart:            "08:00",
		WorkEnd:              "18:00",
		LunchStart:           "13:00",
		LunchEnd:             "14:00",
		LunchEnabled:         true,
		MouseMovePixels:      25,
		MouseIntervalSeconds: 30,
		KeystrokeEnabled:     false,
		KeystrokeKey:         "shift",
		MinimizeOnClose:      true,
		StartActive:          false,
		StartWithWindows:     false,
		ScheduleEnabled:      false,
		Language:             "",
		TrayHintShown:        false,
	}
}

// Store guards a Config with an RWMutex and persists it to disk.
type Store struct {
	mu   sync.RWMutex
	path string
	data Config
}

// Path returns config.json next to the executable, matching get_config_path()
// in the Python build (frozen: dir of the exe).
func Path() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

// Load reads the store from path, merging any on-disk values over the defaults.
// A missing or malformed file leaves the defaults intact (never an error the
// user sees), mirroring ConfigManager.load.
func Load(path string) *Store {
	data := Defaults()
	if b, err := os.ReadFile(path); err == nil {
		// Unmarshal over the defaults: keys present in the file override,
		// absent keys keep their default value.
		_ = json.Unmarshal(b, &data)
	}
	return &Store{path: path, data: data}
}

// Snapshot returns a copy of the current config for safe reads.
func (s *Store) Snapshot() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// Apply mutates the config under the write lock.
func (s *Store) Apply(fn func(*Config)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.data)
}

// Save writes the config to disk. The snapshot is taken under the lock; the file
// write happens outside it so slow disk I/O never blocks readers.
func (s *Store) Save() error {
	s.mu.RLock()
	snapshot := s.data
	s.mu.RUnlock()

	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
