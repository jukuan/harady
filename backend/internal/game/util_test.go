package game

import "testing"

func TestMaskCity(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"Minsk", "_____"},
		{"New York", "___ ____"},
		{"Мінск", "_____"},
		{"Бабруйск-на-Бярэзіне", "________-__-________"},
		{"St John's", "__ ____'_"},
	}
	for _, c := range cases {
		if got := maskCity(c.in); got != c.want {
			t.Errorf("maskCity(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRandString(t *testing.T) {
	a := randString(8)
	b := randString(8)
	if len(a) != 8 || len(b) != 8 {
		t.Fatalf("wrong length: %q %q", a, b)
	}
	for _, r := range a + b {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			t.Errorf("unexpected char %q", r)
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
	// All taken: falls through to a random suffix.
	n := pickBotName(existing)
	if n == "" || seen[n] {
		t.Fatalf("overflow nickname bad: %q", n)
	}
}
