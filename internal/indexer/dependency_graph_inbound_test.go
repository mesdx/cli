package indexer

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesdx/cli/internal/db"
)

// setupInboundTest creates an indexed Navigator for the inbound dependency graph tests.
func setupInboundTest(t *testing.T) (*Navigator, string, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "inbound.db")
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
	nav := &Navigator{DB: d, ProjectID: idx.Store.ProjectID, RepoRoot: repoRoot}
	return nav, repoRoot, func() { _ = d.Close() }
}

// nodeIDsInGraph returns a set of all node IDs across inbound nodes, outbound nodes, and the primary node.
func nodeIDsInGraph(graph *DependencyGraph) map[string]bool {
	ids := make(map[string]bool)
	if graph.PrimaryNode != nil {
		ids[graph.PrimaryNode.ID] = true
	}
	for _, n := range graph.Inbound.Nodes {
		ids[n.ID] = true
	}
	for _, n := range graph.Outbound.Nodes {
		ids[n.ID] = true
	}
	return ids
}

// inboundEdgesFromGraph returns the inbound edges from the Inbound section.
func inboundEdgesFromGraph(graph *DependencyGraph) []DepGraphEdge {
	return graph.Inbound.Edges
}

// outboundEdgesFromGraph returns the outbound edges from the Outbound section.
func outboundEdgesFromGraph(graph *DependencyGraph) []DepGraphEdge {
	return graph.Outbound.Edges
}

// -----------------------------------------------------------------------
// Core inbound-node regression test (applies to all languages)
// -----------------------------------------------------------------------

// TestInboundNodesExistForEveryInboundEdge asserts the fundamental graph
// completeness invariant: every inbound edge's From file MUST have a
// corresponding node in symbolGraph.nodes.
func TestInboundNodesExistForEveryInboundEdge(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	langCases := []struct {
		lang       string
		symbolName string
	}{
		{"go", "CriticalUserModel"},
		{"java", "CriticalUserModel"},
		{"rust", "CriticalUserModel"},
		{"typescript", "CriticalUserModel"},
		{"javascript", "CriticalUserModel"},
		{"python", "CriticalUserModel"},
	}

	for _, tc := range langCases {
		t.Run(tc.lang, func(t *testing.T) {
			defs, err := nav.GoToDefinitionByName(tc.symbolName, "", tc.lang)
			if err != nil || len(defs) == 0 {
				t.Skipf("definition of %q not found in %s", tc.symbolName, tc.lang)
			}

			graph, err := BuildDependencyGraph(nav, &defs[0], defs, tc.lang, repoRoot, 1, 0.0, 500)
			if err != nil {
				t.Fatalf("BuildDependencyGraph: %v", err)
			}

			nodeIDs := nodeIDsInGraph(graph)
			inboundEdges := inboundEdgesFromGraph(graph)

			// Core regression: every inbound edge's From must have a node in Inbound.Nodes.
			for _, e := range inboundEdges {
				if !nodeIDs[e.From] {
					t.Errorf("inbound edge From=%q has no node in inbound.nodes (lang=%s)", e.From, tc.lang)
				}
				if e.FilePath == "" {
					t.Errorf("inbound edge has empty FilePath (lang=%s)", tc.lang)
				}
				if e.Score < 0 || e.Score > 1.0 {
					t.Errorf("inbound edge score %.4f out of [0,1] range (lang=%s)", e.Score, tc.lang)
				}
			}

			// Inbound file nodes must have non-empty Path.
			for _, n := range graph.Inbound.Nodes {
				if n.Path == "" {
					t.Errorf("inbound file node %q has empty Path (lang=%s)", n.ID, tc.lang)
				}
			}
		})
	}
}

// -----------------------------------------------------------------------
// Go: CriticalUserModel inbound/outbound detailed assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_Go_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in go")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "go", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	// Must have inbound edges from at least 3 files (views, services, defs_c).
	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for Go CriticalUserModel, got %d", len(inboundEdges))
	}

	// Every inbound edge must have a matching node in Inbound.Nodes.
	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes", e.From)
		}
	}

	// Each expected file must appear as an inbound edge source.
	expectedSuffixes := []string{
		filepath.Join("views", "critical_user_model_view.go"),
		filepath.Join("services", "critical_user_model_service.go"),
		"defs_c.go",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found", suffix)
		}
	}

	// Outbound: CriticalUserModel embeds/uses BaseModel — expect at least 1 outbound edge.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for Go CriticalUserModel (should reference BaseModel)")
	}

	// Summary fields must be consistent with edge counts.
	assertSummaryConsistency(t, graph, "go")
}

