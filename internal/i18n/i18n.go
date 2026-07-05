// Package i18n provides the localized string table for OpenAlive.
//
// The table is generated from the Python source (lang/strings.py) into
// strings.json and embedded at build time, so the Go and Python builds stay in
// sync from a single source of truth. 8 languages, ~70 keys each.
//
// Only the active language and the English fallback are kept resident; the
// other languages are dropped after parsing and re-read from the embedded JSON
// on SetLang (32 KB parse, negligible for a user-initiated language switch).
package i18n

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync/atomic"
)

//go:embed strings.json
var stringsJSON []byte

// order lists the supported languages in display order. "" (system/auto)
// resolves to English, as does any unknown code.
var order = []string{"es", "en", "pt-BR", "fr", "ja", "zh-CN", "ko", "ht"}

// table holds the resident string maps for one selected language.
type table struct {
	lang    string
	active  map[string]string
	english map[string]string // fallback; same map as active when lang == "en"
}

// current stores *table.
var current atomic.Value

func init() { load("en") }

// load parses the embedded JSON and retains only code's strings plus the
// English fallback. A malformed embedded table is a build-time bug; degrade to
// empty so the app still runs (keys echo back) instead of panicking at startup.
func load(code string) {
	var all map[string]map[string]string
	if err := json.Unmarshal(stringsJSON, &all); err != nil {
		all = map[string]map[string]string{}
	}
	if _, ok := all[code]; !ok {
		code = "en"
	}
	current.Store(&table{lang: code, active: all[code], english: all["en"]})
}

// Available returns the supported language codes in display order.
func Available() []string { return append([]string(nil), order...) }

// SetLang selects the active language. An empty or unknown code falls back to
// English, mirroring lang.set_lang in the Python app.
func SetLang(code string) { load(code) }

// Lang returns the active language code.
func Lang() string { return current.Load().(*table).lang }

// T returns the localized string for key, with {placeholder} substitution from
// alternating key/value pairs in args. Falls back to English, then to the raw
// key, so the UI never shows blanks.
func T(key string, args ...string) string {
	t := current.Load().(*table)

	s, ok := t.active[key]
	if !ok {
		if s, ok = t.english[key]; !ok {
			s = key
		}
	}
	for i := 0; i+1 < len(args); i += 2 {
		s = strings.ReplaceAll(s, "{"+args[i]+"}", args[i+1])
	}
	return s
}
