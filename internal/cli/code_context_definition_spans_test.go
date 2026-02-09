package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesdx/cli/internal/indexer"
)

// testdataRoot returns the absolute path to internal/indexer/testdata.
func testdataRoot(t *testing.T) string {
	t.Helper()
	// We're in internal/cli, go up two levels + down into internal/indexer/testdata
	wd, _ := os.Getwd()
	root := filepath.Join(wd, "..", "indexer", "testdata")
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("testdata dir not found at %s", root)
	}
	return root
}

// ---------- Go: full struct body in fetched code ----------

func TestFetchDefinitionsCode_GoStruct(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "Config",
			Kind: "struct",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 7,
				StartCol:  5,
				EndLine:   11,
				EndCol:    11,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain the doc comments (expanded backward)
	if !strings.Contains(code, "Config holds application configuration") {
		t.Error("expected doc comment in output")
	}
	// Should contain the full struct body
	if !strings.Contains(code, "Host") {
		t.Error("expected field Host in output")
	}
	if !strings.Contains(code, "Port") {
		t.Error("expected field Port in output")
	}
	if !strings.Contains(code, "LogLevel") {
		t.Error("expected field LogLevel in output")
	}
	// Count lines — should be more than 1
	lineCount := strings.Count(code, "\n")
	if lineCount < 5 {
		t.Errorf("expected at least 5 output lines for struct + doc, got %d", lineCount)
	}
}

// ---------- Go: full function body in fetched code ----------

func TestFetchDefinitionsCode_GoFunction(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "NewConfig",
			Kind: "function",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 37,
				StartCol:  5,
				EndLine:   43,
				EndCol:    14,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain doc comment
	if !strings.Contains(code, "NewConfig creates a Config") {
		t.Error("expected doc comment in output")
	}
	// Should contain function body
	if !strings.Contains(code, "return &Config{") {
		t.Error("expected return statement in function body")
	}
	if !strings.Contains(code, "DefaultPort") {
		t.Error("expected DefaultPort in function body")
	}
}

// ---------- Go: method body in fetched code ----------

func TestFetchDefinitionsCode_GoMethod(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "Validate",
			Kind: "method",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 14,
				StartCol:  18,
				EndLine:   22,
				EndCol:    26,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain doc comment
	if !strings.Contains(code, "Validate checks if the config") {
		t.Error("expected doc comment in output")
	}
	// Should contain method body
	if !strings.Contains(code, "host is required") {
		t.Error("expected error message in method body")
	}
	if !strings.Contains(code, "return nil") {
		t.Error("expected return nil in method body")
	}
}

// ---------- Java: full class body ----------

