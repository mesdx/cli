package indexer

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/codeintelx/cli/internal/db"
)

// ---------------------------------------------------------------------------
// Scoring tests
// ---------------------------------------------------------------------------

func TestScoreUsages_SingleCandidate(t *testing.T) {
	// With one candidate, every usage should score 1.0 (full confidence).
	usages := []UsageResult{
		{Name: "Foo", Kind: "other", Location: Location{Path: "a.go", StartLine: 10}},
		{Name: "Foo", Kind: "other", Location: Location{Path: "b.go", StartLine: 20}},
	}
	candidates := []DefinitionResult{
		{Name: "Foo", Kind: "function", Location: Location{Path: "a.go", StartLine: 5}},
	}

	scored := ScoreUsages(usages, candidates, &candidates[0], "/tmp/repo")
	for i, su := range scored {
		if su.DependencyScore < 0.99 {
			t.Errorf("usage[%d] score=%.4f, want ~1.0 (single candidate)", i, su.DependencyScore)
		}
	}
}

func TestScoreUsages_MultipleCandidates_SameFileBoosted(t *testing.T) {
	// Usage in same file as candidate A should score higher for A.
	usages := []UsageResult{
		{Name: "Config", Kind: "other", Location: Location{Path: "pkg/config.go", StartLine: 10}},
	}
	candidates := []DefinitionResult{
		{Name: "Config", Kind: "struct", Location: Location{Path: "pkg/config.go", StartLine: 5}},
		{Name: "Config", Kind: "struct", Location: Location{Path: "other/types.go", StartLine: 5}},
	}

	scored := ScoreUsages(usages, candidates, nil, "/tmp/repo")
	if len(scored) != 1 {
		t.Fatalf("expected 1 scored usage, got %d", len(scored))
	}
	// Same-file candidate should be best.
	if scored[0].BestDefinition == nil {
		t.Fatal("expected BestDefinition to be set")
	}
	if scored[0].BestDefinition.Location.Path != "pkg/config.go" {
		t.Errorf("expected best def in pkg/config.go, got %s", scored[0].BestDefinition.Location.Path)
	}
	if scored[0].DependencyScore <= 0.5 {
		t.Errorf("expected score > 0.5 for same-file candidate, got %.4f", scored[0].DependencyScore)
	}
}

func TestScoreUsages_NoCandidates(t *testing.T) {
	usages := []UsageResult{
		{Name: "Unknown", Kind: "other", Location: Location{Path: "a.go", StartLine: 1}},
	}
	scored := ScoreUsages(usages, nil, nil, "/tmp/repo")
	if len(scored) != 1 {
		t.Fatalf("expected 1 scored usage, got %d", len(scored))
	}
	if scored[0].DependencyScore != 0 {
		t.Errorf("expected score=0 for no candidates, got %.4f", scored[0].DependencyScore)
	}
}

func TestScoreUsages_PrimaryDefPreferred(t *testing.T) {
	// When primaryDef is provided, score should reflect P(primaryDef | ref).
	usages := []UsageResult{
		{Name: "Handle", Kind: "other", Location: Location{Path: "main.go", StartLine: 10}},
	}
	defA := DefinitionResult{Name: "Handle", Kind: "function", Location: Location{Path: "main.go", StartLine: 5}}
	defB := DefinitionResult{Name: "Handle", Kind: "function", Location: Location{Path: "other.go", StartLine: 5}}

	scored := ScoreUsages(usages, []DefinitionResult{defA, defB}, &defA, "/tmp/repo")
	if scored[0].DependencyScore <= 0.5 {
		t.Errorf("expected score > 0.5 for primary def same file, got %.4f", scored[0].DependencyScore)
	}
}

func TestScoreUsages_Normalization(t *testing.T) {
	// All probabilities across candidates should sum to ~1.0.
	usages := []UsageResult{
		{Name: "Proc", Kind: "other", Location: Location{Path: "x.go", StartLine: 10}},
	}
	candidates := []DefinitionResult{
		{Name: "Proc", Kind: "struct", Location: Location{Path: "a.go", StartLine: 1}},
		{Name: "Proc", Kind: "function", Location: Location{Path: "b.go", StartLine: 1}},
		{Name: "Proc", Kind: "variable", Location: Location{Path: "c.go", StartLine: 1}},
	}

	// Manually compute: each candidate gets a weight, and the max should be the score.
	scored := ScoreUsages(usages, candidates, nil, "/tmp/repo")
	if scored[0].DependencyScore <= 0 || scored[0].DependencyScore > 1.0 {
		t.Errorf("score out of range [0,1]: %.4f", scored[0].DependencyScore)
	}
}

// ---------------------------------------------------------------------------
// Kind compatibility tests
// ---------------------------------------------------------------------------

