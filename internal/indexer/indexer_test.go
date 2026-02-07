package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeintelx/cli/internal/db"
	"github.com/codeintelx/cli/internal/symbols"
)

// setupTestDB creates a temporary DB and returns a Store + cleanup func.
func setupTestDB(t *testing.T) (*Store, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := db.Initialize(dbPath); err != nil {
		t.Fatalf("db.Initialize: %v", err)
	}
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	store := &Store{DB: d}
	return store, func() { _ = d.Close() }
}

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	// Walk up to find testdata
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "testdata")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("testdata dir not found at %s", dir)
	}
	return dir
}

func TestGoParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "go", "sample.go"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := &GoParser{}
	result, err := parser.Parse("sample.go", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Check expected symbols
	expectSymbol(t, result, "MaxRetries", symbols.KindConstant)
	expectSymbol(t, result, "DefaultName", symbols.KindVariable)
	expectSymbol(t, result, "Greeter", symbols.KindInterface)
	expectSymbol(t, result, "Person", symbols.KindStruct)
	expectSymbol(t, result, "Name", symbols.KindField)
	expectSymbol(t, result, "Age", symbols.KindField)
	expectSymbol(t, result, "NewPerson", symbols.KindFunction)
	expectSymbol(t, result, "Greet", symbols.KindMethod)
	expectSymbol(t, result, "SayHello", symbols.KindFunction)

	// Check that we have references
	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
	expectRef(t, result, "NewPerson")
	expectRef(t, result, "DefaultName")
	expectRef(t, result, "MaxRetries")
}

func TestJavaParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "java", "Sample.java"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := &JavaParser{}
	result, err := parser.Parse("Sample.java", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	expectSymbol(t, result, "Sample", symbols.KindClass)
	expectSymbol(t, result, "Processor", symbols.KindInterface)
	expectSymbol(t, result, "Status", symbols.KindEnum)
	expectSymbol(t, result, "Worker", symbols.KindClass)
	expectSymbol(t, result, "getName", symbols.KindMethod)
	expectSymbol(t, result, "increment", symbols.KindMethod)
	expectSymbol(t, result, "process", symbols.KindMethod)

	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
}

func TestRustParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "rust", "sample.rs"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := &RustParser{}
	result, err := parser.Parse("sample.rs", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	expectSymbol(t, result, "Person", symbols.KindStruct)
	expectSymbol(t, result, "Status", symbols.KindEnum)
	expectSymbol(t, result, "Greeter", symbols.KindTrait)
	expectSymbol(t, result, "MAX_RETRIES", symbols.KindConstant)
	expectSymbol(t, result, "helpers", symbols.KindModule)
	expectSymbol(t, result, "PersonAlias", symbols.KindTypeAlias)
	expectSymbol(t, result, "new", symbols.KindMethod)
	expectSymbol(t, result, "say_hello", symbols.KindMethod)

	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
}

func TestPythonParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "python", "sample.py"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := &PythonParser{}
	result, err := parser.Parse("sample.py", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	expectSymbol(t, result, "Greeter", symbols.KindClass)
	expectSymbol(t, result, "Person", symbols.KindClass)
	expectSymbol(t, result, "say_hello", symbols.KindFunction)
	expectSymbol(t, result, "format_name", symbols.KindFunction)
	expectSymbol(t, result, "MAX_RETRIES", symbols.KindConstant)
	expectSymbol(t, result, "default_name", symbols.KindVariable)
	expectSymbol(t, result, "greet", symbols.KindMethod)
	expectSymbol(t, result, "display_name", symbols.KindProperty)

	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
}

func TestTypeScriptParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "ts", "sample.ts"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := &TypeScriptParser{}
	result, err := parser.Parse("sample.ts", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	expectSymbol(t, result, "Person", symbols.KindClass)
	expectSymbol(t, result, "Greeter", symbols.KindInterface)
	expectSymbol(t, result, "PersonOptions", symbols.KindTypeAlias)
	expectSymbol(t, result, "Status", symbols.KindEnum)
	expectSymbol(t, result, "sayHello", symbols.KindFunction)
	expectSymbol(t, result, "formatName", symbols.KindFunction)
	expectSymbol(t, result, "Utils", symbols.KindModule)
	expectSymbol(t, result, "constructor", symbols.KindConstructor)

	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
}

