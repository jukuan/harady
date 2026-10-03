package game

import "testing"

func TestNormalizeCity(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"  Мінск ":    "мінск",
		"МІНСК":       "мінск",
		"New York":    "new york",
	}
	for in, want := range cases {
		if got := NormalizeCity(in); got != want {
			t.Errorf("NormalizeCity(%q) = %q, want %q", in, got, want)
		}
	}
}
