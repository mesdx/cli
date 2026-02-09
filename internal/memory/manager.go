package memory

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/codeintelx/cli/internal/search"
)

// IsIndexLockedError checks if the error is related to an index being locked
// by another process (e.g., MCP server).
func IsIndexLockedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Bleve typically returns errors containing "lock" when the index is locked
	return strings.Contains(errStr, "lock") ||
		strings.Contains(errStr, "locked") ||
		strings.Contains(errStr, "LOCK")
}

// Manager orchestrates memory CRUD, indexing, reconciliation, and search.
// The search index is opened once and kept for the lifetime of the Manager.
// Callers must call Close() when done to release the underlying Bleve index.
type Manager struct {
	Store     *MemoryStore
	RepoRoot  string
	MemoryDir string // absolute path to the memory directory
	searchIdx *search.MemoryIndex
	ProjectID int64
}

// NewManager creates a Manager for the given DB, project, and memory dir.
// The search index is opened in read-write mode and kept open for the
// lifetime of the Manager. Call Close() when done.
func NewManager(db *sql.DB, projectID int64, repoRoot, memoryDirAbs string) (*Manager, error) {
	searchBase := filepath.Join(repoRoot, ".codeintelx", "search")
	if err := search.EnsureIndexDir(searchBase); err != nil {
		return nil, fmt.Errorf("create search dir: %w", err)
	}

	idx, err := search.NewMemoryIndex(projectID, searchBase)
	if err != nil {
		return nil, fmt.Errorf("open search index: %w", err)
	}

	return &Manager{
		Store: &MemoryStore{
			DB:        db,
			ProjectID: projectID,
		},
		RepoRoot:  repoRoot,
		MemoryDir: memoryDirAbs,
		searchIdx: idx,
		ProjectID: projectID,
	}, nil
}

// NewManagerReadOnly creates a Manager whose search index is opened in
// read-only mode. This is safe to use while another process (e.g. the MCP
// server) has the index open for writing.
// Only Search / Read / List operations are expected; write operations that
// touch the index will return errors.
func NewManagerReadOnly(db *sql.DB, projectID int64, repoRoot, memoryDirAbs string) (*Manager, error) {
	searchBase := filepath.Join(repoRoot, ".codeintelx", "search")
	idx, err := search.NewMemoryIndexReadOnly(projectID, searchBase)
	if err != nil {
		return nil, fmt.Errorf("open search index (read-only): %w", err)
	}

	return &Manager{
		Store: &MemoryStore{
			DB:        db,
			ProjectID: projectID,
		},
		RepoRoot:  repoRoot,
		MemoryDir: memoryDirAbs,
		searchIdx: idx,
		ProjectID: projectID,
	}, nil
}

// Close releases the underlying search index. Must be called when the
// Manager is no longer needed.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	if m.searchIdx != nil {
		if err := m.searchIdx.Close(); err != nil {
			log.Printf("warning: failed to close search index: %v", err)
			return err
		}
	}
	return nil
}

// removeFromSearch removes a memory from the search index by mdRelPath.
func (m *Manager) removeFromSearch(mdRelPath string) error {
	return m.searchIdx.RemoveByMdRelPath(mdRelPath)
}

// MemoryElement is a fully-loaded memory element (frontmatter + body + paths).
type MemoryElement struct {
	Meta      *CodeintelxMeta `json:"meta"`
	Body      string          `json:"body"`
	MdRelPath string          `json:"mdRelPath"`
	AbsPath   string          `json:"absPath"`
}

// GrepReplaceResult captures the outcome of a grep/replace operation.
type GrepReplaceResult struct {
	MemoryUID    string `json:"memoryId"`
	MdRelPath    string `json:"mdRelPath"`
	Replacements int    `json:"replacements"`
}

// --- CRUD operations ---