func TestJavaScriptParser(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "js", "sample.js"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	// JS uses the same parser as TS
	parser := &TypeScriptParser{}
	result, err := parser.Parse("sample.js", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	expectSymbol(t, result, "Person", symbols.KindClass)
	expectSymbol(t, result, "sayHello", symbols.KindFunction)
	expectSymbol(t, result, "formatName", symbols.KindFunction)
	expectSymbol(t, result, "constructor", symbols.KindConstructor)

	if len(result.Refs) == 0 {
		t.Error("expected at least some refs")
	}
}

func TestFullIndexAndReconcile(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Use testdata as a mini repo
	repoRoot := testdataDir(t)
	idx := &Indexer{Store: store, RepoRoot: repoRoot}

	// Full index
	stats, err := idx.FullIndex([]string{"."})
	if err != nil {
		t.Fatalf("FullIndex: %v", err)
	}

	if stats.Indexed == 0 {
		t.Error("FullIndex indexed 0 files")
	}
	if stats.Symbols == 0 {
		t.Error("FullIndex found 0 symbols")
	}
	if stats.Refs == 0 {
		t.Error("FullIndex found 0 refs")
	}

	// Reconcile should skip all (nothing changed)
	stats2, err := idx.Reconcile([]string{"."})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if stats2.Indexed != 0 {
		t.Errorf("Reconcile indexed %d files, want 0 (nothing changed)", stats2.Indexed)
	}
	if stats2.Skipped != stats.Indexed {
		t.Errorf("Reconcile skipped %d, want %d", stats2.Skipped, stats.Indexed)
	}
}

func TestIndexSingleFile(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repoRoot := testdataDir(t)
	idx := &Indexer{Store: store, RepoRoot: repoRoot}

	if err := idx.Store.EnsureProject(repoRoot); err != nil {
		t.Fatalf("EnsureProject: %v", err)
	}

	// Index single file
	absPath := filepath.Join(repoRoot, "go", "sample.go")
	reindexed, err := idx.IndexSingleFile(absPath)
	if err != nil {
		t.Fatalf("IndexSingleFile: %v", err)
	}
	if !reindexed {
		t.Error("expected file to be indexed (new)")
	}

	// Second call should skip (unchanged)
	reindexed, err = idx.IndexSingleFile(absPath)
	if err != nil {
		t.Fatalf("IndexSingleFile second call: %v", err)
	}
	if reindexed {
		t.Error("expected file to be skipped (unchanged)")
	}
}

func TestLangDetection(t *testing.T) {
	tests := []struct {
		path string
		want Lang
	}{
		{"foo.go", LangGo},
		{"Bar.java", LangJava},
		{"lib.rs", LangRust},
		{"script.py", LangPython},
		{"app.ts", LangTypeScript},
		{"app.tsx", LangTypeScript},
		{"main.js", LangJavaScript},
		{"main.jsx", LangJavaScript},
		{"readme.md", LangUnknown},
		{"image.png", LangUnknown},
	}
	for _, tt := range tests {
		got := DetectLang(tt.path)
		if got != tt.want {
			t.Errorf("DetectLang(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// --- helpers ---

func expectSymbol(t *testing.T, result *symbols.FileResult, name string, kind symbols.SymbolKind) {
	t.Helper()
	for _, s := range result.Symbols {
		if s.Name == name && s.Kind == kind {
			return
		}
	}
	// Show what we have for debugging
	var found []string
	for _, s := range result.Symbols {
		if s.Name == name {
			found = append(found, s.Kind.String())
		}
	}
	if len(found) > 0 {
		t.Errorf("symbol %q found with kinds %v, want %s", name, found, kind)
	} else {
		t.Errorf("symbol %q (%s) not found in %d symbols", name, kind, len(result.Symbols))
	}
}

func expectRef(t *testing.T, result *symbols.FileResult, name string) {
	t.Helper()
	for _, r := range result.Refs {
		if r.Name == name {
			return
		}
	}
	t.Errorf("ref %q not found in %d refs", name, len(result.Refs))
}
