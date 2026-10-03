package store

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/models"
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

	id, err := s.Create(&models.City{Name: "Мінск", Region: "Мінская вобласць"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.GetByID(id)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Мінск" {
		t.Errorf("got %q", got.Name)
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Create(&models.City{Name: "Мінск"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(&models.City{Name: "Мінск"}); !errors.Is(err, ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestGetByNameCaseInsensitive(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Create(&models.City{Name: "Мінск"}); err != nil {
		t.Fatal(err)
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
}

func TestExists(t *testing.T) {
	s := newTestStore(t)
	_, _ = s.Create(&models.City{Name: "Мінск"})

	ok, _ := s.Exists("Мінск")
	if !ok {
		t.Error("exact should be true")
	}
	ok, _ = s.Exists("мінск")
	if !ok {
		t.Error("case-insensitive should be true")
	}
	ok, _ = s.Exists("Гомель")
	if ok {
		t.Error("absent should be false")
	}
	ok, _ = s.Exists("")
	if ok {
		t.Error("empty should be false")
	}
}

func TestFindByFirstLetter(t *testing.T) {
	s := newTestStore(t)
	for _, n := range []string{"Мінск", "Магілёў", "Маладзечна", "Гомель", "Гродна"} {
		if _, err := s.Create(&models.City{Name: n}); err != nil {
			t.Fatal(err)
		}
	}

	got, err := s.FindByFirstLetter('М', nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 М-cities, got %d: %v", len(got), got)
	}
	for _, n := range got {
		if n == "Гомель" || n == "Гродна" {
			t.Errorf("wrong city: %q", n)
		}
	}

	got, _ = s.FindByFirstLetter('М', []string{"мінск"}, 10)
	if len(got) != 2 {
		t.Fatalf("expected 2 after exclusion, got %d: %v", len(got), got)
	}
	for _, n := range got {
		if n == "Мінск" {
			t.Errorf("excluded city returned: %q", n)
		}
	}

	got, _ = s.FindByFirstLetter('Я', nil, 10)
	if len(got) != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	s := newTestStore(t)
	id, _ := s.Create(&models.City{Name: "A"})
	if err := s.Update(&models.City{ID: id, Name: "A", Region: "R"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(&models.City{ID: 999, Name: "Z"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if err := s.Delete(id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetByID(id); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	n1, _ := Seed(s)
	n2, _ := Seed(s)
	if n1 == 0 {
		t.Fatal("first seed inserted nothing")
	}
	if n2 != 0 {
		t.Errorf("second seed should insert 0, got %d", n2)
	}
}

func TestPickDifferent(t *testing.T) {
	if PickDifferent(nil, 0) != nil {
		t.Error("expected nil for empty input")
	}
	cities := []*models.City{{ID: 1}, {ID: 2}, {ID: 3}}
	got := PickDifferent(cities, 1)
	if got == nil || got.ID == 1 {
		t.Errorf("got %#v", got)
	}
}
