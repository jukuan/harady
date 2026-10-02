package store

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/yourname/harady/backend/internal/db"
	"github.com/yourname/harady/backend/internal/models"
)

func newTestStore(t *testing.T) *CityStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return NewCityStore(conn)
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)

	id, err := s.Create(&models.City{
		Name:   "Мінск",
		Region: "Мінская вобласць",
		Clues:  []string{"сталіца", "Няміга"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Мінск" || got.Region != "Мінская вобласць" {
		t.Errorf("unexpected city: %#v", got)
	}
	if len(got.Clues) != 2 || got.Clues[0] != "сталіца" {
		t.Errorf("unexpected clues: %#v", got.Clues)
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Create(&models.City{Name: "Мінск"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := s.Create(&models.City{Name: "Мінск"})
	if !errors.Is(err, ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestGetByNameCaseInsensitive(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Create(&models.City{Name: "Мінск"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	c, err := s.GetByName("мінск")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if c.Name != "Мінск" {
		t.Errorf("got %q", c.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetByID(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if _, err := s.GetByName("Nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	s := newTestStore(t)
	id, _ := s.Create(&models.City{Name: "A"})

	if err := s.Update(&models.City{ID: id, Name: "A", Region: "R", Clues: []string{"x"}}); err != nil {
		t.Fatalf("update: %v", err)
	}
	c, _ := s.GetByID(id)
	if c.Region != "R" || len(c.Clues) != 1 {
		t.Errorf("update not applied: %#v", c)
	}

	if err := s.Update(&models.City{ID: 999, Name: "Z"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	if err := s.Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetByID(id); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestListAndRandom(t *testing.T) {
	s := newTestStore(t)
	for _, n := range []string{"A", "B", "C"} {
		if _, err := s.Create(&models.City{Name: n}); err != nil {
			t.Fatalf("create %s: %v", n, err)
		}
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}

	if n, _ := s.Count(); n != 3 {
		t.Errorf("count = %d", n)
	}

	r, err := s.Random()
	if err != nil || r == nil {
		t.Fatalf("random: %v %v", r, err)
	}

	many, err := s.RandomMany(2)
	if err != nil {
		t.Fatalf("randomMany: %v", err)
	}
	if len(many) != 2 {
		t.Errorf("expected 2, got %d", len(many))
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	n1, err := Seed(s)
	if err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	n2, err := Seed(s)
	if err != nil {
		t.Fatalf("seed 2: %v", err)
	}
	if n1 == 0 {
		t.Fatalf("first seed inserted nothing")
	}
	if n2 != 0 {
		t.Errorf("second seed should insert 0, got %d", n2)
	}
}

func TestPickDifferent(t *testing.T) {
	if PickDifferent(nil, 0) != nil {
		t.Errorf("expected nil for empty input")
	}
	cities := []*models.City{{ID: 1}, {ID: 2}, {ID: 3}}
	got := PickDifferent(cities, 1)
	if got == nil || got.ID == 1 {
		t.Errorf("PickDifferent returned excluded city: %#v", got)
	}
}
