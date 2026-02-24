package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLite returns DATETIME columns as strings; these scanners handle both string and time.Time.

var sqliteTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05 -0700 -0700", // Go time.Time.String() format
	"2006-01-02 15:04:05 -0700 MST",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseSQLiteTime(v interface{}) (time.Time, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	case string:
		for _, layout := range sqliteTimeLayouts {
			if t, err := time.Parse(layout, x); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse %q as time", x)
	}
	return time.Time{}, fmt.Errorf("unsupported time scan type %T", v)
}

// NullableTime scans a nullable SQLite datetime column into *time.Time.
type NullableTime struct{ T **time.Time }

func (s NullableTime) Scan(v interface{}) error {
	if v == nil {
		*s.T = nil
		return nil
	}
	t, err := parseSQLiteTime(v)
	if err != nil {
		return err
	}
	*s.T = &t
	return nil
}

// RequiredTime scans a non-null SQLite datetime column into time.Time.
type RequiredTime struct{ T *time.Time }

func (s RequiredTime) Scan(v interface{}) error {
	if v == nil {
		*s.T = time.Time{}
		return nil
	}
	t, err := parseSQLiteTime(v)
	if err != nil {
		return err
	}
	*s.T = t
	return nil
}

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