// -----------------------------------------------------------------------
// Java: CriticalUserModel inbound/outbound assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_Java_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "java")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in java")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "java", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for Java CriticalUserModel, got %d", len(inboundEdges))
	}

	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes (java)", e.From)
		}
	}

	expectedSuffixes := []string{
		filepath.Join("views", "CriticalUserModelView.java"),
		filepath.Join("services", "CriticalUserModelService.java"),
		"DefsC.java",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found (java)", suffix)
		}
	}

	// CriticalUserModel extends BaseModel → outbound edge expected.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for Java CriticalUserModel (should reference BaseModel)")
	}

	assertSummaryConsistency(t, graph, "java")
}

// -----------------------------------------------------------------------
// Rust: CriticalUserModel inbound/outbound assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_Rust_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "rust")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in rust")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "rust", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for Rust CriticalUserModel, got %d", len(inboundEdges))
	}

	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes (rust)", e.From)
		}
	}

	expectedSuffixes := []string{
		filepath.Join("views", "critical_user_model_view.rs"),
		filepath.Join("services", "critical_user_model_service.rs"),
		"defs_c.rs",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found (rust)", suffix)
		}
	}

	// CriticalUserModel has a BaseModel field → outbound expected.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for Rust CriticalUserModel (should reference BaseModel)")
	}

	assertSummaryConsistency(t, graph, "rust")
}

// -----------------------------------------------------------------------
// TypeScript: CriticalUserModel inbound/outbound assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_TypeScript_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "typescript")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in typescript")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "typescript", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for TypeScript CriticalUserModel, got %d", len(inboundEdges))
	}

	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes (typescript)", e.From)
		}
	}

	expectedSuffixes := []string{
		filepath.Join("views", "criticalUserModelView.ts"),
		filepath.Join("services", "criticalUserModelService.ts"),
		"defs_c.ts",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found (typescript)", suffix)
		}
	}

	// CriticalUserModel extends BaseModel → outbound expected.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for TypeScript CriticalUserModel (should reference BaseModel)")
	}

	assertSummaryConsistency(t, graph, "typescript")
}

// -----------------------------------------------------------------------
// JavaScript: CriticalUserModel inbound/outbound assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_JavaScript_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "javascript")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in javascript")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "javascript", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for JavaScript CriticalUserModel, got %d", len(inboundEdges))
	}

	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes (javascript)", e.From)
		}
	}

	expectedSuffixes := []string{
		filepath.Join("views", "criticalUserModelView.js"),
		filepath.Join("services", "criticalUserModelService.js"),
		"defs_c.js",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found (javascript)", suffix)
		}
	}

	// CriticalUserModel extends BaseModel → outbound expected.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for JavaScript CriticalUserModel (should reference BaseModel)")
	}

	assertSummaryConsistency(t, graph, "javascript")
}

// -----------------------------------------------------------------------
// Python: CriticalUserModel inbound/outbound assertions
// -----------------------------------------------------------------------

func TestBuildDependencyGraph_Python_CriticalUserModel_Inbound(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "python")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in python")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "python", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	inboundEdges := inboundEdgesFromGraph(graph)
	nodeIDs := nodeIDsInGraph(graph)

	if len(inboundEdges) < 3 {
		t.Errorf("expected >= 3 inbound edges for Python CriticalUserModel, got %d", len(inboundEdges))
	}

	for _, e := range inboundEdges {
		if !nodeIDs[e.From] {
			t.Errorf("inbound edge From=%q missing node in inbound.nodes (python)", e.From)
		}
	}

	expectedSuffixes := []string{
		filepath.Join("views", "critical_user_model_view.py"),
		filepath.Join("services", "critical_user_model_service.py"),
		"defs_c.py",
	}
	for _, suffix := range expectedSuffixes {
		found := false
		for _, e := range inboundEdges {
			if strings.HasSuffix(e.FilePath, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected inbound edge from file matching %q — not found (python)", suffix)
		}
	}

	// CriticalUserModel inherits BaseModel → outbound expected.
	outboundEdges := outboundEdgesFromGraph(graph)
	if len(outboundEdges) == 0 {
		t.Error("expected at least 1 outbound edge for Python CriticalUserModel (should reference BaseModel)")
	}

	assertSummaryConsistency(t, graph, "python")
}

