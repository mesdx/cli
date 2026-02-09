package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFileName = "index.db"

// DatabasePath returns the path to the database file in the mesdx directory.
func DatabasePath(mesdxDir string) string {
	return filepath.Join(mesdxDir, dbFileName)
}

// Open opens the database file and returns a *sql.DB handle.
// Callers are responsible for closing.
func Open(dbPath string) (*sql.DB, error) {
	d, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// Enable WAL mode + foreign keys for performance and correctness.
	if _, err := d.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}
	return d, nil
}

// Initialize creates and initializes the SQLite database with the schema.
func Initialize(dbPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	d, err := Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()

	// Run versioned migrations
	if err := Migrate(d); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
