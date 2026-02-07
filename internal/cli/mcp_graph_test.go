package cli

import (
	"strings"
	"testing"

	"github.com/codeintelx/cli/internal/indexer"
)

func TestFormatDependencyGraph_Mermaid(t *testing.T) {
	graph := &indexer.DependencyGraph{
		PrimaryDefinition: &indexer.DefinitionResult{
			Name: "TestFunc",
			Kind: "function",
			Location: indexer.Location{
				Path:      "pkg/foo.go",
				StartLine: 10,
				StartCol:  5,
			},
			Signature: "func TestFunc() error",
		},
		DefinitionCandidates: []indexer.DefinitionResult{
			{Name: "TestFunc", Kind: "function", Location: indexer.Location{Path: "pkg/foo.go", StartLine: 10}},
		},
		SymbolGraph: indexer.SymbolGraph{
			Nodes: []indexer.DepGraphNode{
				{ID: "pkg/foo.go:TestFunc:10", Name: "TestFunc", Kind: "function", Path: "pkg/foo.go", StartLine: 10, EndLine: 20},
			},
			Edges: []indexer.DepGraphEdge{
				{From: "pkg/bar.go", To: "pkg/foo.go:TestFunc:10", Kind: "inbound", Score: 0.85, Count: 3, FilePath: "pkg/bar.go"},
				{From: "pkg/foo.go:TestFunc:10", To: "pkg/util.go:Helper:5", Kind: "outbound", Score: 0.75, Count: 2, FilePath: "pkg/util.go"},
			},
		},
		FileGraph: []indexer.FileGraphEdge{
			{From: "pkg/bar.go", To: "pkg/foo.go", Score: 0.85, Count: 3},
			{From: "pkg/foo.go", To: "pkg/util.go", Score: 0.75, Count: 2},
		},
		Usages: []indexer.ScoredUsage{
			{
				UsageResult: indexer.UsageResult{
					Name:     "TestFunc",
					Location: indexer.Location{Path: "pkg/bar.go", StartLine: 5},
				},
				DependencyScore: 0.85,
			},
		},
	}

	text := formatDependencyGraph(graph)

	// Check for Mermaid syntax
	if !strings.Contains(text, "```mermaid") {
		t.Error("expected Mermaid code blocks in output")
	}
	if !strings.Contains(text, "graph LR") {
		t.Error("expected file dependency graph (graph LR)")
	}
	if !strings.Contains(text, "graph TD") {
		t.Error("expected symbol dependency graph (graph TD)")
	}

	// Check for key content
	if !strings.Contains(text, "TestFunc") {
		t.Error("expected symbol name in output")
	}
	if !strings.Contains(text, "0.85") {
		t.Error("expected scores in output")
	}
	if !strings.Contains(text, "File Dependency Graph") {
		t.Error("expected file graph section")
	}
	if !strings.Contains(text, "Symbol Dependency Graph") {
		t.Error("expected symbol graph section")
	}
	if !strings.Contains(text, "Top Scored Usages") {
		t.Error("expected usages section")
	}
}

func TestFormatDependencyGraph_EmptyGraph(t *testing.T) {
	graph := &indexer.DependencyGraph{
		PrimaryDefinition:    nil,
		DefinitionCandidates: []indexer.DefinitionResult{},
		SymbolGraph:          indexer.SymbolGraph{Nodes: []indexer.DepGraphNode{}, Edges: []indexer.DepGraphEdge{}},
		FileGraph:            []indexer.FileGraphEdge{},
		Usages:               []indexer.ScoredUsage{},
	}

	text := formatDependencyGraph(graph)
	if text == "" {
		t.Error("expected non-empty output for empty graph")
	}
	if !strings.Contains(text, "Dependency Graph") {
		t.Error("expected title in output")
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		path   string
		maxLen int
		want   string
	}{
		{"short.go", 20, "short.go"},
		{"a/very/long/path/to/some/file.go", 20, "a/very/...e/file.go"},
		{"test.go", 5, "te..."},
	}
	for _, tt := range tests {
		got := shortenPath(tt.path, tt.maxLen)
		if len(got) > tt.maxLen {
			t.Errorf("shortenPath(%q, %d) = %q (len %d), exceeds maxLen", tt.path, tt.maxLen, got, len(got))
		}
	}
}

func TestSanitizeMermaidID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with-dash", "with_dash"},
		{"with/slash", "with_slash"},
		{"with:colon", "with_colon"},
		{"123start", "n123start"},
		{"", "node"},
	}
	for _, tt := range tests {
		got := sanitizeMermaidID(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeMermaidID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractSymbolFromNodeID(t *testing.T) {
	tests := []struct {
		nodeID string
		want   string
	}{
		{"pkg/foo.go:MyFunc:42", "MyFunc"},
		{"file.go:Struct:10", "Struct"},
		{"nopath:Name:5", "Name"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		got := extractSymbolFromNodeID(tt.nodeID)
		if got != tt.want {
			t.Errorf("extractSymbolFromNodeID(%q) = %q, want %q", tt.nodeID, got, tt.want)
		}
	}
}
