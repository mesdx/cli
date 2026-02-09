package search

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// BleveIndex is a thin wrapper over a Bleve index stored on disk.
type BleveIndex struct {
	Path  string
	Index bleve.Index
}

// OpenOrCreate opens an existing Bleve index at path, or creates a new one with the given mapping.
func OpenOrCreate(path string, m mapping.IndexMapping) (*BleveIndex, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("mkdir index dir: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		idx, err := bleve.Open(path)
		if err != nil {
			return nil, fmt.Errorf("bleve open: %w", err)
		}
		return &BleveIndex{Path: path, Index: idx}, nil
	}

	idx, err := bleve.New(path, m)
	if err != nil {
		return nil, fmt.Errorf("bleve new: %w", err)
	}
	return &BleveIndex{Path: path, Index: idx}, nil
}

// OpenReadOnly opens an existing Bleve index at path in read-only mode.
// This allows multiple processes to read from the index simultaneously.
func OpenReadOnly(path string) (*BleveIndex, error) {
	idx, err := bleve.OpenUsing(path, map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		return nil, fmt.Errorf("bleve open read-only: %w", err)
	}
	return &BleveIndex{Path: path, Index: idx}, nil
}

// Reset deletes the index directory and recreates an empty index with the given mapping.
func Reset(path string, m mapping.IndexMapping) (*BleveIndex, error) {
	_ = os.RemoveAll(path)
	return OpenOrCreate(path, m)
}

func (b *BleveIndex) Close() error {
	if b == nil || b.Index == nil {
		return nil
	}
	return b.Index.Close()
}
