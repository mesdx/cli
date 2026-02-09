package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// memoryManifest stores the docIDs created for a given memory markdown file.
// This lets us delete old chunk-docs deterministically when a file is updated or removed.
type memoryManifest struct {
	ByMdRelPath map[string][]string `json:"byMdRelPath"`
}

func loadMemoryManifest(path string) (*memoryManifest, error) {
	m := &memoryManifest{ByMdRelPath: map[string][]string{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	if len(data) == 0 {
		return m, nil
	}
	if err := json.Unmarshal(data, m); err != nil {
		// If manifest is corrupt, start fresh (bulk index will rebuild).
		return &memoryManifest{ByMdRelPath: map[string][]string{}}, nil
	}
	if m.ByMdRelPath == nil {
		m.ByMdRelPath = map[string][]string{}
	}
	return m, nil
}

func saveMemoryManifest(path string, m *memoryManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write manifest tmp: %w", err)
	}
	return os.Rename(tmp, path)
}
