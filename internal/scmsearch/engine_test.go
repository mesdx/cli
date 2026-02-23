package scmsearch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testdataDir returns the absolute path to the indexer testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "..", "indexer", "testdata")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("testdata dir not found at %s", dir)
	}
	return dir
}

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	repoRoot := testdataDir(t)
	return NewEngine(repoRoot, []string{"."}, NewQueryCache(32))
}

// ---------------------------------------------------------------------------
// Stub tests
// ---------------------------------------------------------------------------

func TestStubDefsFunctionNamed_Go(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "SayHello"},
		IncludeGlobs: []string{"*.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.MatchesReturned == 0 {
		t.Fatal("expected at least one match for SayHello in Go testdata")
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "SayHello" && strings.Contains(m.FilePath, "sample.go") {
			found = true
			if m.CaptureName != "def.function" {
				t.Errorf("expected capture name def.function, got %s", m.CaptureName)
			}
			if m.StartLine != 33 {
				t.Errorf("expected SayHello at line 33, got %d", m.StartLine)
			}
			break
		}
	}
	if !found {
		t.Errorf("did not find SayHello in go/sample.go; matches: %+v", res.Matches)
	}
}

func TestStubDefsFunctionNamed_Python(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "python",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "say_hello"},
		IncludeGlobs: []string{"sample.py"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "say_hello" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find say_hello in python/sample.py")
	}
}

func TestStubDefsFunctionNamed_Rust(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "rust",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "main"},
		IncludeGlobs: []string{"sample.rs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "main" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find main in rust/sample.rs")
	}
}

func TestStubDefsFunctionNamed_TypeScript(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "typescript",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "sayHello"},
		IncludeGlobs: []string{"sample.ts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "sayHello" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find sayHello in ts/sample.ts")
	}
}

func TestStubDefsFunctionNamed_JavaScript(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "javascript",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "sayHello"},
		IncludeGlobs: []string{"sample.js"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "sayHello" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find sayHello in js/sample.js")
	}
}

func TestStubDefsFunctionNamed_Java(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language: "java",
		StubName: "defs.function.named",
		StubArgs: map[string]string{"name": "getName"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "getName" && strings.Contains(m.FilePath, "Sample.java") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find getName in java/Sample.java")
	}
}

func TestStubDefsClassNamed_Go(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "defs.class.named",
		StubArgs:     map[string]string{"name": "Person"},
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Person" {
			found = true
			if m.CaptureName != "def.class" {
				t.Errorf("expected capture name def.class, got %s", m.CaptureName)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find struct Person in go/sample.go")
	}
}

func TestStubDefsClassNamed_Python(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "python",
		StubName:     "defs.class.named",
		StubArgs:     map[string]string{"name": "Person"},
		IncludeGlobs: []string{"sample.py"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Person" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find class Person in python/sample.py")
	}
}

func TestStubDefsClassNamed_TypeScript(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "typescript",
		StubName:     "defs.class.named",
		StubArgs:     map[string]string{"name": "Person"},
		IncludeGlobs: []string{"sample.ts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Person" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find class Person in ts/sample.ts")
	}
}

func TestStubDefsInterfaceNamed_Go(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "defs.interface.named",
		StubArgs:     map[string]string{"name": "Greeter"},
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Greeter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find interface Greeter in go/sample.go")
	}
}

func TestStubDefsInterfaceNamed_Rust(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "rust",
		StubName:     "defs.interface.named",
		StubArgs:     map[string]string{"name": "Greeter"},
		IncludeGlobs: []string{"sample.rs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Greeter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find trait Greeter in rust/sample.rs")
	}
}

func TestStubRefsTypeNamed_Go(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "refs.type.named",
		StubArgs:     map[string]string{"name": "Person"},
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.MatchesReturned == 0 {
		t.Fatal("expected at least one type ref to Person in go/sample.go")
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Person" && m.CaptureName == "ref.type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one match with text 'Person'")
	}
}

func TestStubRefsCallNamed_Go(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "refs.call.named",
		StubArgs:     map[string]string{"name": "NewPerson"},
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "NewPerson" && m.CaptureName == "ref.call" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find call to NewPerson in go/sample.go")
	}
}

func TestStubDefsMethodNamed_TypeScript(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "typescript",
		StubName:     "defs.method.named",
		StubArgs:     map[string]string{"name": "greet"},
		IncludeGlobs: []string{"sample.ts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "greet" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find method greet in ts/sample.ts")
	}
}

// ---------------------------------------------------------------------------
// Raw query tests
// ---------------------------------------------------------------------------

func TestRawQuery_GoFunctions(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		Query:        `(function_declaration name: (identifier) @fn)`,
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, m := range res.Matches {
		names[m.TextSnippet] = true
	}
	for _, expected := range []string{"NewPerson", "SayHello"} {
		if !names[expected] {
			t.Errorf("expected raw query to find function %s, got: %v", expected, names)
		}
	}
}

