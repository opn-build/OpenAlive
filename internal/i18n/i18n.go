// Package i18n provides the localized string table for OpenAlive.
//
// The table is generated from the Python source (lang/strings.py) into
// strings.json and embedded at build time, so the Go and Python builds stay in
// sync from a single source of truth. 8 languages, ~70 keys each.
package i18n

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync/atomic"
)

//go:embed strings.json
var stringsJSON []byte

// Languages in display order. "" (system/auto) resolves to the first available
// match; we fall back to English when a key or language is missing.
var (
	table   map[string]map[string]string
	order   = []string{"es", "en", "pt-BR", "fr", "ja", "zh-CN", "ko", "ht"}
	current atomic.Value // stores string
)

func init() {
	current.Store("en")
	if err := json.Unmarshal(stringsJSON, &table); err != nil {
		// A malformed embedded table is a build-time bug; degrade to empty so
		// the app still runs (keys echo back) instead of panicking at startup.
		table = map[string]map[string]string{}
	}
}

// Available returns the supported language codes in display order.
func Available() []string { return append([]string(nil), order...) }

// SetLang selects the active language. An empty or unknown code falls back to
// English, mirroring lang.set_lang in the Python app.
func SetLang(code string) {
	if _, ok := table[code]; ok {
		current.Store(code)
		return
	}
	current.Store("en")
}

// Lang returns the active language code.
func Lang() string { return current.Load().(string) }

// T returns the localized string for key, with {placeholder} substitution from
// alternating key/value pairs in args. Falls back to English, then to the raw
// key, so the UI never shows blanks.
func T(key string, args ...string) string {
	lang := current.Load().(string)

	s, ok := table[lang][key]
	if !ok {
		if s, ok = table["en"][key]; !ok {
			s = key
		}
	}
	for i := 0; i+1 < len(args); i += 2 {
		s = strings.ReplaceAll(s, "{"+args[i]+"}", args[i+1])
	}
	return s
}