// Append creates a new memory element and writes it to disk.
func (m *Manager) Append(scope, filePath, title, body string, symbols []SymbolRef) (*MemoryElement, error) {
	if scope == "" {
		scope = "project"
	}
	if scope == "file" && filePath == "" {
		return nil, fmt.Errorf("file path is required for file-scoped memories")
	}
	if scope == "file" {
		absRef := filepath.Join(m.RepoRoot, filePath)
		if _, err := os.Stat(absRef); err != nil {
			return nil, fmt.Errorf("referenced file does not exist: %s", filePath)
		}
	}

	meta := NewMeta(scope, filePath, title, symbols)

	// Generate filename
	filename := generateMdFilename(scope, filePath, title, meta.ID)
	mdRelPath := filename
	absPath := filepath.Join(m.MemoryDir, filename)

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, fmt.Errorf("create memory dir: %w", err)
	}

	// Write markdown file
	data, err := WriteMarkdown(meta, body)
	if err != nil {
		return nil, fmt.Errorf("write markdown: %w", err)
	}
	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// Index in DB
	hash := hashBytes(data)
	if _, err := m.Store.UpsertMemory(meta, mdRelPath, hash); err != nil {
		return nil, fmt.Errorf("upsert memory: %w", err)
	}
	if err := m.indexToSearch(meta, mdRelPath, body); err != nil {
		return nil, fmt.Errorf("index memory for search: %w", err)
	}

	return &MemoryElement{
		Meta:      meta,
		Body:      body,
		MdRelPath: mdRelPath,
		AbsPath:   absPath,
	}, nil
}

// Read reads a memory element by its UID.
func (m *Manager) Read(memoryUID string) (*MemoryElement, error) {
	row, err := m.Store.GetByUID(memoryUID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("memory %q not found", memoryUID)
	}
	return m.loadElement(row)
}

// ReadByPath reads a memory element by its markdown relative path.
func (m *Manager) ReadByPath(mdRelPath string) (*MemoryElement, error) {
	row, err := m.Store.GetByMdRelPath(mdRelPath)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("memory at %q not found", mdRelPath)
	}
	return m.loadElement(row)
}

// List returns memory rows matching the given filters.
func (m *Manager) List(scope, filePath string) ([]MemoryRow, error) {
	return m.Store.ListMemories(scope, filePath)
}

