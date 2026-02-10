package treesitter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParserDir(t *testing.T) {
	// Test with env var set
	testDir := t.TempDir()
	os.Setenv("MESDX_PARSER_DIR", testDir)
	defer os.Unsetenv("MESDX_PARSER_DIR")

	dir, err := ParserDir()
	if err != nil {
		t.Fatalf("ParserDir() with env var: %v", err)
	}
	if dir != testDir {
		t.Errorf("ParserDir() = %q, want %q", dir, testDir)
	}
}

func TestParserDirNotFound(t *testing.T) {
	os.Unsetenv("MESDX_PARSER_DIR")
	
	// This will fail unless parsers are actually installed
	_, err := ParserDir()
	if err == nil {
		// If it succeeds, parsers are installed, which is fine for the test
		return
	}
	
	// Expected error when parsers aren't found
	if err.Error() == "" {
		t.Error("ParserDir() should return descriptive error when not found")
	}
}

func TestVerifyLanguages(t *testing.T) {
	// Create a temp directory with mock parser libs
	testDir := t.TempDir()
	os.Setenv("MESDX_PARSER_DIR", testDir)
	defer os.Unsetenv("MESDX_PARSER_DIR")

	// Create a mock parser file
	mockFile := filepath.Join(testDir, "libtree-sitter-go.so")
	if err := os.WriteFile(mockFile, []byte("mock"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail for missing languages
	err := VerifyLanguages([]string{"go", "python"})
	if err == nil {
		t.Error("VerifyLanguages() should fail when libraries are missing")
	}

	// Should succeed when all requested libs exist
	err = VerifyLanguages([]string{"go"})
	if err != nil {
		t.Errorf("VerifyLanguages() with existing lib: %v", err)
	}
}

func TestRequiredLanguages(t *testing.T) {
	langs := RequiredLanguages()
	expected := []string{"go", "java", "rust", "python", "javascript", "typescript"}
	
	if len(langs) != len(expected) {
		t.Errorf("RequiredLanguages() returned %d languages, want %d", len(langs), len(expected))
	}

	// Check all expected languages are present
	langMap := make(map[string]bool)
	for _, l := range langs {
		langMap[l] = true
	}

	for _, exp := range expected {
		if !langMap[exp] {
			t.Errorf("RequiredLanguages() missing %q", exp)
		}
	}
}
