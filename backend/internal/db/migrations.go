package db

// migration is one schema change. Versions must be unique and monotonically
// increasing. Never edit an already-shipped migration; add a new one instead.
//
// Migrations are for *schema* (table definitions, indexes, columns). Seed
// data lives in internal/store/cities*.txt and is applied via `harady-cli
// seed` / `reseed`, not here.
type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_cities",
		sql: `
CREATE TABLE IF NOT EXISTS cities (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    region     TEXT NOT NULL DEFAULT '',
    clues      TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cities_name ON cities(name);`,
	},
	{
		// source = 'seed'   → came from internal/store/cities*.txt
		// source = 'learned' → crowdsourced at runtime by players (3-strikes)
		//
		// Used by `harady-cli list-learned` for review, and by the future
		// "promote to seed" workflow. Delete by source='learned' to purge.
		version: 2,
		name:    "cities_add_source_column",
		sql:     `ALTER TABLE cities ADD COLUMN source TEXT NOT NULL DEFAULT 'seed';`,
	},
	{
		version: 3,
		name:    "city_submissions",
		sql: `
CREATE TABLE IF NOT EXISTS city_submissions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    city_id     INTEGER NOT NULL,
    is_bot      INTEGER NOT NULL DEFAULT 0,
    room_code   TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_submissions_city     ON city_submissions(city_id);
CREATE INDEX IF NOT EXISTS idx_submissions_city_bot ON city_submissions(city_id, is_bot);
CREATE INDEX IF NOT EXISTS idx_submissions_created  ON city_submissions(created_at);`,
	},
	{
		version: 4,
		name:    "city_submissions_drop_is_bot",
		sql: `
DELETE FROM city_submissions WHERE is_bot = 1;
DROP INDEX IF EXISTS idx_submissions_city;
DROP INDEX IF EXISTS idx_submissions_city_bot;
DROP INDEX IF EXISTS idx_submissions_created;
ALTER TABLE city_submissions DROP COLUMN is_bot;
CREATE INDEX IF NOT EXISTS idx_submissions_city    ON city_submissions(city_id);
CREATE INDEX IF NOT EXISTS idx_submissions_created ON city_submissions(created_at);`,
	},
	{
		version: 5,
		name:    "games_chain_history",
		sql: `
CREATE TABLE IF NOT EXISTS games (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    room_code       TEXT NOT NULL DEFAULT '',
    chain_length    INTEGER NOT NULL DEFAULT 0,
    players_count   INTEGER NOT NULL DEFAULT 0,
    has_bots        INTEGER NOT NULL DEFAULT 0,
    winner_nickname TEXT NOT NULL DEFAULT '',
    started_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_games_chain ON games(chain_length DESC);
CREATE INDEX IF NOT EXISTS idx_games_ended ON games(ended_at DESC);`,
	},
}

const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
