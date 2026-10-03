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
}

const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
