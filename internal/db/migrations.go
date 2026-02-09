package db

import (
	"database/sql"
	"fmt"
	"time"
)

// migration represents a single schema migration.
type migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations is the ordered list of all schema migrations.
// Pre-release: memory tables are included in v2 instead of a separate migration.
var migrations = []migration{
	{
		Version: 1,
		Name:    "initial_schema",
		SQL: `
			CREATE TABLE IF NOT EXISTS meta (
				key TEXT PRIMARY KEY,
				value TEXT
			);

			CREATE TABLE IF NOT EXISTS projects (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				repo_root TEXT UNIQUE NOT NULL,
				created_at TEXT NOT NULL DEFAULT (datetime('now'))
			);

			CREATE TABLE IF NOT EXISTS source_roots (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id INTEGER NOT NULL,
				path TEXT NOT NULL,
				FOREIGN KEY (project_id) REFERENCES projects(id)
			);
		`,
	},
	{
		Version: 2,
		Name:    "add_files_symbols_references_and_memories",
		SQL: `
			-- Indexed source files
			CREATE TABLE IF NOT EXISTS files (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id INTEGER NOT NULL,
				path TEXT NOT NULL,
				lang TEXT NOT NULL DEFAULT '',
				sha256 TEXT NOT NULL DEFAULT '',
				size_bytes INTEGER NOT NULL DEFAULT 0,
				mtime_unix INTEGER NOT NULL DEFAULT 0,
				indexed_at TEXT NOT NULL DEFAULT (datetime('now')),
				FOREIGN KEY (project_id) REFERENCES projects(id),
				UNIQUE (project_id, path)
			);
			CREATE INDEX IF NOT EXISTS idx_files_project ON files(project_id);

			-- Symbol definitions
			CREATE TABLE IF NOT EXISTS symbols (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				file_id INTEGER NOT NULL,
				name TEXT NOT NULL,
				kind INTEGER NOT NULL,
				container_name TEXT NOT NULL DEFAULT '',
				signature TEXT NOT NULL DEFAULT '',
				start_line INTEGER NOT NULL,
				start_col INTEGER NOT NULL,
				end_line INTEGER NOT NULL,
				end_col INTEGER NOT NULL,
				FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_symbols_name_kind ON symbols(name, kind);
			CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(file_id);

			-- Usage references
			CREATE TABLE IF NOT EXISTS refs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				file_id INTEGER NOT NULL,
				name TEXT NOT NULL,
				kind INTEGER NOT NULL DEFAULT 0,
				start_line INTEGER NOT NULL,
				start_col INTEGER NOT NULL,
				end_line INTEGER NOT NULL,
				end_col INTEGER NOT NULL,
				context_container TEXT NOT NULL DEFAULT '',
				FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_refs_name ON refs(name);
			CREATE INDEX IF NOT EXISTS idx_refs_file ON refs(file_id);

			-- Memory elements (markdown-backed, scoped to project or file)
			CREATE TABLE IF NOT EXISTS memories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				project_id INTEGER NOT NULL,
				memory_uid TEXT NOT NULL,
				scope TEXT NOT NULL DEFAULT 'project',
				file_path TEXT NOT NULL DEFAULT '',
				md_rel_path TEXT NOT NULL,
				title TEXT NOT NULL DEFAULT '',
				status TEXT NOT NULL DEFAULT 'active',
				file_status TEXT NOT NULL DEFAULT 'active',
				body_hash TEXT NOT NULL DEFAULT '',
				created_at TEXT NOT NULL DEFAULT (datetime('now')),
				updated_at TEXT NOT NULL DEFAULT (datetime('now')),
				FOREIGN KEY (project_id) REFERENCES projects(id),
				UNIQUE (project_id, md_rel_path)
			);
			CREATE INDEX IF NOT EXISTS idx_memories_project_scope ON memories(project_id, scope, file_path);
			CREATE INDEX IF NOT EXISTS idx_memories_md_rel_path ON memories(project_id, md_rel_path);
			CREATE INDEX IF NOT EXISTS idx_memories_uid ON memories(project_id, memory_uid);

			-- Symbol references attached to memories
			CREATE TABLE IF NOT EXISTS memory_symbols (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				memory_id INTEGER NOT NULL,
				language TEXT NOT NULL,
				name TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'active',
				last_resolved_at TEXT NOT NULL DEFAULT '',
				FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_memory_symbols_memory ON memory_symbols(memory_id);
		`,
	},
}

// Migrate runs all pending versioned migrations inside transactions.
// It creates the schema_migrations table if it does not exist, detects
// databases that were created before versioned migrations were introduced,
// and backfills version 1 for them.
func Migrate(d *sql.DB) error {
	// Ensure the schema_migrations table exists.
	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name    TEXT NOT NULL,
			applied_at TEXT NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Backfill: if the old "meta" table exists but no migration record,
	// record version 1 so we don't re-run it.
	if err := backfillV1(d); err != nil {
		return fmt.Errorf("backfill v1: %w", err)
	}

	// Determine current version.
	var current int
	row := d.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}

	// Apply pending migrations.
	for _, m := range migrations {
		if m.Version <= current {
			continue
		}
		if err := applyMigration(d, m); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.Version, m.Name, err)
		}
	}
	return nil
}

func applyMigration(d *sql.DB, m migration) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(m.SQL); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		m.Version, m.Name, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("record: %w", err)
	}
	return tx.Commit()
}

// backfillV1 checks if the database already has the v1 tables (meta, projects,
// source_roots) but no migration record. If so it records version 1 without
// re-running the DDL.
func backfillV1(d *sql.DB) error {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already recorded
	}

	// Check if "meta" table exists (proxy for the v1 schema already being present).
	var name string
	err = d.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='meta'`).Scan(&name)
	if err == sql.ErrNoRows {
		return nil // fresh DB, nothing to backfill
	}
	if err != nil {
		return err
	}

	// v1 tables exist but migration record is missing — backfill.
	_, err = d.Exec(
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		1, "initial_schema", time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// CurrentVersion returns the highest applied migration version (0 if none).
func CurrentVersion(d *sql.DB) (int, error) {
	var v int
	err := d.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)
	return v, err
}

// LatestVersion returns the latest migration version defined in code.
func LatestVersion() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].Version
}
