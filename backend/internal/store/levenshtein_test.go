package store

import "testing"

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"Мінск", "Мінск", 0},
		{"Мінск", "мінск", 0},
		{"Мінск", "Пінск", 1},
		{"kitten", "sitting", 3},
	}
	for _, c := range cases {
		got := levenshtein(c.a, c.b)
		if got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNormalizeForCompare(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"Мінск":     "мінск",
		"Минск":     "мінск",   // и → і
		"Менск":     "менск",
		"Магілёў":   "магілёу", // ў → у
		"Магілёу":   "магілёу",
		"  БРЭСТ ":  "брэст",
		"Сант’яга":  "сант’яга",
	}
	for in, want := range cases {
		if got := normalizeForCompare(in); got != want {
			t.Errorf("normalizeForCompare(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstLetterLower(t *testing.T) {
	cases := map[string]rune{
		"":        0,
		"  Мінск": 'м',
		"Мінск":   'м',
		"міНСК":   'м',
		"123":     0,
		"-Х":      'х',
	}
	for in, want := range cases {
		if got := firstLetterLower(in); got != want {
			t.Errorf("firstLetterLower(%q) = %q, want %q", in, got, want)
		}
	}
}