// Update updates the body, title, and/or symbols of an existing memory.
func (m *Manager) Update(memoryUID string, title *string, body *string, symbols *[]SymbolRef) (*MemoryElement, error) {
	row, err := m.Store.GetByUID(memoryUID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("memory %q not found", memoryUID)
	}

	absPath := filepath.Join(m.MemoryDir, row.MdRelPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	meta, existingBody, err := ParseMarkdown(data)
	if err != nil {
		return nil, fmt.Errorf("parse markdown: %w", err)
	}

	if title != nil {
		meta.Title = *title
	}
	if body != nil {
		existingBody = *body
	}
	if symbols != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		newSymbols := *symbols
		for i := range newSymbols {
			if newSymbols[i].Status == "" {
				newSymbols[i].Status = "active"
			}
			newSymbols[i].LastResolvedAt = now
		}
		meta.Symbols = newSymbols
	}

	newData, err := WriteMarkdown(meta, existingBody)
	if err != nil {
		return nil, fmt.Errorf("write markdown: %w", err)
	}
	if err := os.WriteFile(absPath, newData, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	hash := hashBytes(newData)
	if _, err := m.Store.UpsertMemory(meta, row.MdRelPath, hash); err != nil {
		return nil, fmt.Errorf("upsert memory: %w", err)
	}
	if err := m.indexToSearch(meta, row.MdRelPath, existingBody); err != nil {
		return nil, fmt.Errorf("index memory for search: %w", err)
	}

	return &MemoryElement{
		Meta:      meta,
		Body:      existingBody,
		MdRelPath: row.MdRelPath,
		AbsPath:   absPath,
	}, nil
}

// Delete soft-deletes a memory: marks status=deleted in frontmatter and DB,
// removes it from the search index, but keeps the file on disk.
func (m *Manager) Delete(memoryUID string) error {
	row, err := m.Store.GetByUID(memoryUID)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("memory %q not found", memoryUID)
	}

	// Update frontmatter on disk
	absPath := filepath.Join(m.MemoryDir, row.MdRelPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	meta, body, err := ParseMarkdown(data)
	if err != nil {
		// If frontmatter is broken, still soft-delete in DB
		return m.Store.SoftDeleteMemory(memoryUID)
	}

	meta.Status = "deleted"
	newData, err := WriteMarkdown(meta, body)
	if err != nil {
		return m.Store.SoftDeleteMemory(memoryUID)
	}
	if err := os.WriteFile(absPath, newData, 0644); err != nil {
		log.Printf("warning: failed to update frontmatter on disk: %v", err)
	}

	if err := m.Store.SoftDeleteMemory(memoryUID); err != nil {
		return err
	}
	// Best-effort remove from search by path.
	_ = m.removeFromSearch(row.MdRelPath)
	return nil
}

// GrepReplace performs a regex find-and-replace on the body of a single memory.
// The target must be identified by memoryUID or mdRelPath.
func (m *Manager) GrepReplace(memoryUID, mdRelPath, pattern, replacement string) (*GrepReplaceResult, error) {
	var row *MemoryRow
	var err error

	if memoryUID != "" {
		row, err = m.Store.GetByUID(memoryUID)
	} else if mdRelPath != "" {
		row, err = m.Store.GetByMdRelPath(mdRelPath)
	} else {
		return nil, fmt.Errorf("either memoryId or mdRelPath is required")
	}
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("memory not found")
	}

	absPath := filepath.Join(m.MemoryDir, row.MdRelPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	meta, body, err := ParseMarkdown(data)
	if err != nil {
		return nil, fmt.Errorf("parse markdown: %w", err)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindAllStringIndex(body, -1)
	if len(matches) == 0 {
		return &GrepReplaceResult{
			MemoryUID:    row.MemoryUID,
			MdRelPath:    row.MdRelPath,
			Replacements: 0,
		}, nil
	}

	newBody := re.ReplaceAllString(body, replacement)

	newData, err := WriteMarkdown(meta, newBody)
	if err != nil {
		return nil, fmt.Errorf("write markdown: %w", err)
	}
	if err := os.WriteFile(absPath, newData, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	hash := hashBytes(newData)
	if _, err := m.Store.UpsertMemory(meta, row.MdRelPath, hash); err != nil {
		return nil, fmt.Errorf("upsert memory: %w", err)
	}
	if err := m.indexToSearch(meta, row.MdRelPath, newBody); err != nil {
		return nil, fmt.Errorf("index memory for search: %w", err)
	}

	return &GrepReplaceResult{
		MemoryUID:    row.MemoryUID,
		MdRelPath:    row.MdRelPath,
		Replacements: len(matches),
	}, nil
}

// Search performs a full-text search over memory elements.
func (m *Manager) Search(query, scope, filePath string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	hits, err := m.searchIdx.Search(query, scope, filePath, limit)
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(hits))
	for _, h := range hits {
		results = append(results, SearchResult{
			MemoryRow: MemoryRow{
				MemoryUID:  h.MemoryUID,
				Scope:      h.Scope,
				FilePath:   h.FilePath,
				MdRelPath:  h.MdRelPath,
				Title:      h.Title,
				Status:     "active",
				FileStatus: "active",
			},
			Score:   h.Score,
			Snippet: h.Snippet,
		})
	}
	return results, nil
}

// --- Indexing ---

// BulkIndex scans the memory directory and indexes all markdown files.
// Files with unparsable frontmatter are merged into the canonical project.md.
func (m *Manager) BulkIndex() error {
	if err := os.MkdirAll(m.MemoryDir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	if err := m.searchIdx.Reset(); err != nil {
		return fmt.Errorf("reset memory search index: %w", err)
	}

	return filepath.Walk(m.MemoryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}
		if indexErr := m.IndexFile(path); indexErr != nil {
			log.Printf("memory: failed to index %s: %v", path, indexErr)
		}
		return nil
	})
}

// IndexFile indexes a single markdown memory file.
// If frontmatter is unparsable, the content is merged into project.md.
func (m *Manager) IndexFile(absPath string) error {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	mdRelPath, err := filepath.Rel(m.MemoryDir, absPath)
	if err != nil {
		return err
	}

	meta, body, parseErr := ParseMarkdown(data)
	if parseErr != nil {
		// Frontmatter not parsable — merge into project.md
		return m.salvageMerge(mdRelPath, string(data))
	}

	hash := hashBytes(data)
	// Determine "validity" for indexing.
	if meta.Scope == "file" && meta.File != "" {
		repoFile := filepath.Join(m.RepoRoot, meta.File)
		if _, err := os.Stat(repoFile); os.IsNotExist(err) {
			meta.FileStatus = "deleted"
		}
	}

	if _, err := m.Store.UpsertMemory(meta, mdRelPath, hash); err != nil {
		return err
	}
	return m.indexToSearch(meta, mdRelPath, body)
}

// RemoveFile handles deletion of a memory markdown file from disk.
// Marks the memory as having a deleted file status.
func (m *Manager) RemoveFile(absPath string) error {
	mdRelPath, err := filepath.Rel(m.MemoryDir, absPath)
	if err != nil {
		return err
	}
	_ = m.removeFromSearch(mdRelPath)
	return m.Store.DeleteByMdRelPath(mdRelPath)
}

// --- Reconcile ---

// Reconcile checks all indexed memories and updates file/symbol statuses.
// It also removes stale entries for markdown files that no longer exist.
func (m *Manager) Reconcile() error {
	allPaths, err := m.Store.AllMemoryPaths()
	if err != nil {
		return fmt.Errorf("load memory paths: %w", err)
	}

	for mdRelPath := range allPaths {
		absPath := filepath.Join(m.MemoryDir, mdRelPath)

		// Check if the markdown file itself still exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			if err := m.Store.DeleteByMdRelPath(mdRelPath); err != nil {
				log.Printf("memory: failed to delete stale entry %s: %v", mdRelPath, err)
			}
			_ = m.removeFromSearch(mdRelPath)
			continue
		}

		// Re-read and reconcile the memory
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		meta, body, parseErr := ParseMarkdown(data)
		if parseErr != nil {
			continue
		}

		row, err := m.Store.GetByMdRelPath(mdRelPath)
		if err != nil || row == nil {
			continue
		}

		changed := false

		// Check if the referenced file still exists (file-scoped memories)
		if meta.Scope == "file" && meta.File != "" {
			repoFile := filepath.Join(m.RepoRoot, meta.File)
			if _, err := os.Stat(repoFile); os.IsNotExist(err) {
				if meta.FileStatus != "deleted" {
					meta.FileStatus = "deleted"
					changed = true
				}
				// Update DB — also remove from search so it won't appear in results
				if err := m.Store.UpdateFileStatus(row.ID, "deleted"); err != nil {
					log.Printf("memory: failed to update file_status for %s: %v", mdRelPath, err)
				}
				_ = m.removeFromSearch(mdRelPath)
			} else if meta.FileStatus == "deleted" {
				// File was restored
				meta.FileStatus = "active"
				changed = true
				if err := m.Store.UpdateFileStatus(row.ID, "active"); err != nil {
					log.Printf("memory: failed to restore file_status for %s: %v", mdRelPath, err)
				}
				_ = m.indexToSearch(meta, mdRelPath, body)
			}
		}

		// Check symbol references
		for i, sym := range meta.Symbols {
			exists, err := m.Store.SymbolExistsInIndex(sym.Language, sym.Name)
			if err != nil {
				continue
			}
			now := time.Now().UTC().Format(time.RFC3339)
			if !exists && sym.Status != "deleted" {
				meta.Symbols[i].Status = "deleted"
				meta.Symbols[i].LastResolvedAt = now
				changed = true
				_ = m.Store.UpdateSymbolStatus(row.ID, sym.Language, sym.Name, "deleted")
			} else if exists && sym.Status == "deleted" {
				// Symbol restored
				meta.Symbols[i].Status = "active"
				meta.Symbols[i].LastResolvedAt = now
				changed = true
				_ = m.Store.UpdateSymbolStatus(row.ID, sym.Language, sym.Name, "active")
			}
		}

		// Write back updated frontmatter if anything changed
		if changed {
			newData, err := WriteMarkdown(meta, body)
			if err != nil {
				log.Printf("memory: failed to write updated frontmatter for %s: %v", mdRelPath, err)
				continue
			}
			if err := os.WriteFile(absPath, newData, 0644); err != nil {
				log.Printf("memory: failed to write file %s: %v", absPath, err)
			}
		}
	}

	return nil
}