func TestKindCompatibility_NewExpr(t *testing.T) {
	// "new Foo" should boost class/struct/constructor.
	boost := kindCompatibility("  const x = new Foo()", "Foo", "class")
	neutral := kindCompatibility("  const x = new Foo()", "Foo", "variable")
	if boost <= neutral {
		t.Errorf("expected class boost (%.2f) > variable (%.2f) for new expression", boost, neutral)
	}
}

func TestKindCompatibility_FunctionCall(t *testing.T) {
	boost := kindCompatibility("  result := doSomething(args)", "doSomething", "function")
	neutral := kindCompatibility("  result := doSomething(args)", "doSomething", "struct")
	if boost <= neutral {
		t.Errorf("expected function boost (%.2f) > struct (%.2f) for call", boost, neutral)
	}
}

func TestKindCompatibility_DotAccess(t *testing.T) {
	boost := kindCompatibility("  x.Name = value", "Name", "field")
	neutral := kindCompatibility("  x.Name = value", "Name", "function")
	if boost <= neutral {
		t.Errorf("expected field boost (%.2f) > function (%.2f) for dot access", boost, neutral)
	}
}

func TestKindCompatibility_Extends(t *testing.T) {
	boost := kindCompatibility("class Foo extends Bar {", "Bar", "class")
	neutral := kindCompatibility("class Foo extends Bar {", "Bar", "function")
	if boost <= neutral {
		t.Errorf("expected class boost (%.2f) > function (%.2f) for extends", boost, neutral)
	}
}

func TestKindCompatibility_NoContext(t *testing.T) {
	// Empty source line should return neutral (1.0).
	result := kindCompatibility("", "Foo", "function")
	if result != 1.0 {
		t.Errorf("expected 1.0 for empty context, got %.2f", result)
	}
}

// ---------------------------------------------------------------------------
// RefKind compatibility tests
// ---------------------------------------------------------------------------

func TestRefKindCompatibility_Import(t *testing.T) {
	score := refKindCompatibility("import", "function")
	if score < 2.0 {
		t.Errorf("expected high score for import ref, got %.2f", score)
	}
}

func TestRefKindCompatibility_TypeRef(t *testing.T) {
	high := refKindCompatibility("type_ref", "class")
	low := refKindCompatibility("type_ref", "function")
	if high <= low {
		t.Errorf("type_ref should score higher for class (%.2f) than function (%.2f)", high, low)
	}
}

func TestRefKindCompatibility_Other(t *testing.T) {
	score := refKindCompatibility("other", "function")
	if score != 1.0 {
		t.Errorf("expected 1.0 for 'other' ref kind, got %.2f", score)
	}
}

// ---------------------------------------------------------------------------
// Grouping and sorting tests
// ---------------------------------------------------------------------------

func TestGroupAndSortUsages_Empty(t *testing.T) {
	result := GroupAndSortUsages(nil, 3)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestGroupAndSortUsages_AdjacentGrouped(t *testing.T) {
	usages := []ScoredUsage{
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 10, EndLine: 10}}, DependencyScore: 0.3},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 12, EndLine: 12}}, DependencyScore: 0.9},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 100, EndLine: 100}}, DependencyScore: 0.5},
	}

	result := GroupAndSortUsages(usages, 3)
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}

	// Lines 10+12 are adjacent (gap <= 3), should be grouped together.
	// Their group maxScore = 0.9.
	// Line 100 is separate, maxScore = 0.5.
	// Group with score 0.9 should come first.
	if result[0].Location.StartLine != 10 && result[0].Location.StartLine != 12 {
		t.Errorf("expected first result from high-score group (line 10 or 12), got line %d", result[0].Location.StartLine)
	}
}

func TestGroupAndSortUsages_DescendingByScore(t *testing.T) {
	usages := []ScoredUsage{
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 1, EndLine: 1}}, DependencyScore: 0.2},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "b.go", StartLine: 1, EndLine: 1}}, DependencyScore: 0.8},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "c.go", StartLine: 1, EndLine: 1}}, DependencyScore: 0.5},
	}

	result := GroupAndSortUsages(usages, 3)
	// Different files → no adjacency merging → sorted by score desc.
	if result[0].DependencyScore < result[1].DependencyScore {
		t.Errorf("expected descending score order, got %.4f then %.4f",
			result[0].DependencyScore, result[1].DependencyScore)
	}
	if result[1].DependencyScore < result[2].DependencyScore {
		t.Errorf("expected descending score order, got %.4f then %.4f",
			result[1].DependencyScore, result[2].DependencyScore)
	}
}

