package memory

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeintelx/cli/internal/db"
)

// setupTestDB creates a temporary SQLite database with all migrations applied.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := db.Initialize(dbPath); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	// Create a project
	_, err = d.Exec(`INSERT INTO projects (repo_root) VALUES (?)`, "/tmp/test-repo")
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	return d
}

// --- Frontmatter tests ---

func TestParseMarkdownValid(t *testing.T) {
	md := `---
codeintelx:
  id: "abc123"
  scope: "file"
  file: "internal/foo.go"
  title: "Foo explanation"
  status: "active"
  fileStatus: "active"
  symbols:
    - language: "go"
      name: "FooFunc"
      status: "active"
---

This is the body.
`
	meta, body, err := ParseMarkdown([]byte(md))
	if err != nil {
		t.Fatalf("ParseMarkdown: %v", err)
	}
	if meta.ID != "abc123" {
		t.Errorf("ID = %q, want %q", meta.ID, "abc123")
	}
	if meta.Scope != "file" {
		t.Errorf("Scope = %q, want %q", meta.Scope, "file")
	}
	if meta.File != "internal/foo.go" {
		t.Errorf("File = %q, want %q", meta.File, "internal/foo.go")
	}
	if meta.Title != "Foo explanation" {
		t.Errorf("Title = %q, want %q", meta.Title, "Foo explanation")
	}
	if len(meta.Symbols) != 1 {
		t.Fatalf("Symbols len = %d, want 1", len(meta.Symbols))
	}
	if meta.Symbols[0].Name != "FooFunc" {
		t.Errorf("Symbol name = %q, want %q", meta.Symbols[0].Name, "FooFunc")
	}
	if !strings.Contains(body, "This is the body.") {
		t.Errorf("body = %q, want to contain %q", body, "This is the body.")
	}
}

func TestParseMarkdownNoFrontmatter(t *testing.T) {
	md := "# Just a normal markdown\n\nHello world."
	_, _, err := ParseMarkdown([]byte(md))
	if err == nil {
		t.Fatal("expected error for no frontmatter, got nil")
	}
}

func TestParseMarkdownMissingID(t *testing.T) {
	md := `---
codeintelx:
  scope: "project"
---

Body text.
`
	_, _, err := ParseMarkdown([]byte(md))
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestWriteMarkdownRoundTrip(t *testing.T) {
	meta := &CodeintelxMeta{
		ID:         "test-id-1",
		Scope:      "project",
		Title:      "Test Title",
		Status:     "active",
		FileStatus: "active",
		Symbols: []SymbolRef{
			{Language: "go", Name: "TestFunc", Status: "active"},
		},
	}
	body := "Some body content here."

	data, err := WriteMarkdown(meta, body)
	if err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}

	// Parse it back
	parsed, parsedBody, err := ParseMarkdown(data)
	if err != nil {
		t.Fatalf("ParseMarkdown round-trip: %v", err)
	}
	if parsed.ID != meta.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, meta.ID)
	}
	if parsed.Title != meta.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, meta.Title)
	}
	if len(parsed.Symbols) != 1 || parsed.Symbols[0].Name != "TestFunc" {
		t.Errorf("Symbols not preserved")
	}
	if !strings.Contains(parsedBody, body) {
		t.Errorf("body = %q, want to contain %q", parsedBody, body)
	}
}

// --- Ngram tests ---

func TestTrigramsBasic(t *testing.T) {
	grams := Trigrams("hello world")
	if len(grams) == 0 {
		t.Fatal("expected trigrams, got none")
	}
	// "hello world" normalized is "hello world" (11 chars) → 9 unique trigrams
	// "hel", "ell", "llo", "lo ", "o w", " wo", "wor", "orl", "rld"
	expected := map[string]bool{
		"hel": true, "ell": true, "llo": true, "lo ": true,
		"o w": true, " wo": true, "wor": true, "orl": true, "rld": true,
	}
	for _, g := range grams {
		if !expected[g] {
			t.Errorf("unexpected trigram %q", g)
		}
	}
}

