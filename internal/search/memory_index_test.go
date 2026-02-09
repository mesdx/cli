package search

import (
	"testing"
)

func TestMemoryIndexBasicSearch(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewMemoryIndex(1, dir)
	if err != nil {
		t.Fatalf("NewMemoryIndex: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// Index a document.
	err = idx.IndexMemory(
		"uid-1", "project", "", "project-auth.md", "Authentication Flow",
		"active", "active",
		"How the auth flow works with JWT tokens and refresh.",
		nil,
	)
	if err != nil {
		t.Fatalf("IndexMemory: %v", err)
	}

	// Search for it.
	hits, err := idx.Search("authentication JWT tokens", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected search results, got none")
	}
	if hits[0].MemoryUID != "uid-1" {
		t.Errorf("top hit uid = %q, want %q", hits[0].MemoryUID, "uid-1")
	}
}

func TestMemoryIndexExcludesDeleted(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewMemoryIndex(1, dir)
	if err != nil {
		t.Fatalf("NewMemoryIndex: %v", err)
	}
	defer func() { _ = idx.Close() }()

	err = idx.IndexMemory(
		"uid-del", "project", "", "project-del.md", "Deleted Memory",
		"deleted", "active",
		"This should not appear.",
		nil,
	)
	if err != nil {
		t.Fatalf("IndexMemory: %v", err)
	}

	hits, err := idx.Search("deleted memory", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, h := range hits {
		if h.MemoryUID == "uid-del" {
			t.Error("deleted memory appeared in search results")
		}
	}
}

func TestMemoryIndexFileScoped(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewMemoryIndex(1, dir)
	if err != nil {
		t.Fatalf("NewMemoryIndex: %v", err)
	}
	defer func() { _ = idx.Close() }()

	err = idx.IndexMemory(
		"uid-file", "file", "src/main.go", "file-main.md", "Main explanation",
		"active", "active",
		"How the main function works.",
		nil,
	)
	if err != nil {
		t.Fatalf("IndexMemory: %v", err)
	}

	hits, err := idx.Search("main function", "file", "src/main.go", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected file-scoped search results")
	}
}

func TestMemoryIndexMultiHeaderChunking(t *testing.T) {
	dir := t.TempDir()
	idx, err := NewMemoryIndex(1, dir)
	if err != nil {
		t.Fatalf("NewMemoryIndex: %v", err)
	}
	defer func() { _ = idx.Close() }()

	body := `## Overview

This document covers authentication.

### JWT Handling

The system uses RS256 signed JWT tokens.

## Database

PostgreSQL stores user data.`

	err = idx.IndexMemory(
		"uid-multi", "project", "", "project-multi.md", "Multi Section Doc",
		"active", "active",
		body,
		nil,
	)
	if err != nil {
		t.Fatalf("IndexMemory: %v", err)
	}

	// Search for JWT content (in a later chunk)
	hits, err := idx.Search("JWT RS256 tokens", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected search results from later chunk, got none")
	}
	if hits[0].MemoryUID != "uid-multi" {
		t.Errorf("top hit uid = %q, want %q", hits[0].MemoryUID, "uid-multi")
	}
}

// --- Chunking unit tests ---

func TestChunkByHeadersAnyLevel(t *testing.T) {
	body := `## License Overview

This project is licensed under the FSL.

### Key Points

- Copyright: 2026 CodeIntelX contributors

## Permitted Uses

The software can be used for any purpose.`

	chunks := ChunkByHeaders(body)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "License Overview" {
		t.Errorf("chunk[0].Heading = %q, want %q", chunks[0].Heading, "License Overview")
	}
	if chunks[1].Heading != "Key Points" {
		t.Errorf("chunk[1].Heading = %q, want %q", chunks[1].Heading, "Key Points")
	}
	if chunks[2].Heading != "Permitted Uses" {
		t.Errorf("chunk[2].Heading = %q, want %q", chunks[2].Heading, "Permitted Uses")
	}
}

func TestChunkByHeadersWithPreamble(t *testing.T) {
	body := `Some preamble text.

# Title

Content under title.`

	chunks := ChunkByHeaders(body)
	if len(chunks) < 2 {
		t.Fatalf("expected >= 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "" {
		t.Errorf("first chunk should be preamble (empty heading), got %q", chunks[0].Heading)
	}
	if chunks[1].Heading != "Title" {
		t.Errorf("second chunk heading = %q, want %q", chunks[1].Heading, "Title")
	}
}

func TestChunkByHeadersNoPreamble(t *testing.T) {
	body := `# First

Content A.

## Second

Content B.`

	chunks := ChunkByHeaders(body)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "First" {
		t.Errorf("chunk[0].Heading = %q, want %q", chunks[0].Heading, "First")
	}
	if chunks[1].Heading != "Second" {
		t.Errorf("chunk[1].Heading = %q, want %q", chunks[1].Heading, "Second")
	}
}

func TestChunkByHeadersEmptyBody(t *testing.T) {
	chunks := ChunkByHeaders("")
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty body, got %d", len(chunks))
	}
}

func TestChunkByHeadersNoHeaders(t *testing.T) {
	body := "Just plain text without any headers."
	chunks := ChunkByHeaders(body)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Heading != "" {
		t.Errorf("expected preamble chunk (empty heading), got %q", chunks[0].Heading)
	}
}

func TestChunkByHeadersH6(t *testing.T) {
	body := `###### Deep Header

Deep content.`

	chunks := ChunkByHeaders(body)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Heading != "Deep Header" {
		t.Errorf("heading = %q, want %q", chunks[0].Heading, "Deep Header")
	}
}

func TestChunkByHeadersNotHeader(t *testing.T) {
	// "####### " (7 hashes) should NOT be treated as a header.
	body := `####### Not a header

Some text.`

	chunks := ChunkByHeaders(body)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk (no header split), got %d", len(chunks))
	}
	if chunks[0].Heading != "" {
		t.Errorf("expected preamble, got heading %q", chunks[0].Heading)
	}
}