func TestRawQuery_PythonClasses(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "python",
		Query:        `(class_definition name: (identifier) @cls)`,
		IncludeGlobs: []string{"sample.py"},
	})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, m := range res.Matches {
		names[m.TextSnippet] = true
	}
	for _, expected := range []string{"Greeter", "Person"} {
		if !names[expected] {
			t.Errorf("expected raw query to find class %s, got: %v", expected, names)
		}
	}
}

func TestRawQuery_TypeScriptInterfaces(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "typescript",
		Query:        `(interface_declaration name: (type_identifier) @iface)`,
		IncludeGlobs: []string{"sample.ts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Greeter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected raw query to find interface Greeter in ts/sample.ts")
	}
}

func TestRawQuery_JavaEnums(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "java",
		Query:        `(enum_declaration name: (identifier) @enum)`,
		IncludeGlobs: []string{"Sample.java"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected raw query to find enum Status in java/Sample.java")
	}
}

func TestRawQuery_RustStructs(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "rust",
		Query:        `(struct_item name: (type_identifier) @st)`,
		IncludeGlobs: []string{"sample.rs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range res.Matches {
		if m.TextSnippet == "Person" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected raw query to find struct Person in rust/sample.rs")
	}
}

func TestRawQuery_WithPredicate(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		Query:        `(function_declaration name: (identifier) @fn (#match? @fn "^New"))`,
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range res.Matches {
		if !strings.HasPrefix(m.TextSnippet, "New") {
			t.Errorf("expected match starting with New, got %q", m.TextSnippet)
		}
	}
	if res.Summary.MatchesReturned == 0 {
		t.Error("expected at least one function matching ^New")
	}
}

// ---------------------------------------------------------------------------
// Glob filter tests
// ---------------------------------------------------------------------------

func TestIncludeGlobs_NarrowsToSingleFile(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		Query:        `(function_declaration name: (identifier) @fn)`,
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range res.Matches {
		if !strings.HasSuffix(m.FilePath, "sample.go") {
			t.Errorf("expected all matches in sample.go, got %s", m.FilePath)
		}
	}
}

func TestExcludeGlobs_SkipsFiles(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		Query:        `(function_declaration name: (identifier) @fn)`,
		ExcludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range res.Matches {
		if strings.HasSuffix(m.FilePath, "sample.go") {
			t.Errorf("sample.go should have been excluded, found match in %s", m.FilePath)
		}
	}
}

// ---------------------------------------------------------------------------
// Context and AST metadata tests
// ---------------------------------------------------------------------------

func TestContextLines(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "SayHello"},
		IncludeGlobs: []string{"sample.go"},
		ContextLines: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Matches) == 0 {
		t.Fatal("expected matches")
	}
	m := res.Matches[0]
	if len(m.ContextBefore) == 0 {
		t.Error("expected ContextBefore lines")
	}
	if len(m.ContextAfter) == 0 {
		t.Error("expected ContextAfter lines")
	}
	if m.Line == "" {
		t.Error("expected Line to be populated")
	}
}

func TestASTParents(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:       "go",
		StubName:       "defs.function.named",
		StubArgs:       map[string]string{"name": "NewPerson"},
		IncludeGlobs:   []string{"sample.go"},
		ASTParentDepth: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Matches) == 0 {
		t.Fatal("expected matches")
	}
	m := res.Matches[0]
	if len(m.ASTParents) == 0 {
		t.Error("expected AST parents to be populated")
	}
	if m.ASTParents[0] != "function_declaration" {
		t.Errorf("expected first parent to be function_declaration, got %s", m.ASTParents[0])
	}
}

func TestNodeType(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "SayHello"},
		IncludeGlobs: []string{"sample.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Matches) == 0 {
		t.Fatal("expected matches")
	}
	if res.Matches[0].NodeType != "identifier" {
		t.Errorf("expected node type identifier, got %s", res.Matches[0].NodeType)
	}
}

// ---------------------------------------------------------------------------
// Safety cap tests
// ---------------------------------------------------------------------------

func TestMaxMatches(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:   "go",
		Query:      `(identifier) @id`,
		MaxMatches: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.MatchesReturned > 5 {
		t.Errorf("expected at most 5 matches, got %d", res.Summary.MatchesReturned)
	}
}

func TestMaxMatchesPerFile(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language:          "go",
		Query:             `(identifier) @id`,
		IncludeGlobs:      []string{"sample.go"},
		MaxMatchesPerFile: 3,
		MaxMatches:        1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.MatchesReturned > 3 {
		t.Errorf("expected at most 3 matches per file, got %d", res.Summary.MatchesReturned)
	}
}

// ---------------------------------------------------------------------------
// Cache tests
// ---------------------------------------------------------------------------

