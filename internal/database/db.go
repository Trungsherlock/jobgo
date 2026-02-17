package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_key=ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign key: %w", err)
	}

	return &DB{db}, nil
}

func (d *DB) Migrate(migrationsDir string) error {
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migration directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Apply each migration that hasn't been applied yet
	for _, f := range files {
		var count int
		err := d.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", f).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking migrations %s: %w", f, err)
		}
		if count > 0 {
			continue // already applied
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", f, err)
		}

		tx, err := d.Begin()
		if err != nil {
			return fmt.Errorf("starting transaction for %s: %w", f, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", f, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", f); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", f, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commiting migration %s: %w", f, err)
		}

		fmt.Printf("Applied migration: %s\n", f)
	}
	return nil
}