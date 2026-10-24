package store

import (
	"database/sql"
	"time"
)

type submission struct {
	cityID   int64
	roomCode string
}

type CityStat struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
	Count  int    `json:"count"`
}

type StatsSummary struct {
	TotalCities      int        `json:"total_cities"`
	TotalSubmissions int        `json:"total_submissions"`
	UniqueCitiesUsed int        `json:"unique_cities_used"`
	NeverUsed        int        `json:"never_used"`
	LearnedCities    int        `json:"learned_cities"`
	FirstSubmission  *time.Time `json:"first_submission,omitempty"`
	LastSubmission   *time.Time `json:"last_submission,omitempty"`
}

func (s *CityStore) RecordSubmission(cityID int64, roomCode string) {
	if cityID <= 0 { return }
	select {
	case s.submissions <- submission{cityID: cityID, roomCode: roomCode}:
	default:
	}
}

func (s *CityStore) submissionWorker() {
	for sub := range s.submissions {
		_, _ = s.db.Exec(
			`INSERT INTO city_submissions (city_id, room_code) VALUES (?, ?)`,
			sub.cityID, sub.roomCode,
		)
	}
}

func (s *CityStore) TopCities(limit int) ([]CityStat, error) {
	return s.rankedCities(limit, "DESC")
}

func (s *CityStore) BottomCities(limit int) ([]CityStat, error) {
	return s.rankedCities(limit, "ASC")
}

func (s *CityStore) rankedCities(limit int, dir string) ([]CityStat, error) {
	if limit <= 0 || limit > 1000 { limit = 50 }
	q := `
SELECT c.id, c.name, c.region, COUNT(s.id) AS cnt
FROM cities c
LEFT JOIN city_submissions s ON s.city_id = c.id
GROUP BY c.id, c.name, c.region
ORDER BY cnt ` + dir + `, c.name COLLATE NOCASE ASC
LIMIT ?`
	rows, err := s.db.Query(q, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []CityStat
	for rows.Next() {
		var c CityStat
		if err := rows.Scan(&c.ID, &c.Name, &c.Region, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CityStore) CityStatByName(name string) (*CityStat, error) {
	city, err := s.FindSimilar(name)
	if err != nil { return nil, err }
	q := `
SELECT c.id, c.name, c.region, COUNT(s.id)
FROM cities c
LEFT JOIN city_submissions s ON s.city_id = c.id
WHERE c.id = ?
GROUP BY c.id, c.name, c.region`
	var out CityStat
	if err := s.db.QueryRow(q, city.ID).Scan(&out.ID, &out.Name, &out.Region, &out.Count); err != nil {
		if err == sql.ErrNoRows { return nil, ErrNotFound }
		return nil, err
	}
	return &out, nil
}

func (s *CityStore) Summary() (*StatsSummary, error) {
	var sum StatsSummary
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM cities`).Scan(&sum.TotalCities); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM city_submissions`).Scan(&sum.TotalSubmissions); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(DISTINCT city_id) FROM city_submissions`).Scan(&sum.UniqueCitiesUsed); err != nil {
		return nil, err
	}
	sum.NeverUsed = sum.TotalCities - sum.UniqueCitiesUsed
	if sum.NeverUsed < 0 { sum.NeverUsed = 0 }
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM cities WHERE source = 'learned'`).Scan(&sum.LearnedCities); err != nil {
		return nil, err
	}
	var firstStr, lastStr sql.NullString
	if err := s.db.QueryRow(
		`SELECT MIN(created_at), MAX(created_at) FROM city_submissions`,
	).Scan(&firstStr, &lastStr); err != nil {
		return nil, err
	}
	if firstStr.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", firstStr.String); err == nil {
			sum.FirstSubmission = &t
		}
	}
	if lastStr.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastStr.String); err == nil {
			sum.LastSubmission = &t
		}
	}
	return &sum, nil
}

func (s *CityStore) ResetSubmissions() error {
	if _, err := s.db.Exec(`DELETE FROM city_submissions`); err != nil {
		return err
	}
	_, _ = s.db.Exec(`DELETE FROM sqlite_sequence WHERE name = 'city_submissions'`)
	return nil
}
