package indexer

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Indexer orchestrates file discovery, hashing, parsing, and DB storage.
type Indexer struct {
	Store    *Store
	RepoRoot string
}

// New creates an Indexer for the given DB and repo root.
func New(d *sql.DB, repoRoot string) *Indexer {
	return &Indexer{
		Store:    &Store{DB: d},
		RepoRoot: repoRoot,
	}
}

// excluded directories for walking.
var excludedDirs = map[string]bool{
	".git":         true,
	".mesdx":  true,
	"node_modules": true,
	".venv":        true,
	"venv":         true,
	".env":         true,
	"vendor":       true,
	"target":       true,
	"build":        true,
	"dist":         true,
	".idea":        true,
	".vscode":      true,
	"__pycache__":  true,
	".mypy_cache":  true,
}

// FullIndex performs a bulk index over all source roots.
// It wipes existing file/symbol/ref data for the project and re-indexes everything.
func (idx *Indexer) FullIndex(sourceRoots []string) (*IndexStats, error) {
	if err := idx.Store.EnsureProject(idx.RepoRoot); err != nil {
		return nil, fmt.Errorf("ensure project: %w", err)
	}
	if err := idx.Store.EnsureSourceRoots(sourceRoots); err != nil {
		return nil, fmt.Errorf("ensure source roots: %w", err)
	}

	// Wipe existing indexed data
	if err := idx.Store.DeleteAllFiles(); err != nil {
		return nil, fmt.Errorf("delete all files: %w", err)
	}

	stats := &IndexStats{}
	for _, root := range sourceRoots {
		absRoot := root
		if root == "." {
			absRoot = idx.RepoRoot
		} else if !filepath.IsAbs(root) {
			absRoot = filepath.Join(idx.RepoRoot, root)
		}

		if err := filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if info != nil && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				if excludedDirs[info.Name()] || (strings.HasPrefix(info.Name(), ".") && info.Name() != ".") {
					return filepath.SkipDir
				}
				return nil
			}
			lang := DetectLang(path)
			if lang == LangUnknown {
				return nil
			}
			if err := idx.indexFile(path, lang, info, stats); err != nil {
				stats.Errors++
			}
			return nil
		}); err != nil {
			return stats, fmt.Errorf("walk %s: %w", absRoot, err)
		}
	}
	return stats, nil
}

// Reconcile performs an incremental reconciliation:
// - indexes new/changed files (by SHA)
// - removes entries for deleted files
func (idx *Indexer) Reconcile(sourceRoots []string) (*IndexStats, error) {
	if err := idx.Store.EnsureProject(idx.RepoRoot); err != nil {
		return nil, err
	}

	// Get existing file→SHA map from DB
	existing, err := idx.Store.AllFiles()
	if err != nil {
		return nil, fmt.Errorf("load existing files: %w", err)
	}

	seen := map[string]bool{}
	stats := &IndexStats{}

	for _, root := range sourceRoots {
		absRoot := root
		if root == "." {
			absRoot = idx.RepoRoot
		} else if !filepath.IsAbs(root) {
			absRoot = filepath.Join(idx.RepoRoot, root)
		}

		_ = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if info != nil && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				if excludedDirs[info.Name()] || (strings.HasPrefix(info.Name(), ".") && info.Name() != ".") {
					return filepath.SkipDir
				}
				return nil
			}
			lang := DetectLang(path)
			if lang == LangUnknown {
				return nil
			}

			relPath, _ := filepath.Rel(idx.RepoRoot, path)
			seen[relPath] = true

			// Compute SHA
			sha, err := fileSHA256(path)
			if err != nil {
				return nil
			}

			if oldSHA, exists := existing[relPath]; exists && oldSHA == sha {
				stats.Skipped++
				return nil // unchanged
			}

			if err := idx.indexFile(path, lang, info, stats); err != nil {
				stats.Errors++
			}
			return nil
		})
	}

	// Remove stale entries
	for p := range existing {
		if !seen[p] {
			_ = idx.Store.DeleteFile(p)
			stats.Deleted++
		}
	}

	return stats, nil
}

// IndexSingleFile indexes (or re-indexes) a single file by absolute path.
// Returns true if the file was actually re-indexed (SHA changed or new).
func (idx *Indexer) IndexSingleFile(absPath string) (bool, error) {
	lang := DetectLang(absPath)
	if lang == LangUnknown {
		return false, nil
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return false, err
	}

	relPath, err := filepath.Rel(idx.RepoRoot, absPath)
	if err != nil {
		return false, err
	}

	sha, err := fileSHA256(absPath)
	if err != nil {
		return false, err
	}

	existing, err := idx.Store.GetFile(relPath)
	if err != nil {
		return false, err
	}
	if existing != nil && existing.SHA256 == sha {
		return false, nil // unchanged
	}

	stats := &IndexStats{}
	if err := idx.indexFile(absPath, lang, info, stats); err != nil {
		return false, err
	}
	return true, nil
}

// RemoveSingleFile removes a file from the index by absolute path.
func (idx *Indexer) RemoveSingleFile(absPath string) error {
	relPath, err := filepath.Rel(idx.RepoRoot, absPath)
	if err != nil {
		return err
	}
	return idx.Store.DeleteFile(relPath)
}

// indexFile reads, hashes, parses and stores a single file.
func (idx *Indexer) indexFile(absPath string, lang Lang, info os.FileInfo, stats *IndexStats) error {
	src, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	sha := sha256Bytes(src)
	relPath, err := filepath.Rel(idx.RepoRoot, absPath)
	if err != nil {
		return err
	}

	parser := GetParser(lang)
	if parser == nil {
		return nil
	}

	result, err := parser.Parse(absPath, src)
	if err != nil {
		return fmt.Errorf("parse %s: %w", relPath, err)
	}

	if err := idx.Store.UpsertFile(relPath, lang, sha, info.Size(), info.ModTime().Unix(), result); err != nil {
		return fmt.Errorf("upsert %s: %w", relPath, err)
	}

	stats.Indexed++
	stats.Symbols += len(result.Symbols)
	stats.Refs += len(result.Refs)
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sha256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// IndexStats holds statistics about an indexing run.
type IndexStats struct {
	Indexed int
	Skipped int
	Deleted int
	Errors  int
	Symbols int
	Refs    int
}

func (s *IndexStats) String() string {
	return fmt.Sprintf("indexed=%d skipped=%d deleted=%d errors=%d symbols=%d refs=%d",
		s.Indexed, s.Skipped, s.Deleted, s.Errors, s.Symbols, s.Refs)
}
