package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateCreatesAllTables(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := Initialize(dbPath); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = d.Close() }()

	// All expected tables must exist.
	for _, table := range []string{"schema_migrations", "meta", "projects", "source_roots", "files", "symbols", "refs", "memories", "memory_symbols", "memory_ngrams"} {
		var name string
		err := d.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := Initialize(dbPath); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = d.Close() }()

	// Running migrate again should not error.
	if err := Migrate(d); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	v, err := CurrentVersion(d)
	if err != nil {
		t.Fatalf("CurrentVersion: %v", err)
	}
	if v != LatestVersion() {
		t.Errorf("version = %d, want %d", v, LatestVersion())
	}
}

func TestMigrateRecordsMigrationVersions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := Initialize(dbPath); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = d.Close() }()

	rows, err := d.Query(`SELECT version, name FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var versions []int
	for rows.Next() {
		var v int
		var name string
		if err := rows.Scan(&v, &name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		versions = append(versions, v)
		if name == "" {
			t.Errorf("migration %d has empty name", v)
		}
	}

	if len(versions) != len(migrations) {
		t.Errorf("got %d migration records, want %d", len(versions), len(migrations))
	}
}

func TestBackfillV1(t *testing.T) {
	// Simulate an old database that has the v1 tables but no schema_migrations record.
	dbPath := filepath.Join(t.TempDir(), "old.db")

	d, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Create v1 tables manually (simulating old code path).
	_, err = d.Exec(`
		CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT);
		CREATE TABLE projects (id INTEGER PRIMARY KEY AUTOINCREMENT, repo_root TEXT UNIQUE NOT NULL, created_at TEXT NOT NULL DEFAULT (datetime('now')));
		CREATE TABLE source_roots (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL, path TEXT NOT NULL);
	`)
	if err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	_ = d.Close()

	// Now run Migrate — it should detect v1 tables, backfill, then apply v2.
	d2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = d2.Close() }()

	if err := Migrate(d2); err != nil {
		t.Fatalf("Migrate on old DB: %v", err)
	}

	v, err := CurrentVersion(d2)
	if err != nil {
		t.Fatalf("CurrentVersion: %v", err)
	}
	if v != LatestVersion() {
		t.Errorf("version after backfill = %d, want %d", v, LatestVersion())
	}

	// Verify v1 is recorded.
	var count int
	if err := d2.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&count); err != nil {
		t.Fatalf("failed to query v1 migration record: %v", err)
	}
	if count != 1 {
		t.Errorf("v1 migration record count = %d, want 1", count)
	}
}

func TestInitializeCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	dbPath := filepath.Join(dir, "test.db")

	if err := Initialize(dbPath); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("db file not created: %v", err)
	}
}