// -----------------------------------------------------------------------
// Summary field consistency tests
// -----------------------------------------------------------------------

// assertSummaryConsistency verifies that the Inbound/Outbound/Metrics summary
// fields are consistent with the actual edge slices in those sections.
func assertSummaryConsistency(t *testing.T, graph *DependencyGraph, lang string) {
	t.Helper()

	inboundEdges := graph.Inbound.Edges
	outboundEdges := graph.Outbound.Edges

	// Metrics counts must match section edge counts.
	if graph.Metrics.InboundEdgeCount != len(inboundEdges) {
		t.Errorf("[%s] Metrics.InboundEdgeCount=%d, want %d",
			lang, graph.Metrics.InboundEdgeCount, len(inboundEdges))
	}
	if graph.Metrics.OutboundEdgeCount != len(outboundEdges) {
		t.Errorf("[%s] Metrics.OutboundEdgeCount=%d, want %d",
			lang, graph.Metrics.OutboundEdgeCount, len(outboundEdges))
	}
	if graph.Metrics.InboundUsageCount != len(graph.Usages) {
		t.Errorf("[%s] Metrics.InboundUsageCount=%d, want %d (Usages len)",
			lang, graph.Metrics.InboundUsageCount, len(graph.Usages))
	}

	// Inbound section totals must agree.
	if graph.Inbound.TotalUsages != len(graph.Usages) {
		t.Errorf("[%s] Inbound.TotalUsages=%d, want %d",
			lang, graph.Inbound.TotalUsages, len(graph.Usages))
	}

	// Each inbound edge node must appear in Inbound.Nodes.
	inboundNodeIDs := make(map[string]bool, len(graph.Inbound.Nodes))
	for _, n := range graph.Inbound.Nodes {
		inboundNodeIDs[n.ID] = true
	}
	for _, e := range inboundEdges {
		if !inboundNodeIDs[e.From] {
			t.Errorf("[%s] inbound edge From=%q has no node in Inbound.Nodes", lang, e.From)
		}
	}

	// Each outbound edge node must appear in Outbound.Nodes.
	outboundNodeIDs := make(map[string]bool, len(graph.Outbound.Nodes))
	for _, n := range graph.Outbound.Nodes {
		outboundNodeIDs[n.ID] = true
	}
	for _, e := range outboundEdges {
		if !outboundNodeIDs[e.To] {
			t.Errorf("[%s] outbound edge To=%q has no node in Outbound.Nodes", lang, e.To)
		}
	}

	// Scores must be in [0,1].
	if graph.Inbound.Score < 0 || graph.Inbound.Score > 1.0 {
		t.Errorf("[%s] Inbound.Score=%.4f out of [0,1]", lang, graph.Inbound.Score)
	}
	if graph.Outbound.Score < 0 || graph.Outbound.Score > 1.0 {
		t.Errorf("[%s] Outbound.Score=%.4f out of [0,1]", lang, graph.Outbound.Score)
	}
	if graph.Metrics.ImpactScore < 0 || graph.Metrics.ImpactScore > 1.0 {
		t.Errorf("[%s] Metrics.ImpactScore=%.4f out of [0,1]", lang, graph.Metrics.ImpactScore)
	}
	if graph.Metrics.CouplingScore < 0 || graph.Metrics.CouplingScore > 1.0 {
		t.Errorf("[%s] Metrics.CouplingScore=%.4f out of [0,1]", lang, graph.Metrics.CouplingScore)
	}

	// If inbound edges exist, impact score must be > 0.
	if len(inboundEdges) > 0 && graph.Metrics.ImpactScore <= 0 {
		t.Errorf("[%s] ImpactScore=%.4f but %d inbound edges exist — expected > 0",
			lang, graph.Metrics.ImpactScore, len(inboundEdges))
	}
}

// -----------------------------------------------------------------------
// Never-empty safeguard test
// -----------------------------------------------------------------------

