package store

import (
	"errors"

	"github.com/yourname/harady/backend/internal/models"
)

// SeedCities is the starter set used by `harady-cli seed`.
//
// This is the ONLY place that knows about specific city names. If you deploy
// Harady in another language, edit this slice (or replace the whole DB) —
// nothing else in the codebase cares.
//
// Later we will swap this out for a proper `seed` sub-package with many
// built-in city packs; the `Seed(cs)` signature below will stay the same.
func SeedCities() []models.City {
	return []models.City{
		{Name: "Мінск", Region: "Мінская вобласць",
			Clues: []string{"Сталіца Беларусі", "Тут знаходзіцца Няміга", "Самы вялікі горад краіны"}},
		{Name: "Гомель", Region: "Гомельская вобласць",
			Clues: []string{"Стаіць на рацэ Сож", "Другі па велічыні горад", "Паўднёвы ўсход краіны"}},
		{Name: "Магілёў", Region: "Магілёўская вобласць",
			Clues: []string{"Горад на Дняпры", "Вядомы фестывалямі", "Усходняя Беларусь"}},
		{Name: "Віцебск", Region: "Віцебская вобласць",
			Clues: []string{"Горад на Заходняй Дзвіне", "Радзіма Марка Шагала", "Славянскі базар"}},
		{Name: "Гродна", Region: "Гродзенская вобласць",
			Clues: []string{"Горад на Нёмане", "Блізка да Польшчы", "Стары замак"}},
		{Name: "Брэст", Region: "Брэсцкая вобласць",
			Clues: []string{"Заходняя брама краіны", "Знакамітая крэпасць", "На мяжы з Польшчай"}},
		{Name: "Баранавічы", Region: "Брэсцкая вобласць",
			Clues: []string{"Вялікі чыгуначны вузел", "У Брэсцкай вобласці", "Горад у цэнтры"}},
		{Name: "Бабруйск", Region: "Магілёўская вобласць",
			Clues: []string{"Горад на Бярэзіне", "Вядомы бабрамі", "У Магілёўскай вобласці"}},
		{Name: "Полацк", Region: "Віцебская вобласць",
			Clues: []string{"Самы стары горад", "На Заходняй Дзвіне", "Радзіма Скарыны"}},
		{Name: "Ліда", Region: "Гродзенская вобласць",
			Clues: []string{"Горад у Гродзенскай вобласці", "Вядомы замкам", "Недалёка ад Гродна"}},
	}
}

// Seed inserts all starter cities, skipping duplicates.
// Returns the number of rows actually inserted.
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
