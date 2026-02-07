package repo

import (
	"fmt"
	"path/filepath"
	"strings"
)

// IsParentOrChild checks if path1 is a parent or child of path2.
// Returns true if one is contained within the other.
func IsParentOrChild(path1, path2 string) bool {
	// Normalize paths
	abs1, err1 := filepath.Abs(path1)
	abs2, err2 := filepath.Abs(path2)
	if err1 != nil || err2 != nil {
		// If we can't get absolute paths, do string comparison
		return strings.HasPrefix(path1+string(filepath.Separator), path2+string(filepath.Separator)) ||
			strings.HasPrefix(path2+string(filepath.Separator), path1+string(filepath.Separator))
	}

	// Check if one is a prefix of the other
	rel, err := filepath.Rel(abs1, abs2)
	if err != nil {
		return false
	}

	// If relative path doesn't start with "..", then abs2 is inside abs1
	if !strings.HasPrefix(rel, "..") && rel != "." {
		return true
	}

	// Check the reverse
	rel2, err := filepath.Rel(abs2, abs1)
	if err != nil {
		return false
	}

	// If relative path doesn't start with "..", then abs1 is inside abs2
	return !strings.HasPrefix(rel2, "..") && rel2 != "."
}

// ValidateSelectedDirs validates that selected directories don't have parent/child relationships
// and are unique. Returns an error if validation fails.
// The repoRoot parameter is the absolute path to the repository root.
// Selected directories can be "." (for root) or relative paths like "src", "cmd", etc.
func ValidateSelectedDirs(repoRoot string, selectedDirs []string) error {
	// Normalize all selected directories to absolute paths
	absPaths := make([]string, 0, len(selectedDirs))
	seen := make(map[string]bool)

	for _, dir := range selectedDirs {
		var abs string
		if dir == "." {
			// Root directory
			var err error
			abs, err = filepath.Abs(repoRoot)
			if err != nil {
				abs = repoRoot
			}
		} else {
			// Subdirectory - join with repo root
			normalized := filepath.Join(repoRoot, dir)
			var err error
			abs, err = filepath.Abs(normalized)
			if err != nil {
				abs = normalized
			}
		}

		// Check for duplicates
		if seen[abs] {
			return fmt.Errorf("duplicate directory selected: %s", dir)
		}
		seen[abs] = true
		absPaths = append(absPaths, abs)
	}

	// Check for parent/child relationships
	for i, abs1 := range absPaths {
		for j, abs2 := range absPaths {
			if i >= j {
				continue
			}

			if IsParentOrChild(abs1, abs2) {
				// Map back to original directory names for error message
				dir1 := selectedDirs[i]
				dir2 := selectedDirs[j]
				if dir1 == "." {
					dir1 = "repository root"
				}
				if dir2 == "." {
					dir2 = "repository root"
				}
				return fmt.Errorf("directories cannot be parent/child of each other: %s and %s", dir1, dir2)
			}
		}
	}

	return nil
}
