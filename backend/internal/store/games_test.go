package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jukuan/harady/backend/internal/db"
)

func newGamesStore(t *testing.T) *CityStore {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return NewCityStore(conn)
}

func TestRecordGameAndLongest(t *testing.T) {
	s := newGamesStore(t)
	now := time.Now()
	for _, g := range []GameRecord{
		{RoomCode: "a", ChainLength: 12, PlayersCount: 3, WinnerNickname: "Алесь", StartedAt: now.Add(-10 * time.Minute), EndedAt: now.Add(-5 * time.Minute)},
		{RoomCode: "b", ChainLength: 41, PlayersCount: 4, HasBots: true, WinnerNickname: "Ганна", StartedAt: now.Add(-20 * time.Minute), EndedAt: now.Add(-10 * time.Minute)},
		{RoomCode: "c", ChainLength: 8, PlayersCount: 2, StartedAt: now.Add(-3 * time.Minute), EndedAt: now.Add(-2 * time.Minute)},
	} {
		if err := s.RecordGame(g); err != nil {
			t.Fatal(err)
		}
	}
	longest, err := s.LongestGames(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(longest) != 3 || longest[0].ChainLength != 41 {
		t.Fatalf("longest=%+v", longest)
	}
	recent, err := s.RecentGames(10)
	if err != nil {
		t.Fatal(err)
	}
	if recent[0].RoomCode != "c" {
		t.Errorf("recent[0]=%+v", recent[0])
	}
}

func TestGameSummary(t *testing.T) {
	s := newGamesStore(t)
	now := time.Now()
	sum, _ := s.GameSummary()
	if sum.TotalGames != 0 {
		t.Errorf("empty: %+v", sum)
	}
	for _, g := range []GameRecord{
		{ChainLength: 10, StartedAt: now, EndedAt: now},
		{ChainLength: 25, HasBots: true, StartedAt: now, EndedAt: now},
		{ChainLength: 15, StartedAt: now, EndedAt: now},
	} {
		if err := s.RecordGame(g); err != nil {
			t.Fatal(err)
		}
	}
	sum, _ = s.GameSummary()
	if sum.TotalGames != 3 || sum.MaxChainLength != 25 || sum.MaxChainHuman != 15 {
		t.Errorf("sum=%+v", sum)
	}
	if sum.TotalMoves != 50 || sum.GamesWithBots != 1 {
		t.Errorf("moves=%d bots=%d", sum.TotalMoves, sum.GamesWithBots)
	}
}
