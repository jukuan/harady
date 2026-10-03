package game

import "testing"

func TestFirstLetter(t *testing.T) {
	cases := map[string]rune{
		"":         0,
		"Minsk":    'M',
		"мінск":    'М',
		"  minsk":  'M',
		"-Мінск":   'М',
		"Бабруйск": 'Б',
	}
	for in, want := range cases {
		if got := FirstLetter(in); got != want {
			t.Errorf("FirstLetter(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLastMeaningfulLetter(t *testing.T) {
	cases := map[string]rune{
		"":                     0,
		"Мінск":                'К',
		"Віцебск":              'К',
		"Брэст":                'Т',
		"Гродна":               'А',
		"Гомель":               'Л', // ь skipped → previous letter
		"Магілёў":              'У', // ў → у
		"Кіеў":                 'У', // ў → у
		"Сіднэй":               'І', // й → і
		"Баранавічы":           'Ч', // ы skipped → previous letter
		"Йорк":                 'К',
		"St John's":            'S',
		"Бабруйск-на-Бярэзіне": 'Е',
	}
	for in, want := range cases {
		if got := LastMeaningfulLetter(in); got != want {
			t.Errorf("LastMeaningfulLetter(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPickBotNameUnique(t *testing.T) {
	existing := map[string]*Player{}
	seen := map[string]bool{}
	for i := 0; i < len(botNicknames); i++ {
		n := pickBotName(existing)
		if seen[n] {
			t.Fatalf("duplicate name: %q", n)
		}
		seen[n] = true
		existing["k"+n] = &Player{Nickname: n}
	}
	n := pickBotName(existing)
	if n == "" || seen[n] {
		t.Fatalf("overflow nickname bad: %q", n)
	}
}