// ReconcileFileRef reconciles memories that reference a specific repo-relative file path.
// This is used by fsnotify-driven reconciliation when source files are created/removed.
func (m *Manager) ReconcileFileRef(fileRelPath string) error {
	rows, err := m.Store.ListMemories("file", fileRelPath)
	if err != nil {
		return err
	}
	for _, r := range rows {
		absPath := filepath.Join(m.MemoryDir, r.MdRelPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			// If the memory file itself vanished, let memory watcher clean it up.
			continue
		}
		meta, body, err := ParseMarkdown(data)
		if err != nil {
			continue
		}

		repoFile := filepath.Join(m.RepoRoot, meta.File)
		_, statErr := os.Stat(repoFile)
		fileExists := statErr == nil

		changed := false
		if !fileExists {
			if meta.FileStatus != "deleted" {
				meta.FileStatus = "deleted"
				changed = true
			}
			_ = m.Store.UpdateFileStatus(r.ID, "deleted")
			_ = m.removeFromSearch(r.MdRelPath)
		} else {
			if meta.FileStatus == "deleted" {
				meta.FileStatus = "active"
				changed = true
				_ = m.Store.UpdateFileStatus(r.ID, "active")
			}
			_ = m.indexToSearch(meta, r.MdRelPath, body)
		}

		if changed {
			newData, err := WriteMarkdown(meta, body)
			if err == nil {
				_ = os.WriteFile(absPath, newData, 0644)
			}
		}
	}
	return nil
}

