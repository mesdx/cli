package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverAllDirs recursively discovers all directories in the repo root,
// excluding common ignored directories and hidden directories.
func DiscoverAllDirs(repoRoot string) ([]string, error) {
	var dirs []string
	excluded := map[string]bool{
		".git":            true,
		codeintelxDirName: true,
		"node_modules":    true,
		".venv":           true,
		"venv":            true,
		".env":            true,
		"vendor":          true,
		"target":          true,
		"build":            true,
		"dist":             true,
		".idea":            true,
		".vscode":          true,
		".DS_Store":        true,
	}

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		// Get relative path from repo root
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return nil
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Check if any component of the path is excluded
		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			if excluded[part] {
				// Skip this directory and all its children
				return filepath.SkipDir
			}
			// Skip hidden directories
			if strings.HasPrefix(part, ".") && part != "." {
				return filepath.SkipDir
			}
		}

		dirs = append(dirs, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return dirs, nil
}
