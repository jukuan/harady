package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/models"
)

func newStatsStore(t *testing.T) *CityStore {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = conn.Close() })
	return NewCityStore(conn)
}

func recordAndWait(s *CityStore, cityID int64, room string) {
	s.RecordSubmission(cityID, room)
	time.Sleep(20 * time.Millisecond)
}

func TestTopAndBottomCities(t *testing.T) {
	s := newStatsStore(t)
	ids := map[string]int64{}
	for _, name := range []string{"Мінск", "Гомель", "Гродна"} {
		id, _ := s.Create(&models.City{Name: name})
		ids[name] = id
	}
	for i := 0; i < 3; i++ { recordAndWait(s, ids["Мінск"], "r1") }
	recordAndWait(s, ids["Гомель"], "r1")
	top, _ := s.TopCities(10)
	if len(top) == 0 || top[0].Name != "Мінск" || top[0].Count != 3 {
		t.Fatalf("top=%+v", top)
	}
	bottom, _ := s.BottomCities(10)
	if len(bottom) == 0 || bottom[0].Name != "Гродна" || bottom[0].Count != 0 {
		t.Fatalf("bottom=%+v", bottom)
	}
}

func TestSummary(t *testing.T) {
	s := newStatsStore(t)
	id1, _ := s.Create(&models.City{Name: "Мінск"})
	s.Create(&models.City{Name: "Гомель"})
	s.Create(&models.City{Name: "Гродна"})
	recordAndWait(s, id1, "r1")
	recordAndWait(s, id1, "r1")
	sum, _ := s.Summary()
	if sum.TotalCities != 3 || sum.TotalSubmissions != 2 ||
		sum.UniqueCitiesUsed != 1 || sum.NeverUsed != 2 {
		t.Errorf("sum=%+v", sum)
	}
}