func TestCacheHits(t *testing.T) {
	cache := NewQueryCache(32)
	e := NewEngine(testdataDir(t), []string{"."}, cache)

	req := SearchRequest{
		Language:     "go",
		StubName:     "defs.function.named",
		StubArgs:     map[string]string{"name": "SayHello"},
		IncludeGlobs: []string{"sample.go"},
	}

	_, err := e.Search(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	h1, m1 := cache.Stats()

	_, err = e.Search(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	h2, m2 := cache.Stats()

	if h2 <= h1 {
		t.Errorf("expected cache hits to increase on second run; first=%d, second=%d", h1, h2)
	}
	if m2 != m1 {
		t.Errorf("expected no new cache misses on second run; first=%d, second=%d", m1, m2)
	}
}

// ---------------------------------------------------------------------------
// Deterministic ordering test
// ---------------------------------------------------------------------------

func TestDeterministicOrdering(t *testing.T) {
	e := newTestEngine(t)
	req := SearchRequest{
		Language: "go",
		Query:    `(function_declaration name: (identifier) @fn)`,
	}

	res1, err := e.Search(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	res2, err := e.Search(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if len(res1.Matches) != len(res2.Matches) {
		t.Fatalf("different number of matches: %d vs %d", len(res1.Matches), len(res2.Matches))
	}

	for i := range res1.Matches {
		a, b := res1.Matches[i], res2.Matches[i]
		if a.FilePath != b.FilePath || a.StartLine != b.StartLine || a.StartCol != b.StartCol {
			t.Errorf("match %d differs: %s:%d:%d vs %s:%d:%d",
				i, a.FilePath, a.StartLine, a.StartCol, b.FilePath, b.StartLine, b.StartCol)
		}
	}
}

// ---------------------------------------------------------------------------
// Error handling tests
// ---------------------------------------------------------------------------

func TestErrorOnMissingQueryAndStub(t *testing.T) {
	e := newTestEngine(t)
	_, err := e.Search(context.Background(), SearchRequest{Language: "go"})
	if err == nil {
		t.Error("expected error when neither query nor stubName provided")
	}
}

func TestErrorOnUnknownStub(t *testing.T) {
	e := newTestEngine(t)
	_, err := e.Search(context.Background(), SearchRequest{
		Language: "go",
		StubName: "nonexistent.stub",
	})
	if err == nil {
		t.Error("expected error for unknown stub")
	}
}

func TestErrorOnInvalidQuery(t *testing.T) {
	e := newTestEngine(t)
	// Use a completely nonexistent node type which tree-sitter rejects
	res, err := e.Search(context.Background(), SearchRequest{
		Language:     "go",
		Query:        `(nonexistent_node_type_xyz_123 name: (identifier) @x)`,
		IncludeGlobs: []string{"sample.go"},
	})
	// Tree-sitter may fail to compile or simply return 0 matches; both are acceptable
	if err != nil {
		return // compile error is one valid outcome
	}
	if res.Summary.MatchesReturned != 0 {
		t.Errorf("expected 0 matches for bogus node type, got %d", res.Summary.MatchesReturned)
	}
}

func TestErrorOnMissingStubArg(t *testing.T) {
	e := newTestEngine(t)
	_, err := e.Search(context.Background(), SearchRequest{
		Language: "go",
		StubName: "defs.function.named",
		StubArgs: map[string]string{},
	})
	if err == nil {
		t.Error("expected error when required stub arg 'name' is missing")
	}
}

// ---------------------------------------------------------------------------
// Multi-file / cross-file tests
// ---------------------------------------------------------------------------

func TestCrossFileSearch(t *testing.T) {
	e := newTestEngine(t)
	res, err := e.Search(context.Background(), SearchRequest{
		Language: "go",
		StubName: "defs.function.named",
		StubArgs: map[string]string{"name": "SayHello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Summary.FilesScanned < 2 {
		t.Errorf("expected to scan multiple Go files, scanned %d", res.Summary.FilesScanned)
	}
}

// ---------------------------------------------------------------------------
// Format test
// ---------------------------------------------------------------------------

func TestFormatResults(t *testing.T) {
	result := &SearchResult{
		Matches: []Match{
			{
				FilePath:    "go/sample.go",
				StartLine:   33,
				StartCol:    5,
				EndLine:     33,
				EndCol:      13,
				CaptureName: "def.function",
				NodeType:    "identifier",
				TextSnippet: "SayHello",
				Line:        "func SayHello() {",
				ASTParents:  []string{"function_declaration", "source_file"},
			},
		},
		Summary: Summary{
			FilesScanned:    10,
			FilesMatched:    1,
			MatchesReturned: 1,
			DurationMs:      42,
			CacheHits:       5,
			CacheMisses:     1,
		},
	}

	text := FormatResults(result)
	checks := []string{
		"1 matches",
		"go/sample.go",
		"@def.function",
		"SayHello",
		"function_declaration",
	}
	for _, c := range checks {
		if !strings.Contains(text, c) {
			t.Errorf("formatted output missing %q", c)
		}
	}
}

func TestFormatResults_Empty(t *testing.T) {
	text := FormatResults(&SearchResult{})
	if !strings.Contains(text, "No matches found") {
		t.Error("expected 'No matches found' for empty results")
	}
}