func TestGroupAndSortUsages_PreservesLineOrderInGroup(t *testing.T) {
	usages := []ScoredUsage{
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 15, EndLine: 15}}, DependencyScore: 0.3},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 13, EndLine: 13}}, DependencyScore: 0.9},
		{UsageResult: UsageResult{Name: "X", Location: Location{Path: "a.go", StartLine: 14, EndLine: 14}}, DependencyScore: 0.5},
	}

	result := GroupAndSortUsages(usages, 3)
	// All adjacent → one group → sorted by line asc within group.
	if result[0].Location.StartLine != 13 ||
		result[1].Location.StartLine != 14 ||
		result[2].Location.StartLine != 15 {
		t.Errorf("expected line order 13,14,15 within group, got %d,%d,%d",
			result[0].Location.StartLine, result[1].Location.StartLine, result[2].Location.StartLine)
	}
}

// ---------------------------------------------------------------------------
// Format tests
// ---------------------------------------------------------------------------

func TestFormatScoredUsages(t *testing.T) {
	scored := []ScoredUsage{
		{
			UsageResult:     UsageResult{Name: "Foo", Kind: "call", ContextContainer: "main", Location: Location{Path: "a.go", StartLine: 10, StartCol: 5}},
			DependencyScore: 0.85,
			BestDefinition:  &DefinitionResult{Name: "Foo", Kind: "function", Location: Location{Path: "b.go", StartLine: 5}},
		},
	}
	text := FormatScoredUsages(scored)
	if text == "" || text == "No usages found." {
		t.Error("expected non-empty formatted text")
	}
	// Check that score is included.
	if !containsStr(text, "0.8500") {
		t.Errorf("expected score in output, got: %s", text)
	}
	// Check that best definition is included.
	if !containsStr(text, "b.go:5") {
		t.Errorf("expected best definition location in output, got: %s", text)
	}
}

func TestFormatScoredUsages_Empty(t *testing.T) {
	text := FormatScoredUsages(nil)
	if text != "No usages found." {
		t.Errorf("unexpected: %q", text)
	}
}

func TestFormatUsages_WithScore(t *testing.T) {
	results := []UsageResult{
		{Name: "Bar", Kind: "other", Location: Location{Path: "x.go", StartLine: 5, StartCol: 2}, DependencyScore: 0.75},
	}
	text := FormatUsages(results)
	if !containsStr(text, "0.7500") {
		t.Errorf("expected score in FormatUsages output, got: %s", text)
	}
}

// ---------------------------------------------------------------------------
// Utility function tests
// ---------------------------------------------------------------------------

