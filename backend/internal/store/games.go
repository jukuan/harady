package store

import (
	"database/sql"
	"time"
)

type GameRecord struct {
	RoomCode       string
	ChainLength    int
	PlayersCount   int
	HasBots        bool
	WinnerNickname string
	StartedAt      time.Time
	EndedAt        time.Time
}

type GameStat struct {
	ID              int64     `json:"id"`
	RoomCode        string    `json:"room_code"`
	ChainLength     int       `json:"chain_length"`
	PlayersCount    int       `json:"players_count"`
	HasBots         bool      `json:"has_bots"`
	WinnerNickname  string    `json:"winner_nickname,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	DurationSeconds int       `json:"duration_seconds"`
}

type GameSummary struct {
	TotalGames     int        `json:"total_games"`
	MaxChainLength int        `json:"max_chain_length"`
	MaxChainHuman  int        `json:"max_chain_human_only"`
	AvgChainLength float64    `json:"avg_chain_length"`
	TotalMoves     int        `json:"total_moves"`
	GamesWithBots  int        `json:"games_with_bots"`
	LastGameAt     *time.Time `json:"last_game_at,omitempty"`
}

const gameCols = `id, room_code, chain_length, players_count, has_bots, winner_nickname, started_at, ended_at`

func scanGame(row rowScanner) (*GameStat, error) {
	var (
		g    GameStat
		bots int
	)
	if err := row.Scan(&g.ID, &g.RoomCode, &g.ChainLength, &g.PlayersCount,
		&bots, &g.WinnerNickname, &g.StartedAt, &g.EndedAt); err != nil {
		return nil, err
	}
	g.HasBots = bots == 1
	g.DurationSeconds = int(g.EndedAt.Sub(g.StartedAt).Seconds())
	if g.DurationSeconds < 0 {
		g.DurationSeconds = 0
	}
	return &g, nil
}

func (s *CityStore) RecordGame(g GameRecord) error {
	bots := 0
	if g.HasBots {
		bots = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO games (room_code, chain_length, players_count, has_bots, winner_nickname, started_at, ended_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		g.RoomCode, g.ChainLength, g.PlayersCount, bots, g.WinnerNickname, g.StartedAt, g.EndedAt,
	)
	return err
}

func (s *CityStore) LongestGames(limit int) ([]*GameStat, error) {
	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT `+gameCols+` FROM games ORDER BY chain_length DESC, ended_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectGames(rows)
}

func (s *CityStore) RecentGames(limit int) ([]*GameStat, error) {
	if limit <= 0 || limit > 1000 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT `+gameCols+` FROM games ORDER BY ended_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectGames(rows)
}

func collectGames(rows *sql.Rows) ([]*GameStat, error) {
	var out []*GameStat
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *CityStore) GameSummary() (*GameSummary, error) {
	var (
		sum      GameSummary
		maxChain sql.NullInt64
		maxHuman sql.NullInt64
		avg      sql.NullFloat64
		total    sql.NullInt64
		withBots sql.NullInt64
		lastAt   sql.NullString
	)
	err := s.db.QueryRow(`
		SELECT
			COUNT(*),
			MAX(chain_length),
			COALESCE(SUM(chain_length), 0),
			AVG(chain_length),
			SUM(has_bots),
			MAX(ended_at),
			(SELECT MAX(chain_length) FROM games WHERE has_bots = 0)
		FROM games
	`).Scan(
		&sum.TotalGames, &maxChain, &total, &avg, &withBots, &lastAt, &maxHuman,
	)
	if err != nil {
		return nil, err
	}
	if maxChain.Valid { sum.MaxChainLength = int(maxChain.Int64) }
	if maxHuman.Valid { sum.MaxChainHuman = int(maxHuman.Int64) }
	if avg.Valid { sum.AvgChainLength = avg.Float64 }
	if total.Valid { sum.TotalMoves = int(total.Int64) }
	if withBots.Valid { sum.GamesWithBots = int(withBots.Int64) }
	if lastAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", lastAt.String); err == nil {
			sum.LastGameAt = &t
		}
	}
	return &sum, nil
}

func (s *CityStore) ResetGames() error {
	if _, err := s.db.Exec(`DELETE FROM games`); err != nil {
		return err
	}
	_, _ = s.db.Exec(`DELETE FROM sqlite_sequence WHERE name = 'games'`)
	return nil
}
