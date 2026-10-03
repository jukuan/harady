package store

import (
	"strings"
	"testing"
)

func TestSeedCitiesLoadsEveryFile(t *testing.T) {
	cities := SeedCities()
	if len(cities) < 200 {
		t.Fatalf("expected at least 200 cities, got %d", len(cities))
	}

	// Every city we care about must be present with a non-empty name.
	// Region is optional but seeded entries should have one.
	missing := 0
	for _, c := range cities {
		if c.Name == "" {
			missing++
		}
	}
	if missing > 0 {
		t.Errorf("%d cities have empty names", missing)
	}
}

func TestSeedCitiesCoversUandI(t *testing.T) {
	cities := SeedCities()

	byLetter := map[rune]int{}
	for _, c := range cities {
		for _, r := range c.Name {
			if r == ' ' || r == '-' {
				continue
			}
			byLetter[r]++
			break
		}
	}

	// After the ў→у / й→і rule change these two letters matter for gameplay.
	for _, letter := range []rune{'У', 'І'} {
		if byLetter[letter] < 5 {
			t.Errorf("expected >= 5 cities starting with %q, got %d", letter, byLetter[letter])
		}
	}
}

func TestSeedCitiesNoBlankRegions(t *testing.T) {
	for _, c := range SeedCities() {
		if strings.TrimSpace(c.Name) == "" {
			t.Errorf("blank name in a seed entry: %+v", c)
		}
	}
}