func TestFetchDefinitionsCode_JavaClass(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "AppConfig",
			Kind: "class",
			Location: indexer.Location{
				Path:      "java/DefsA.java",
				StartLine: 9,
				StartCol:  13,
				EndLine:   44,
				EndCol:    22,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain Javadoc (expanded backward)
	if !strings.Contains(code, "AppConfig holds application") {
		t.Error("expected Javadoc in output")
	}
	// Should contain full class body
	if !strings.Contains(code, "getHost") {
		t.Error("expected getHost method in class body")
	}
	if !strings.Contains(code, "validate") {
		t.Error("expected validate method in class body")
	}
}

// ---------- Rust: struct with doc comments ----------

func TestFetchDefinitionsCode_RustStruct(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "AppConfig",
			Kind: "struct",
			Location: indexer.Location{
				Path:      "rust/defs_a.rs",
				StartLine: 5,
				StartCol:  11,
				EndLine:   9,
				EndCol:    20,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain doc comments
	if !strings.Contains(code, "AppConfig holds application") {
		t.Error("expected doc comment in output")
	}
	// Should contain struct fields
	if !strings.Contains(code, "host: String") {
		t.Error("expected host field in struct")
	}
	if !strings.Contains(code, "port: u16") {
		t.Error("expected port field in struct")
	}
}

// ---------- Python: class with methods ----------

func TestFetchDefinitionsCode_PythonClass(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "AppConfig",
			Kind: "class",
			Location: indexer.Location{
				Path:      "python/defs_a.py",
				StartLine: 11,
				StartCol:  6,
				EndLine:   27,
				EndCol:    15,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain comment above (doc expansion)
	if !strings.Contains(code, "AppConfig holds") {
		t.Error("expected doc comment in output")
	}
	// Should contain class body (multiple methods)
	if !strings.Contains(code, "def __init__") {
		t.Error("expected __init__ in class body")
	}
	if !strings.Contains(code, "def validate") {
		t.Error("expected validate in class body")
	}
	if !strings.Contains(code, "@property") {
		t.Error("expected @property decorator in class body")
	}
}

// ---------- Python: standalone function ----------

func TestFetchDefinitionsCode_PythonFunction(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "say_hello",
			Kind: "function",
			Location: indexer.Location{
				Path:      "python/defs_a.py",
				StartLine: 38,
				StartCol:  4,
				EndLine:   41,
				EndCol:    13,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain comment above
	if !strings.Contains(code, "say_hello is a standalone") {
		t.Error("expected doc comment in output")
	}
	// Should contain function body
	if !strings.Contains(code, "greeting") {
		t.Error("expected function body content")
	}
}

// ---------- TypeScript: class with full body ----------

func TestFetchDefinitionsCode_TSClass(t *testing.T) {
	repoRoot := testdataRoot(t)

	// AppConfig class: line 11 is "export class AppConfig {", closes at line 33
	results := []indexer.DefinitionResult{
		{
			Name: "AppConfig",
			Kind: "class",
			Location: indexer.Location{
				Path:      "ts/defs_a.ts",
				StartLine: 11,
				StartCol:  13,
				EndLine:   33,
				EndCol:    22,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Should contain JSDoc (expanded backward from line 11 to line 7)
	if !strings.Contains(code, "AppConfig holds application") {
		t.Error("expected JSDoc in output")
	}
	// Should contain class body
	if !strings.Contains(code, "validate()") {
		t.Error("expected validate method in class body")
	}
	if !strings.Contains(code, "get address()") {
		t.Error("expected getter in class body")
	}
}

// ---------- Multiple definitions concatenated ----------

func TestFetchDefinitionsCode_MultipleDefinitions(t *testing.T) {
	repoRoot := testdataRoot(t)

	results := []indexer.DefinitionResult{
		{
			Name: "Config",
			Kind: "struct",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 7,
				StartCol:  5,
				EndLine:   11,
				EndCol:    11,
			},
		},
		{
			Name: "Formatter",
			Kind: "interface",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 25,
				StartCol:  5,
				EndLine:   28,
				EndCol:    14,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Both definitions should be present
	if !strings.Contains(code, "Config holds") {
		t.Error("expected Config doc in output")
	}
	if !strings.Contains(code, "Formatter formats") {
		t.Error("expected Formatter doc in output")
	}
	// Should have separator between them
	if !strings.Contains(code, "\n\n") {
		t.Error("expected blank line separator between definitions")
	}
}

// ---------- Header shows correct line range with doc expansion ----------

func TestFetchDefinitionsCode_HeaderShowsDocRange(t *testing.T) {
	repoRoot := testdataRoot(t)

	// Config struct starts at line 7, but doc starts at line 5
	results := []indexer.DefinitionResult{
		{
			Name: "Config",
			Kind: "struct",
			Location: indexer.Location{
				Path:      "go/defs_a.go",
				StartLine: 7,
				StartCol:  5,
				EndLine:   11,
				EndCol:    11,
			},
		},
	}

	code, err := fetchDefinitionsCode(repoRoot, results)
	if err != nil {
		t.Fatalf("fetchDefinitionsCode: %v", err)
	}

	// Header should show expanded range: 5-11 (doc starts at line 5)
	if !strings.Contains(code, "--- go/defs_a.go:5-11") {
		t.Errorf("expected header with doc-expanded range, got:\n%s", code)
	}
}
