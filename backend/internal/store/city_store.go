package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"strings"

	"github.com/yourname/harady/backend/internal/models"
)

var (
	ErrNotFound  = errors.New("city not found")
	ErrDuplicate = errors.New("city with this name already exists")
)

type CityStore struct{ db *sql.DB }

func NewCityStore(db *sql.DB) *CityStore { return &CityStore{db: db} }

const cols = `id, name, region, clues, created_at, updated_at`

type rowScanner interface{ Scan(...any) error }

func scanCity(row rowScanner) (*models.City, error) {
	var (
		c         models.City
		cluesJSON string
	)
	if err := row.Scan(&c.ID, &c.Name, &c.Region, &cluesJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	if cluesJSON != "" {
		_ = json.Unmarshal([]byte(cluesJSON), &c.Clues)
	}
	return &c, nil
}

func (s *CityStore) List() ([]*models.City, error) {
	rows, err := s.db.Query(`SELECT ` + cols + ` FROM cities ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.City
	for rows.Next() {
		c, err := scanCity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CityStore) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM cities`).Scan(&n)
	return n, err
}

func (s *CityStore) GetByName(name string) (*models.City, error) {
	// Fast path: exact match. Works for any Unicode.
	row := s.db.QueryRow(`SELECT `+cols+` FROM cities WHERE name = ? LIMIT 1`, name)
	c, err := scanCity(row)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// Fallback: case-insensitive compare in Go.
	// SQLite's built-in lower() is ASCII-only, so it can't fold Cyrillic.
	target := strings.ToLower(name)
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, c := range all {
		if strings.ToLower(c.Name) == target {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *CityStore) GetByID(id int64) (*models.City, error) {
	row := s.db.QueryRow(`SELECT `+cols+` FROM cities WHERE id = ?`, id)
	c, err := scanCity(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *CityStore) Random() (*models.City, error) {
	row := s.db.QueryRow(`SELECT ` + cols + ` FROM cities ORDER BY RANDOM() LIMIT 1`)
	c, err := scanCity(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *CityStore) RandomMany(n int) ([]*models.City, error) {
	rows, err := s.db.Query(`SELECT `+cols+` FROM cities ORDER BY RANDOM() LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.City
	for rows.Next() {
		c, err := scanCity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CityStore) Create(c *models.City) (int64, error) {
	clues, _ := json.Marshal(c.Clues)
	res, err := s.db.Exec(
		`INSERT INTO cities (name, region, clues) VALUES (?, ?, ?)`,
		c.Name, c.Region, string(clues),
	)
	if err != nil {
		if isUniqueErr(err) {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (s *CityStore) Update(c *models.City) error {
	clues, _ := json.Marshal(c.Clues)
	res, err := s.db.Exec(
		`UPDATE cities
		 SET name=?, region=?, clues=?, updated_at=CURRENT_TIMESTAMP
		 WHERE id=?`,
		c.Name, c.Region, string(clues), c.ID,
	)
	if err != nil {
		if isUniqueErr(err) {
			return ErrDuplicate
		}
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *CityStore) Delete(id int64) error {
	res, err := s.db.Exec(`DELETE FROM cities WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE")
}

// PickDifferent returns a random city that isn't `exceptID`.
// Used by guesser bots so they never accidentally say the answer.
func PickDifferent(cities []*models.City, exceptID int64) *models.City {
	if len(cities) == 0 {
		return nil
	}
	for i := 0; i < 8; i++ {
		c := cities[rand.Intn(len(cities))]
		if c.ID != exceptID {
			return c
		}
	}
	return cities[0]
}
