package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// Open opens (and migrates) the SQLite database at path.
// Parent directories are created automatically. Migrations are idempotent:
// each one is recorded in schema_migrations and only applied once.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir db dir: %w", err)
	}
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		path,
	)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := applyMigrations(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func applyMigrations(conn *sql.DB) error {
	if _, err := conn.Exec(migrationsTable); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}
	for _, m := range migrations {
		var applied int
		if err := conn.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, m.version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if applied > 0 {
			continue
		}
		if _, err := conn.Exec(m.sql); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", m.version, m.name, err)
		}
		if _, err := conn.Exec(
			`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`,
			m.version, m.name,
		); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		// testing.Testing() is true inside `go test`; keep test output clean.
		if !testing.Testing() {
			log.Printf("db: applied migration %d (%s)", m.version, m.name)
		}
	}
	return nil
}