func TestTrigramsDedup(t *testing.T) {
	grams := Trigrams("aaa aaa")
	// "aaa aaa" normalized is "aaa aaa"
	// "aaa", "aa ", "a a", " aa", "aaa" — "aaa" appears twice, should be deduped
	seen := map[string]int{}
	for _, g := range grams {
		seen[g]++
	}
	for g, count := range seen {
		if count > 1 {
			t.Errorf("trigram %q appears %d times, expected 1", g, count)
		}
	}
}

func TestTrigramsShortInput(t *testing.T) {
	grams := Trigrams("ab")
	if len(grams) != 1 || grams[0] != "ab" {
		t.Errorf("short input: got %v, want [\"ab\"]", grams)
	}

	grams = Trigrams("")
	if len(grams) != 0 {
		t.Errorf("empty input: got %v, want empty", grams)
	}
}

// --- Store + Manager integration tests ---

func TestManagerAppendReadDelete(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	mgr := NewManager(d, 1, repoRoot, memDir)

	// Append
	elem, err := mgr.Append("project", "", "Test Memory", "This is a test body.", nil)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if elem.Meta.ID == "" {
		t.Fatal("expected non-empty memory ID")
	}
	if elem.Meta.Scope != "project" {
		t.Errorf("Scope = %q, want %q", elem.Meta.Scope, "project")
	}

	// File should exist on disk
	if _, err := os.Stat(elem.AbsPath); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	// Read
	read, err := mgr.Read(elem.Meta.ID)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if read.Meta.Title != "Test Memory" {
		t.Errorf("Title = %q, want %q", read.Meta.Title, "Test Memory")
	}
	if !strings.Contains(read.Body, "This is a test body.") {
		t.Errorf("Body = %q, want to contain test body", read.Body)
	}

	// Delete (soft)
	if err := mgr.Delete(elem.Meta.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// File should still exist on disk
	if _, err := os.Stat(elem.AbsPath); err != nil {
		t.Fatalf("file should still exist after soft delete: %v", err)
	}

	// Read should still work (but status is deleted)
	read2, err := mgr.Read(elem.Meta.ID)
	if err != nil {
		t.Fatalf("Read after delete: %v", err)
	}
	if read2.Meta.Status != "deleted" {
		t.Errorf("Status = %q, want %q", read2.Meta.Status, "deleted")
	}
}

func TestManagerUpdate(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	mgr := NewManager(d, 1, repoRoot, memDir)

	elem, err := mgr.Append("project", "", "Original Title", "Original body.", nil)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	newTitle := "Updated Title"
	newBody := "Updated body."
	updated, err := mgr.Update(elem.Meta.ID, &newTitle, &newBody, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Meta.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", updated.Meta.Title, "Updated Title")
	}
	if !strings.Contains(updated.Body, "Updated body.") {
		t.Errorf("Body not updated")
	}
}

func TestManagerSearch(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	mgr := NewManager(d, 1, repoRoot, memDir)

	_, err := mgr.Append("project", "", "Authentication Flow", "How the auth flow works with JWT tokens and refresh.", nil)
	if err != nil {
		t.Fatalf("Append 1: %v", err)
	}
	_, err = mgr.Append("project", "", "Database Schema", "The database uses PostgreSQL with migrations.", nil)
	if err != nil {
		t.Fatalf("Append 2: %v", err)
	}

	// Search for auth-related
	results, err := mgr.Search("authentication JWT tokens", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}
	// The auth memory should score higher than the database one
	if results[0].Title != "Authentication Flow" {
		t.Errorf("top result title = %q, want %q", results[0].Title, "Authentication Flow")
	}
}

func TestManagerSearchSkipsDeleted(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	mgr := NewManager(d, 1, repoRoot, memDir)

	elem, err := mgr.Append("project", "", "Secret Memory", "This memory should not appear in search.", nil)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Delete it
	if err := mgr.Delete(elem.Meta.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Search should not find it
	results, err := mgr.Search("secret memory", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.MemoryUID == elem.Meta.ID {
			t.Errorf("deleted memory appeared in search results")
		}
	}
}

func TestManagerGrepReplace(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	mgr := NewManager(d, 1, repoRoot, memDir)

	elem, err := mgr.Append("project", "", "GrepTest", "The foo bar baz and foo again.", nil)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	result, err := mgr.GrepReplace(elem.Meta.ID, "", "foo", "qux")
	if err != nil {
		t.Fatalf("GrepReplace: %v", err)
	}
	if result.Replacements != 2 {
		t.Errorf("Replacements = %d, want 2", result.Replacements)
	}

	// Read back and verify
	read, err := mgr.Read(elem.Meta.ID)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if strings.Contains(read.Body, "foo") {
		t.Errorf("body still contains 'foo' after replace")
	}
	if !strings.Contains(read.Body, "qux") {
		t.Errorf("body does not contain 'qux' after replace")
	}
}

func TestManagerBulkIndexAndReconcile(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file-scoped memory referencing a file that doesn't exist
	meta := &CodeintelxMeta{
		ID:         "reconcile-test",
		Scope:      "file",
		File:       "nonexistent/file.go",
		Title:      "Ghost file",
		Status:     "active",
		FileStatus: "active",
	}
	data, err := WriteMarkdown(meta, "Body about a file that doesn't exist.")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "file-ghost.md"), data, 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(d, 1, repoRoot, memDir)

	// Bulk index
	if err := mgr.BulkIndex(); err != nil {
		t.Fatalf("BulkIndex: %v", err)
	}

	// Reconcile — should mark the file as deleted
	if err := mgr.Reconcile(); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// Read back — fileStatus should be "deleted"
	elem, err := mgr.Read("reconcile-test")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if elem.Meta.FileStatus != "deleted" {
		t.Errorf("FileStatus = %q, want %q", elem.Meta.FileStatus, "deleted")
	}

	// Search should NOT return it
	results, err := mgr.Search("ghost file", "", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range results {
		if r.MemoryUID == "reconcile-test" {
			t.Errorf("memory with deleted file appeared in search results")
		}
	}
}

func TestSalvageMerge(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a markdown file without proper frontmatter
	badMd := "# Just a Title\n\nSome content without frontmatter."
	if err := os.WriteFile(filepath.Join(memDir, "bad.md"), []byte(badMd), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(d, 1, repoRoot, memDir)

	// Bulk index should merge the bad file into project.md
	if err := mgr.BulkIndex(); err != nil {
		t.Fatalf("BulkIndex: %v", err)
	}

	// project.md should exist and contain the merged content
	projectMd, err := os.ReadFile(filepath.Join(memDir, "project.md"))
	if err != nil {
		t.Fatalf("project.md not created: %v", err)
	}
	if !strings.Contains(string(projectMd), "Imported (unparseable frontmatter): bad.md") {
		t.Errorf("project.md does not contain expected import section")
	}
	if !strings.Contains(string(projectMd), "Just a Title") {
		t.Errorf("project.md does not contain original content")
	}
}

func TestManagerFileScoped(t *testing.T) {
	d := setupTestDB(t)
	repoRoot := t.TempDir()
	memDir := filepath.Join(repoRoot, "memory")

	// Create the referenced file so reconcile doesn't mark it deleted
	repoFile := filepath.Join(repoRoot, "src", "main.go")
	if err := os.MkdirAll(filepath.Dir(repoFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(repoFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(d, 1, repoRoot, memDir)

	elem, err := mgr.Append("file", "src/main.go", "Main explanation", "How the main function works.", nil)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if elem.Meta.Scope != "file" {
		t.Errorf("Scope = %q, want %q", elem.Meta.Scope, "file")
	}
	if elem.Meta.File != "src/main.go" {
		t.Errorf("File = %q, want %q", elem.Meta.File, "src/main.go")
	}

	// Search by file filter
	results, err := mgr.Search("main function", "file", "src/main.go", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results for file-scoped memory")
	}
}
