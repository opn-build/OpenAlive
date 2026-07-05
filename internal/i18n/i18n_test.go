package i18n

import "testing"

func TestEmbeddedTableLoads(t *testing.T) {
	defer SetLang("en")
	// Every advertised language must exist in the embedded JSON: SetLang only
	// keeps the requested code when it parsed successfully, else it falls back
	// to "en".
	for _, lang := range order {
		SetLang(lang)
		if Lang() != lang {
			t.Errorf("SetLang(%q) resolved to %q; language missing from strings.json", lang, Lang())
		}
		if tb := current.Load().(*table); len(tb.active) == 0 {
			t.Errorf("language %q loaded an empty string table", lang)
		}
	}
}

func TestSubstitutionAndFallback(t *testing.T) {
	SetLang("en")
	if got := T("status.state", "label", "Active"); got != "Status: Active" {
		t.Errorf("substitution failed: %q", got)
	}

	// Unknown language falls back to English.
	SetLang("xx")
	if Lang() != "en" {
		t.Errorf("unknown lang should fall back to en, got %q", Lang())
	}

	// Unknown key echoes the key.
	if got := T("no.such.key"); got != "no.such.key" {
		t.Errorf("unknown key should echo, got %q", got)
	}
}

func TestSpanishResolves(t *testing.T) {
	SetLang("es")
	if got := T("tab.status"); got != "Estado" {
		t.Errorf("es tab.status = %q, want Estado", got)
	}
}