// --- Salvage merge ---

const projectMdFilename = "project.md"

// salvageMerge merges unparseable markdown content into the canonical project.md.
func (m *Manager) salvageMerge(sourceMdRelPath, content string) error {
	projectMdPath := filepath.Join(m.MemoryDir, projectMdFilename)

	// Read existing project.md or create fresh
	var existingData []byte
	if data, err := os.ReadFile(projectMdPath); err == nil {
		existingData = data
	}

	var meta *CodeintelxMeta
	var existingBody string

	if len(existingData) > 0 {
		parsed, body, err := ParseMarkdown(existingData)
		if err != nil {
			// project.md itself is broken, recreate it
			meta = NewMeta("project", "", "Project Memory", nil)
			existingBody = string(existingData)
		} else {
			meta = parsed
			existingBody = body
		}
	} else {
		meta = NewMeta("project", "", "Project Memory", nil)
	}

	// Append merged section
	section := fmt.Sprintf(
		"\n\n## Imported (unparseable frontmatter): %s\n\n_Imported at: %s_\n\n%s\n",
		sourceMdRelPath,
		time.Now().UTC().Format(time.RFC3339),
		strings.TrimSpace(content),
	)
	existingBody += section

	newData, err := WriteMarkdown(meta, existingBody)
	if err != nil {
		return fmt.Errorf("write project.md: %w", err)
	}
	if err := os.WriteFile(projectMdPath, newData, 0644); err != nil {
		return fmt.Errorf("write project.md file: %w", err)
	}

	// Re-index project.md
	hash := hashBytes(newData)
	if _, err := m.Store.UpsertMemory(meta, projectMdFilename, hash); err != nil {
		return err
	}
	return m.indexToSearch(meta, projectMdFilename, existingBody)
}

// --- Helpers ---

func hashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func generateMdFilename(scope, filePath, title, id string) string {
	var slug string
	if title != "" {
		slug = slugify(title)
	} else {
		slug = id
		if len(slug) > 8 {
			slug = slug[:8]
		}
	}

	switch scope {
	case "file":
		sanitized := sanitizePath(filePath)
		return fmt.Sprintf("file-%s-%s.md", sanitized, slug)
	default:
		return fmt.Sprintf("project-%s.md", slug)
	}
}

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	result := b.String()
	// Collapse multiple dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func sanitizePath(path string) string {
	s := strings.ReplaceAll(path, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

// loadElement reads a memory from disk given its DB row.
func (m *Manager) loadElement(row *MemoryRow) (*MemoryElement, error) {
	absPath := filepath.Join(m.MemoryDir, row.MdRelPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	meta, body, err := ParseMarkdown(data)
	if err != nil {
		// Return what we can from DB
		return &MemoryElement{
			Meta: &CodeintelxMeta{
				ID:         row.MemoryUID,
				Scope:      row.Scope,
				File:       row.FilePath,
				Title:      row.Title,
				Status:     row.Status,
				FileStatus: row.FileStatus,
			},
			Body:      string(data),
			MdRelPath: row.MdRelPath,
			AbsPath:   absPath,
		}, nil
	}

	return &MemoryElement{
		Meta:      meta,
		Body:      body,
		MdRelPath: row.MdRelPath,
		AbsPath:   absPath,
	}, nil
}

func (m *Manager) indexToSearch(meta *CodeintelxMeta, mdRelPath, body string) error {
	if meta == nil {
		return nil
	}

	// Exclude deleted from index.
	if meta.Status == "deleted" || meta.FileStatus == "deleted" {
		return m.searchIdx.RemoveByMdRelPath(mdRelPath)
	}

	// Validity: file-scoped memories must reference an existing file.
	if meta.Scope == "file" {
		if meta.File == "" {
			return m.searchIdx.RemoveByMdRelPath(mdRelPath)
		}
		repoFile := filepath.Join(m.RepoRoot, meta.File)
		if _, err := os.Stat(repoFile); err != nil {
			return m.searchIdx.RemoveByMdRelPath(mdRelPath)
		}
	}

	symbols := make([]string, 0, len(meta.Symbols))
	for _, s := range meta.Symbols {
		if s.Name != "" {
			symbols = append(symbols, s.Name)
		}
	}

	return m.searchIdx.IndexMemory(
		meta.ID,
		meta.Scope,
		meta.File,
		mdRelPath,
		meta.Title,
		meta.Status,
		meta.FileStatus,
		body,
		symbols,
	)
}