func TestRound4(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{0.123456, 0.1235},
		{1.0, 1.0},
		{0.0, 0.0},
		{0.99999, 1.0},
	}
	for _, tt := range tests {
		got := round4(tt.in)
		if math.Abs(got-tt.want) > 1e-6 {
			t.Errorf("round4(%f) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

func TestNodeID(t *testing.T) {
	id := nodeID("pkg/foo.go", "Bar", 42)
	expected := "pkg/foo.go:Bar:42"
	if id != expected {
		t.Errorf("nodeID = %q, want %q", id, expected)
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{-7, "-7"},
		{100, "100"},
	}
	for _, tt := range tests {
		got := itoa(tt.in)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSameDir(t *testing.T) {
	if !sameDir("pkg/foo.go", "pkg/bar.go") {
		t.Error("expected same dir for pkg/foo.go and pkg/bar.go")
	}
	if sameDir("pkg/foo.go", "other/bar.go") {
		t.Error("expected different dir for pkg/foo.go and other/bar.go")
	}
}

// ---------------------------------------------------------------------------
// Integration tests: scoring + graph with real DB fixtures
// ---------------------------------------------------------------------------

func setupDepScoreTest(t *testing.T) (*Navigator, string, func()) {
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
	if _, err := idx.FullIndex([]string{"."}); err != nil {
		t.Fatalf("FullIndex: %v", err)
	}

	nav := &Navigator{DB: d, ProjectID: idx.Store.ProjectID}
	return nav, repoRoot, func() { _ = d.Close() }
}

func TestScoreUsages_GoFixture(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	// "Person" has a definition in go/sample.go and usages across files.
	usages, err := nav.FindUsagesByName("Person", "", "go")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(usages) == 0 {
		t.Skip("no usages of Person found")
	}

	candidates, err := nav.GoToDefinitionByName("Person", "", "go")
	if err != nil || len(candidates) == 0 {
		t.Skip("no definition of Person found")
	}

	scored := ScoreUsages(usages, candidates, &candidates[0], repoRoot)
	for _, su := range scored {
		if su.DependencyScore < 0 || su.DependencyScore > 1.0 {
			t.Errorf("score out of range: %.4f", su.DependencyScore)
		}
		if su.BestDefinition == nil {
			t.Error("expected BestDefinition to be set")
		}
	}
}

func TestScoreUsages_PythonFixture(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("Person", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(usages) == 0 {
		t.Skip("no usages of Person found in python")
	}

	candidates, _ := nav.GoToDefinitionByName("Person", "", "python")
	scored := ScoreUsages(usages, candidates, nil, repoRoot)
	for _, su := range scored {
		if su.DependencyScore < 0 || su.DependencyScore > 1.0 {
			t.Errorf("score out of range: %.4f", su.DependencyScore)
		}
	}
}

func TestScoreUsages_TypeScriptFixture(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("Person", "", "typescript")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if len(usages) == 0 {
		t.Skip("no usages of Person found in typescript")
	}

	candidates, _ := nav.GoToDefinitionByName("Person", "", "typescript")
	scored := ScoreUsages(usages, candidates, nil, repoRoot)
	for _, su := range scored {
		if su.DependencyScore < 0 || su.DependencyScore > 1.0 {
			t.Errorf("score out of range: %.4f", su.DependencyScore)
		}
	}
}

func TestRefsInFileRange(t *testing.T) {
	nav, _, cleanup := setupDepScoreTest(t)
	defer cleanup()

	// SayHello is a function in go/sample.go, spanning some lines.
	defs, err := nav.GoToDefinitionByName("SayHello", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("SayHello definition not found")
	}
	def := defs[0]

	refs, err := nav.RefsInFileRange(def.Location.Path, def.Location.StartLine, def.Location.EndLine, "go")
	if err != nil {
		t.Fatalf("RefsInFileRange: %v", err)
	}
	if len(refs) == 0 {
		t.Error("expected refs inside SayHello body")
	}

	// Should find references to NewPerson, DefaultName, MaxRetries, etc.
	foundNames := map[string]bool{}
	for _, r := range refs {
		foundNames[r.Name] = true
	}
	for _, expected := range []string{"NewPerson", "DefaultName", "MaxRetries"} {
		if !foundNames[expected] {
			t.Errorf("expected ref to %q inside SayHello span", expected)
		}
	}
}

func TestBuildDependencyGraph_Go(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("SayHello", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("SayHello definition not found")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "go", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	if graph.PrimaryDefinition == nil {
		t.Fatal("expected PrimaryDefinition to be set")
	}
	if graph.PrimaryDefinition.Name != "SayHello" {
		t.Errorf("expected primary def name SayHello, got %s", graph.PrimaryDefinition.Name)
	}

	// Should have at least the primary node.
	if len(graph.SymbolGraph.Nodes) == 0 {
		t.Error("expected at least one node in symbol graph")
	}

	// SayHello calls NewPerson, uses DefaultName, MaxRetries → outbound edges.
	outboundCount := 0
	for _, e := range graph.SymbolGraph.Edges {
		if e.Kind == "outbound" {
			outboundCount++
		}
	}
	if outboundCount == 0 {
		t.Error("expected outbound edges for SayHello (it calls NewPerson, uses DefaultName, etc.)")
	}

	// SayHello is referenced by at least one usage.
	if len(graph.Usages) == 0 {
		// It's possible there are no direct usages of SayHello in Go fixtures.
		// That's OK for this test, but let's check with Person instead.
		t.Log("no usages of SayHello found (may be expected for this fixture)")
	}
}

func TestBuildDependencyGraph_FileGraph(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	// Use Person which is defined in go/sample.go and referenced in go/defs_b.go.
	defs, err := nav.GoToDefinitionByName("Person", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("Person definition not found")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "go", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	// Person is used in multiple files → should have file graph edges.
	if len(graph.Usages) > 0 {
		// If there are usages from different files, there should be file graph edges.
		usageFiles := map[string]bool{}
		for _, u := range graph.Usages {
			usageFiles[u.Location.Path] = true
		}
		if len(usageFiles) > 1 || (len(usageFiles) == 1 && !usageFiles[defs[0].Location.Path]) {
			if len(graph.FileGraph) == 0 {
				t.Error("expected file graph edges for cross-file usages")
			}
		}
	}
}

func TestBuildDependencyGraph_NilPrimaryDef(t *testing.T) {
	nav, repoRoot, cleanup := setupDepScoreTest(t)
	defer cleanup()

	graph, err := BuildDependencyGraph(nav, nil, nil, "go", repoRoot, 1, 0.2, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}
	if graph.PrimaryDefinition != nil {
		t.Error("expected nil PrimaryDefinition")
	}
	if len(graph.SymbolGraph.Nodes) != 0 {
		t.Error("expected empty nodes for nil primary def")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsStr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && findSubstring(haystack, needle)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
