package i18n

import "testing"

func TestEmbeddedTableLoads(t *testing.T) {
	if len(table) != 8 {
		t.Fatalf("expected 8 languages, got %d", len(table))
	}
	for _, lang := range order {
		if _, ok := table[lang]; !ok {
			t.Errorf("missing language %q in table", lang)
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
