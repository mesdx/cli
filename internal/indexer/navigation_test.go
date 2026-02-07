package indexer

import (
	"path/filepath"
	"testing"

	"github.com/codeintelx/cli/internal/db"
)

func setupNavigationTest(t *testing.T) (*Navigator, *Indexer, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := db.Initialize(dbPath); err != nil {
		t.Fatalf("db.Initialize: %v", err)
	}
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}

	repoRoot := testdataDir(t)
	idx := New(d, repoRoot)

	// Full index the testdata
	if _, err := idx.FullIndex([]string{"."}); err != nil {
		t.Fatalf("FullIndex: %v", err)
	}

	nav := &Navigator{DB: d, ProjectID: idx.Store.ProjectID}
	return nav, idx, func() { d.Close() }
}

func TestGoToDefinitionByName(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	tests := []struct {
		name      string
		lang      string
		wantFound bool
	}{
		{"Person", "go", true},
		{"NewPerson", "go", true},
		{"Greeter", "go", true},
		{"SayHello", "go", true},
		{"NonExistent", "go", false},
	}

	for _, tt := range tests {
		results, err := nav.GoToDefinitionByName(tt.name, "", tt.lang)
		if err != nil {
			t.Fatalf("GoToDefinitionByName(%q, %q): %v", tt.name, tt.lang, err)
		}
		if tt.wantFound && len(results) == 0 {
			t.Errorf("GoToDefinitionByName(%q, %q) returned 0 results, want >0", tt.name, tt.lang)
		}
		if !tt.wantFound && len(results) > 0 {
			t.Errorf("GoToDefinitionByName(%q, %q) returned %d results, want 0", tt.name, tt.lang, len(results))
		}
	}
}

func TestGoToDefinitionByNameReturnsCorrectInfo(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// Test finding Person across different languages
	langs := []string{"go", "python", "rust"}
	for _, lang := range langs {
		results, err := nav.GoToDefinitionByName("Person", "", lang)
		if err != nil {
			t.Fatalf("GoToDefinitionByName(Person, %s): %v", lang, err)
		}

		if len(results) == 0 {
			t.Errorf("expected at least one result for Person in %s", lang)
			continue
		}

		for _, r := range results {
			if r.Name != "Person" {
				t.Errorf("unexpected name %q", r.Name)
			}
			if r.Location.StartLine <= 0 {
				t.Errorf("invalid start line: %d", r.Location.StartLine)
			}
		}
	}
}

func TestGoToDefinitionByPosition(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// In the Go fixture, NewPerson is on line 27 (0-indexed col ~5)
	// The symbol "NewPerson" starts at some position. Let's look up by name first to know the position.
	defs, err := nav.GoToDefinitionByName("NewPerson", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("cannot find NewPerson definition to test cursor-based lookup")
	}

	// Use the found position for a cursor-based lookup
	def := defs[0]
	results, err := nav.GoToDefinitionByPosition(def.Location.Path, def.Location.StartLine, def.Location.StartCol, "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByPosition: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result from cursor-based lookup")
	}
}

func TestFindUsagesByName(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// "Person" should have usages across files
	results, err := nav.FindUsagesByName("Person", "", "go")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one usage of Person")
	}
}

func TestFindUsagesByNameGoFixture(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// "NewPerson" should have usage references in the Go fixture
	results, err := nav.FindUsagesByName("NewPerson", "", "go")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one usage of NewPerson")
	}
}

func TestFindUsagesByPosition(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// Find a known ref position
	usages, err := nav.FindUsagesByName("NewPerson", "", "go")
	if err != nil || len(usages) == 0 {
		t.Skip("cannot find NewPerson usage to test cursor-based lookup")
	}

	ref := usages[0]
	results, err := nav.FindUsagesByPosition(ref.Location.Path, ref.Location.StartLine, ref.Location.StartCol, "go")
	if err != nil {
		t.Fatalf("FindUsagesByPosition: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result from cursor-based usage lookup")
	}
}

func TestFindUsagesNonExistent(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	results, err := nav.FindUsagesByName("CompletelyNonExistentSymbol12345", "", "go")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent symbol, got %d", len(results))
	}
}

func TestFormatDefinitions(t *testing.T) {
	results := []DefinitionResult{
		{
			Name:      "Foo",
			Kind:      "function",
			Signature: "func Foo(x int) error",
			Location:  Location{Path: "pkg/foo.go", StartLine: 10, StartCol: 5},
		},
	}
	text := FormatDefinitions(results)
	if text == "" || text == "No definitions found." {
		t.Error("expected non-empty formatted text")
	}
}

func TestFormatUsages(t *testing.T) {
	results := []UsageResult{
		{
			Name:             "Foo",
			Kind:             "call",
			ContextContainer: "main",
			Location:         Location{Path: "pkg/bar.go", StartLine: 20, StartCol: 3},
		},
	}
	text := FormatUsages(results)
	if text == "" || text == "No usages found." {
		t.Error("expected non-empty formatted text")
	}
}

func TestFormatEmptyResults(t *testing.T) {
	if text := FormatDefinitions(nil); text != "No definitions found." {
		t.Errorf("unexpected: %q", text)
	}
	if text := FormatUsages(nil); text != "No usages found." {
		t.Errorf("unexpected: %q", text)
	}
}