// TestNeverEmptySafeguard verifies that inbound edges are returned even
// when minScore is set very high (e.g., 0.99), by falling back to top-per-file.
func TestNeverEmptySafeguard(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	// Use a symbol known to have multiple usages.
	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel definition not found in go")
	}

	// Build with an impossibly high minScore so the normal filter drops everything.
	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "go", repoRoot, 1, 0.99, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	// Verify: all raw usages score 1.0 (single candidate), so they should pass 0.99.
	// But even if they didn't, the safeguard should preserve at least one per file.
	inboundEdges := graph.Inbound.Edges
	inboundNodeIDs := make(map[string]bool, len(graph.Inbound.Nodes))
	for _, n := range graph.Inbound.Nodes {
		inboundNodeIDs[n.ID] = true
	}

	if len(inboundEdges) == 0 {
		t.Error("expected at least 1 inbound edge even with high minScore (never-empty safeguard)")
	}
	for _, e := range inboundEdges {
		if !inboundNodeIDs[e.From] {
			t.Errorf("never-empty safeguard: inbound edge From=%q missing node in Inbound.Nodes", e.From)
		}
	}
}

// -----------------------------------------------------------------------
// Inbound TotalFiles matches unique file count
// -----------------------------------------------------------------------

func TestInboundTotalFilesMatchesUniqueFiles(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	langCases := []struct {
		lang string
		sym  string
	}{
		{"go", "CriticalUserModel"},
		{"java", "CriticalUserModel"},
		{"rust", "CriticalUserModel"},
		{"typescript", "CriticalUserModel"},
		{"javascript", "CriticalUserModel"},
		{"python", "CriticalUserModel"},
	}

	for _, tc := range langCases {
		t.Run(tc.lang, func(t *testing.T) {
			defs, err := nav.GoToDefinitionByName(tc.sym, "", tc.lang)
			if err != nil || len(defs) == 0 {
				t.Skipf("definition not found")
			}

			graph, err := BuildDependencyGraph(nav, &defs[0], defs, tc.lang, repoRoot, 1, 0.0, 500)
			if err != nil {
				t.Fatalf("BuildDependencyGraph: %v", err)
			}

			// Count unique files in inbound edges.
			uniqueFiles := map[string]bool{}
			for _, e := range graph.Inbound.Edges {
				uniqueFiles[e.FilePath] = true
			}

			if graph.Inbound.TotalFiles != len(uniqueFiles) {
				t.Errorf("[%s] Inbound.TotalFiles=%d, want %d (unique inbound files)",
					tc.lang, graph.Inbound.TotalFiles, len(uniqueFiles))
			}
		})
	}
}

// -----------------------------------------------------------------------
// Metrics consistency: Inbound.Edges and Outbound.Edges match Metrics counts
// -----------------------------------------------------------------------

// TestMetricsMatchSectionEdgeCounts verifies that Metrics counters equal the
// actual lengths of the typed edge sections — the canonical source of truth.
func TestMetricsMatchSectionEdgeCounts(t *testing.T) {
	nav, repoRoot, cleanup := setupInboundTest(t)
	defer cleanup()

	defs, err := nav.GoToDefinitionByName("CriticalUserModel", "", "go")
	if err != nil || len(defs) == 0 {
		t.Skip("CriticalUserModel not found")
	}

	graph, err := BuildDependencyGraph(nav, &defs[0], defs, "go", repoRoot, 1, 0.0, 500)
	if err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}

	if graph.Metrics.InboundEdgeCount != len(graph.Inbound.Edges) {
		t.Errorf("Metrics.InboundEdgeCount=%d != len(Inbound.Edges)=%d",
			graph.Metrics.InboundEdgeCount, len(graph.Inbound.Edges))
	}
	if graph.Metrics.OutboundEdgeCount != len(graph.Outbound.Edges) {
		t.Errorf("Metrics.OutboundEdgeCount=%d != len(Outbound.Edges)=%d",
			graph.Metrics.OutboundEdgeCount, len(graph.Outbound.Edges))
	}
	if graph.Metrics.InboundUsageCount != len(graph.Usages) {
		t.Errorf("Metrics.InboundUsageCount=%d != len(Usages)=%d",
			graph.Metrics.InboundUsageCount, len(graph.Usages))
	}

	// Both sections must be non-empty for a well-connected symbol.
	if len(graph.Inbound.Edges) == 0 {
		t.Error("expected non-empty Inbound.Edges")
	}
	if len(graph.Outbound.Edges) == 0 {
		t.Error("expected non-empty Outbound.Edges")
	}
}
