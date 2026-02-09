package memory

import (
	"database/sql"
	"fmt"
	"time"
)

// MemoryStore wraps DB operations for memory elements.
type MemoryStore struct {
	DB        *sql.DB
	ProjectID int64
}

// MemoryRow represents a row in the memories table.
type MemoryRow struct {
	ID         int64  `json:"rowId"`
	MemoryUID  string `json:"memoryId"`
	Scope      string `json:"scope"`
	FilePath   string `json:"filePath,omitempty"`
	MdRelPath  string `json:"mdRelPath"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status"`
	FileStatus string `json:"fileStatus"`
	BodyHash   string `json:"-"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// SearchResult represents a search hit with score.
type SearchResult struct {
	MemoryRow
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet,omitempty"`
}

// UpsertMemory inserts or updates a memory record and its symbols.
func (s *MemoryStore) UpsertMemory(meta *CodeintelxMeta, mdRelPath, bodyHash string) (int64, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().UTC().Format(time.RFC3339)

	var rowID int64
	err = tx.QueryRow(
		`SELECT id FROM memories WHERE project_id = ? AND md_rel_path = ?`,
		s.ProjectID, mdRelPath,
	).Scan(&rowID)

	if err == sql.ErrNoRows {
		res, err := tx.Exec(
			`INSERT INTO memories (project_id, memory_uid, scope, file_path, md_rel_path, title, status, file_status, body_hash, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.ProjectID, meta.ID, meta.Scope, meta.File, mdRelPath,
			meta.Title, meta.Status, meta.FileStatus, bodyHash, now, now,
		)
		if err != nil {
			return 0, fmt.Errorf("insert memory: %w", err)
		}
		rowID, _ = res.LastInsertId()
	} else if err != nil {
		return 0, err
	} else {
		_, err = tx.Exec(
			`UPDATE memories SET memory_uid=?, scope=?, file_path=?, title=?, status=?, file_status=?, body_hash=?, updated_at=? WHERE id=?`,
			meta.ID, meta.Scope, meta.File, meta.Title, meta.Status, meta.FileStatus, bodyHash, now, rowID,
		)
		if err != nil {
			return 0, fmt.Errorf("update memory: %w", err)
		}
		// Delete old symbols
		if _, err := tx.Exec(`DELETE FROM memory_symbols WHERE memory_id = ?`, rowID); err != nil {
			return 0, err
		}
	}

	// Insert symbols
	for _, sym := range meta.Symbols {
		if _, err := tx.Exec(
			`INSERT INTO memory_symbols (memory_id, language, name, status, last_resolved_at) VALUES (?, ?, ?, ?, ?)`,
			rowID, sym.Language, sym.Name, sym.Status, sym.LastResolvedAt,
		); err != nil {
			return 0, fmt.Errorf("insert memory symbol: %w", err)
		}
	}

	return rowID, tx.Commit()
}

// SoftDeleteMemory marks a memory as deleted. The file content is preserved.
func (s *MemoryStore) SoftDeleteMemory(memoryUID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.DB.Exec(
		`UPDATE memories SET status='deleted', updated_at=? WHERE project_id=? AND memory_uid=?`,
		now, s.ProjectID, memoryUID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("memory %q not found", memoryUID)
	}
	return nil
}

// GetByUID returns a memory row by its UID.
func (s *MemoryStore) GetByUID(uid string) (*MemoryRow, error) {
	var r MemoryRow
	err := s.DB.QueryRow(
		`SELECT id, memory_uid, scope, file_path, md_rel_path, title, status, file_status, body_hash, created_at, updated_at
		 FROM memories WHERE project_id=? AND memory_uid=?`,
		s.ProjectID, uid,
	).Scan(&r.ID, &r.MemoryUID, &r.Scope, &r.FilePath, &r.MdRelPath, &r.Title, &r.Status, &r.FileStatus, &r.BodyHash, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetByMdRelPath returns a memory row by its markdown relative path.
func (s *MemoryStore) GetByMdRelPath(mdRelPath string) (*MemoryRow, error) {
	var r MemoryRow
	err := s.DB.QueryRow(
		`SELECT id, memory_uid, scope, file_path, md_rel_path, title, status, file_status, body_hash, created_at, updated_at
		 FROM memories WHERE project_id=? AND md_rel_path=?`,
		s.ProjectID, mdRelPath,
	).Scan(&r.ID, &r.MemoryUID, &r.Scope, &r.FilePath, &r.MdRelPath, &r.Title, &r.Status, &r.FileStatus, &r.BodyHash, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListMemories returns memories matching the given scope and optional file filter.
func (s *MemoryStore) ListMemories(scope, filePath string) ([]MemoryRow, error) {
	query := `SELECT id, memory_uid, scope, file_path, md_rel_path, title, status, file_status, body_hash, created_at, updated_at
		FROM memories WHERE project_id = ?`
	args := []interface{}{s.ProjectID}

	if scope != "" {
		query += ` AND scope = ?`
		args = append(args, scope)
	}
	if filePath != "" {
		query += ` AND file_path = ?`
		args = append(args, filePath)
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []MemoryRow
	for rows.Next() {
		var r MemoryRow
		if err := rows.Scan(&r.ID, &r.MemoryUID, &r.Scope, &r.FilePath, &r.MdRelPath, &r.Title, &r.Status, &r.FileStatus, &r.BodyHash, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// AllMemoryPaths returns all md_rel_path values for the project.
func (s *MemoryStore) AllMemoryPaths() (map[string]string, error) {
	rows, err := s.DB.Query(
		`SELECT md_rel_path, body_hash FROM memories WHERE project_id = ?`, s.ProjectID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[string]string)
	for rows.Next() {
		var path, hash string
		if err := rows.Scan(&path, &hash); err != nil {
			return nil, err
		}
		m[path] = hash
	}
	return m, rows.Err()
}

// DeleteByMdRelPath hard-deletes a memory by its markdown path.
func (s *MemoryStore) DeleteByMdRelPath(mdRelPath string) error {
	var rowID int64
	err := s.DB.QueryRow(
		`SELECT id FROM memories WHERE project_id = ? AND md_rel_path = ?`,
		s.ProjectID, mdRelPath,
	).Scan(&rowID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := s.DB.Exec(`DELETE FROM memory_symbols WHERE memory_id = ?`, rowID); err != nil {
		return err
	}
	_, err = s.DB.Exec(`DELETE FROM memories WHERE id = ?`, rowID)
	return err
}

// UpdateFileStatus updates the file_status for a memory.
func (s *MemoryStore) UpdateFileStatus(rowID int64, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec(
		`UPDATE memories SET file_status=?, updated_at=? WHERE id=?`,
		status, now, rowID,
	)
	return err
}

// SymbolExistsInIndex checks whether a symbol name exists in the code index
// for the given language and project.
func (s *MemoryStore) SymbolExistsInIndex(language, name string) (bool, error) {
	// Map language strings to file lang identifiers used in the files table
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*) FROM symbols sy
		JOIN files f ON sy.file_id = f.id
		WHERE f.project_id = ? AND sy.name = ? AND f.lang = ?`,
		s.ProjectID, name, language,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateSymbolStatus updates a specific memory_symbol status.
func (s *MemoryStore) UpdateSymbolStatus(memoryRowID int64, language, name, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec(
		`UPDATE memory_symbols SET status=?, last_resolved_at=? WHERE memory_id=? AND language=? AND name=?`,
		status, now, memoryRowID, language, name,
	)
	return err
}
