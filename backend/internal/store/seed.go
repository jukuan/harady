package store

import (
	"embed"
	"errors"
	"io/fs"
	"sort"
	"strings"

	"github.com/jukuan/harady/backend/internal/models"
)

// All cities*.txt files under this package are embedded at build time. Files
// are processed in lexical order (cities01.txt, cities02.txt, …), and within
// each file, in the order lines appear. Duplicates across files are silently
// skipped by Seed.
//
// To add a new batch of cities: drop a new citiesNN.txt here and rebuild.
// No code change required.
//
//go:embed cities*.txt
var citiesFS embed.FS

// SeedCities parses every embedded cities*.txt into models.
// Lines are "Name|Region". Empty lines and lines starting with '#' are skipped.
func SeedCities() []models.City {
	names, err := fs.Glob(citiesFS, "cities*.txt")
	if err != nil {
		return nil
	}
	sort.Strings(names)

	var out []models.City
	for _, name := range names {
		raw, err := citiesFS.ReadFile(name)
		if err != nil {
			continue
		}
		out = append(out, parseCitiesFile(string(raw))...)
	}
	return out
}

func parseCitiesFile(data string) []models.City {
	var out []models.City
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, region, _ := strings.Cut(line, "|")
		name = strings.TrimSpace(name)
		region = strings.TrimSpace(region)
		if name == "" {
			continue
		}
		out = append(out, models.City{Name: name, Region: region})
	}
	return out
}

// Seed inserts all cities, skipping duplicates. Returns the number inserted.
// Safe to run repeatedly — the cities table has a UNIQUE constraint on name.
func Seed(cs *CityStore) (int, error) {
	added := 0
	for _, c := range SeedCities() {
		_, err := cs.Create(&c)
		if errors.Is(err, ErrDuplicate) {
			continue
		}
		if err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}
