package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const dbFileName = "index.db"

// DatabasePath returns the path to the database file in the codeintelx directory.
func DatabasePath(codeintelxDir string) string {
	return filepath.Join(codeintelxDir, dbFileName)
}

// Initialize creates and initializes the SQLite database with the schema.
func Initialize(dbPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database (creates file if it doesn't exist)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Run migrations
	if err := migrate(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// migrate runs the database schema migrations.
func migrate(db *sql.DB) error {
	schema := `
		-- Meta table for key-value storage
		CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT
		);

		-- Projects table
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_root TEXT UNIQUE NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		-- Source roots table
		CREATE TABLE IF NOT EXISTS source_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}
