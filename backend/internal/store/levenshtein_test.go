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
		{"Мінск", "мінск", 0},       // case folded
		{"Мінск", "Минск", 1},       // и→і
		{"Мінск", "Менск", 1},       // е→і
		{"Минск", "Менск", 1},
		{"Мінск", "Пінск", 1},       // only first letter differs
		// 5 edits: match М, insert а, insert г, match і, sub н→л, sub с→ё, sub к→ў.
		{"Мінск", "Магілёў", 5},
		{"kitten", "sitting", 3},
	}
	for _, c := range cases {
		got := levenshtein(c.a, c.b)
		if got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestFirstLetterLower(t *testing.T) {
	cases := map[string]rune{
		"":     0,
		"  Мінск": 'м',
		"Мінск": 'м',
		"міНСК": 'м',
		"123":   0,
		"-Х":    'х',
	}
	for in, want := range cases {
		if got := firstLetterLower(in); got != want {
			t.Errorf("firstLetterLower(%q) = %q, want %q", in, got, want)
		}
	}
}
