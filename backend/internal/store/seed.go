package store

import (
	_ "embed"
	"errors"
	"strings"

	"github.com/jukuan/harady/backend/internal/models"
)

//go:embed cities.txt
var seedData string

// SeedCities parses the embedded cities.txt into models.
// Lines are "Name|Region". Empty lines and lines starting with '#' are skipped.
func SeedCities() []models.City {
	var out []models.City
	for _, line := range strings.Split(seedData, "\n") {
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
