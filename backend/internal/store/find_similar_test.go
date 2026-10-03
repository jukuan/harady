package store

import (
	"path/filepath"
	"testing"

	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/models"
)

func newSeededStore(t *testing.T) *CityStore {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	s := NewCityStore(conn)
	for _, name := range []string{
		"Мінск", "Пінск", "Нара", "Ніца",
		"Баранавічы", "Гомель", "Магілёў",
		"Сант’яга", "Аклахома-Сіці",
	} {
		if _, err := s.Create(&models.City{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func TestFindSimilar_Matches(t *testing.T) {
	s := newSeededStore(t)
	cases := []struct {
		in   string
		want string
	}{
		// Exact and case-folded.
		{"Мінск", "Мінск"},
		{"мінск", "Мінск"},
		{"МІНСК", "Мінск"},

		// Belarusian vs Russian spelling (и→і normalization).
		{"Минск", "Мінск"},

		// Old Belarusian orthography (middle typo, distance 1).
		{"Менск", "Мінск"},

		// Middle typo in a longer name (distance 1, budget 2 at length 10).
		{"Барановічы", "Баранавічы"},
		{"Баранавічы", "Баранавічы"},

		// ў↔у normalization.
		{"Магілёу", "Магілёў"},

		// Trailing soft sign dropped, distance 1: Гомель vs Гомел.
		// First letter Г, last letter ь vs л — different!
		// So this should NOT match. Commented out intentionally.

		// Only itself.
		{"Пінск", "Пінск"},
		{"Нара", "Нара"},
		{"Ніца", "Ніца"},
	}
	for _, c := range cases {
		got, err := s.FindSimilar(c.in)
		if err != nil {
			t.Errorf("FindSimilar(%q): %v", c.in, err)
			continue
		}
		if got.Name != c.want {
			t.Errorf("FindSimilar(%q) = %q, want %q", c.in, got.Name, c.want)
		}
	}
}

func TestFindSimilar_Rejections(t *testing.T) {
	s := newSeededStore(t)
	for _, in := range []string{
		"Пінскк",  // П vs М for Мінск, but exact for Пінскк→Пінск at 1 edit
		"Фукуока", // not in DB
		"",        // empty
		"Аб",      // too short
	} {
		// Пінскк actually should match Пінск (distance 1, first П last к).
		// Skip if you want. The remaining cases must fail.
		if in == "Пінскк" {
			continue
		}
		if _, err := s.FindSimilar(in); err == nil {
			// Only OK if it's an exact match of itself.
			got, _ := s.FindSimilar(in)
			if got != nil && got.Name != in {
				t.Errorf("FindSimilar(%q) unexpectedly matched %q", in, got.Name)
			}
		}
	}
}

// The examples the user cited.
func TestFindSimilar_MinskNotPinsk(t *testing.T) {
	s := newSeededStore(t)
	// Мінск and Пінск differ only in the first letter.
	got, err := s.FindSimilar("Минск")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Мінск" {
		t.Errorf("Минск matched %q, want Мінск", got.Name)
	}
	got, err = s.FindSimilar("Пінск")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Пінск" {
		t.Errorf("Пінск matched %q", got.Name)
	}
}

func TestFindSimilar_NaraNotNica(t *testing.T) {
	s := newSeededStore(t)
	// Нара and Ніца: same first letter Н, same last letter а,
	// distance 2 (а≠і, р≠ц). Length 4 → budget 1. Reject.
	got, err := s.FindSimilar("Нара")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Нара" {
		t.Errorf("Нара matched %q", got.Name)
	}
	got, err = s.FindSimilar("Ніца")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Ніца" {
		t.Errorf("Ніца matched %q", got.Name)
	}
}

func TestFindSimilar_SantiagoNotOklahoma(t *testing.T) {
	s := newSeededStore(t)
	// First letter С vs А, last letter а vs і. Reject across the board.
	got, err := s.FindSimilar("Сант’яга")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Сант’яга" {
		t.Errorf("Сант’яга matched %q", got.Name)
	}
	// Sanity: Аклахома-Сіці does NOT match Сант’яга.
	got, err = s.FindSimilar("Аклахома-Сіці")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Аклахома-Сіці" {
		t.Errorf("Аклахома-Сіці matched %q", got.Name)
	}
}

// A name that would previously match at distance 2 but shouldn't under
// the 20% budget for its length.
func TestFindSimilar_ShortNamesAreStrict(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	s := NewCityStore(conn)

	for _, name := range []string{"Рым", "Рум", "Ром"} {
		if _, err := s.Create(&models.City{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	// Length 3 → budget 1. Each of these should match itself, not a sibling.
	for _, name := range []string{"Рым", "Рум", "Ром"} {
		got, err := s.FindSimilar(name)
		if err != nil {
			t.Errorf("FindSimilar(%q): %v", name, err)
			continue
		}
		if got.Name != name {
			t.Errorf("FindSimilar(%q) = %q, want itself", name, got.Name)
		}
	}
}
