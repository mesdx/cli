package indexer

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mesdx/cli/internal/symbols"
)

// Store wraps DB operations for the indexer.
type Store struct {
	DB        *sql.DB
	ProjectID int64
}

// EnsureProject ensures a project row exists and returns its ID.
func (s *Store) EnsureProject(repoRoot string) error {
	var id int64
	err := s.DB.QueryRow(`SELECT id FROM projects WHERE repo_root = ?`, repoRoot).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := s.DB.Exec(
			`INSERT INTO projects (repo_root) VALUES (?)`, repoRoot,
		)
		if err != nil {
			return fmt.Errorf("insert project: %w", err)
		}
		id, _ = res.LastInsertId()
	} else if err != nil {
		return fmt.Errorf("query project: %w", err)
	}
	s.ProjectID = id
	return nil
}

// EnsureSourceRoots replaces the stored source roots for the project.
func (s *Store) EnsureSourceRoots(roots []string) error {
	if _, err := s.DB.Exec(`DELETE FROM source_roots WHERE project_id = ?`, s.ProjectID); err != nil {
		return err
	}
	for _, r := range roots {
		if _, err := s.DB.Exec(
			`INSERT INTO source_roots (project_id, path) VALUES (?, ?)`,
			s.ProjectID, r,
		); err != nil {
			return err
		}
	}
	return nil
}

// FileRow represents a row in the files table.
type FileRow struct {
	ID     int64
	Path   string
	SHA256 string
}

// GetFile looks up an indexed file by path.
func (s *Store) GetFile(path string) (*FileRow, error) {
	var f FileRow
	err := s.DB.QueryRow(
		`SELECT id, path, sha256 FROM files WHERE project_id = ? AND path = ?`,
		s.ProjectID, path,
	).Scan(&f.ID, &f.Path, &f.SHA256)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// AllFiles returns all indexed file paths and their SHA for the project.
func (s *Store) AllFiles() (map[string]string, error) {
	rows, err := s.DB.Query(
		`SELECT path, sha256 FROM files WHERE project_id = ?`, s.ProjectID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := map[string]string{}
	for rows.Next() {
		var path, sha string
		if err := rows.Scan(&path, &sha); err != nil {
			return nil, err
		}
		m[path] = sha
	}
	return m, rows.Err()
}

// UpsertFile inserts or updates a file record and replaces its symbols/refs.
func (s *Store) UpsertFile(path string, lang Lang, sha string, sizeBytes int64, mtimeUnix int64, fr *symbols.FileResult) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// Upsert file row
	var fileID int64
	err = tx.QueryRow(
		`SELECT id FROM files WHERE project_id = ? AND path = ?`,
		s.ProjectID, path,
	).Scan(&fileID)

	now := time.Now().UTC().Format(time.RFC3339)
	if err == sql.ErrNoRows {
		res, err := tx.Exec(
			`INSERT INTO files (project_id, path, lang, sha256, size_bytes, mtime_unix, indexed_at) VALUES (?,?,?,?,?,?,?)`,
			s.ProjectID, path, string(lang), sha, sizeBytes, mtimeUnix, now,
		)
		if err != nil {
			return fmt.Errorf("insert file: %w", err)
		}
		fileID, _ = res.LastInsertId()
	} else if err != nil {
		return err
	} else {
		// Update existing
		if _, err := tx.Exec(
			`UPDATE files SET lang=?, sha256=?, size_bytes=?, mtime_unix=?, indexed_at=? WHERE id=?`,
			string(lang), sha, sizeBytes, mtimeUnix, now, fileID,
		); err != nil {
			return fmt.Errorf("update file: %w", err)
		}
		// Delete old symbols and refs for this file
		if _, err := tx.Exec(`DELETE FROM symbols WHERE file_id = ?`, fileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM refs WHERE file_id = ?`, fileID); err != nil {
			return err
		}
	}

	// Insert symbols
	for _, sym := range fr.Symbols {
		if _, err := tx.Exec(
			`INSERT INTO symbols (file_id, name, kind, container_name, signature, start_line, start_col, end_line, end_col)
			 VALUES (?,?,?,?,?,?,?,?,?)`,
			fileID, sym.Name, int(sym.Kind), sym.ContainerName, sym.Signature,
			sym.StartLine, sym.StartCol, sym.EndLine, sym.EndCol,
		); err != nil {
			return fmt.Errorf("insert symbol %q: %w", sym.Name, err)
		}
	}

	// Insert refs
	for _, ref := range fr.Refs {
		if _, err := tx.Exec(
			`INSERT INTO refs (file_id, name, kind, start_line, start_col, end_line, end_col, context_container)
			 VALUES (?,?,?,?,?,?,?,?)`,
			fileID, ref.Name, int(ref.Kind),
			ref.StartLine, ref.StartCol, ref.EndLine, ref.EndCol,
			ref.ContextContainer,
		); err != nil {
			return fmt.Errorf("insert ref %q: %w", ref.Name, err)
		}
	}

	return tx.Commit()
}

// DeleteFile removes a file and its associated symbols/refs.
func (s *Store) DeleteFile(path string) error {
	var fileID int64
	err := s.DB.QueryRow(
		`SELECT id FROM files WHERE project_id = ? AND path = ?`,
		s.ProjectID, path,
	).Scan(&fileID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	// CASCADE should handle symbols/refs, but be explicit.
	if _, err := s.DB.Exec(`DELETE FROM symbols WHERE file_id = ?`, fileID); err != nil {
		return err
	}
	if _, err := s.DB.Exec(`DELETE FROM refs WHERE file_id = ?`, fileID); err != nil {
		return err
	}
	_, err = s.DB.Exec(`DELETE FROM files WHERE id = ?`, fileID)
	return err
}

// DeleteAllFiles removes all files (and cascaded symbols/refs) for the project.
func (s *Store) DeleteAllFiles() error {
	rows, err := s.DB.Query(`SELECT id FROM files WHERE project_id = ?`, s.ProjectID)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := s.DB.Exec(`DELETE FROM symbols WHERE file_id = ?`, id); err != nil {
			return err
		}
		if _, err := s.DB.Exec(`DELETE FROM refs WHERE file_id = ?`, id); err != nil {
			return err
		}
		if _, err := s.DB.Exec(`DELETE FROM files WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}
