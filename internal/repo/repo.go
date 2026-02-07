package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const codeintelxDirName = ".codeintelx"

// FindRoot finds the repository root directory.
// It first checks if .git exists in the current directory or any parent.
// If not found, it uses the current working directory as the repo root.
func FindRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Walk up the directory tree looking for .git
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, use cwd as repo root
			break
		}
		dir = parent
	}

	// No .git found, use current working directory
	return cwd, nil
}

// CodeintelxDir returns the path to the .codeintelx directory for a given repo root.
func CodeintelxDir(repoRoot string) string {
	return filepath.Join(repoRoot, codeintelxDirName)
}

// DiscoverSubdirs discovers immediate subdirectories in the repo root,
// excluding common ignored directories.
func DiscoverSubdirs(repoRoot string) ([]string, error) {
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var subdirs []string
	excluded := map[string]bool{
		".git":            true,
		codeintelxDirName: true,
		"node_modules":    true,
		".venv":           true,
		"venv":            true,
		".env":            true,
		"vendor":          true,
		"target":          true,
		"build":           true,
		"dist":            true,
		".idea":           true,
		".vscode":         true,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if excluded[name] {
			continue
		}

		// Skip hidden directories (except .codeintelx which is already excluded)
		if strings.HasPrefix(name, ".") {
			continue
		}

		subdirs = append(subdirs, name)
	}

	return subdirs, nil
}
