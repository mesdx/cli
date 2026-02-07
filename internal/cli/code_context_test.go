package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeWindows(t *testing.T) {
	tests := []struct {
		name     string
		windows  []codeWindow
		expected []codeWindow
	}{
		{
			name:     "empty",
			windows:  []codeWindow{},
			expected: nil,
		},
		{
			name:     "single window",
			windows:  []codeWindow{{startLine: 1, endLine: 5}},
			expected: []codeWindow{{startLine: 1, endLine: 5}},
		},
		{
			name: "overlapping windows",
			windows: []codeWindow{
				{startLine: 1, endLine: 5},
				{startLine: 3, endLine: 8},
			},
			expected: []codeWindow{{startLine: 1, endLine: 8}},
		},
		{
			name: "adjacent windows (no gap)",
			windows: []codeWindow{
				{startLine: 1, endLine: 5},
				{startLine: 6, endLine: 10},
			},
			expected: []codeWindow{{startLine: 1, endLine: 10}},
		},
		{
			name: "windows with gap",
			windows: []codeWindow{
				{startLine: 1, endLine: 5},
				{startLine: 8, endLine: 10},
			},
			expected: []codeWindow{
				{startLine: 1, endLine: 5},
				{startLine: 8, endLine: 10},
			},
		},
		{
			name: "multiple overlapping",
			windows: []codeWindow{
				{startLine: 1, endLine: 3},
				{startLine: 2, endLine: 5},
				{startLine: 4, endLine: 8},
				{startLine: 10, endLine: 12},
			},
			expected: []codeWindow{
				{startLine: 1, endLine: 8},
				{startLine: 10, endLine: 12},
			},
		},
		{
			name: "unsorted input",
			windows: []codeWindow{
				{startLine: 10, endLine: 12},
				{startLine: 1, endLine: 3},
				{startLine: 5, endLine: 7},
			},
			expected: []codeWindow{
				{startLine: 1, endLine: 3},
				{startLine: 5, endLine: 7},
				{startLine: 10, endLine: 12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeWindows(tt.windows)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d windows, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("window %d: expected %+v, got %+v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestSafeJoinPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		repoRoot string
		relPath  string
		wantErr  bool
	}{
		{
			name:     "simple path",
			repoRoot: tmpDir,
			relPath:  "foo/bar.go",
			wantErr:  false,
		},
		{
			name:     "path with dots",
			repoRoot: tmpDir,
			relPath:  "foo/../bar.go",
			wantErr:  false,
		},
		{
			name:     "path traversal attempt",
			repoRoot: tmpDir,
			relPath:  "../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "absolute path",
			repoRoot: tmpDir,
			relPath:  "/etc/passwd",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeJoinPath(tt.repoRoot, tt.relPath)
			if tt.wantErr {
				if result != "" {
					t.Errorf("expected empty result for unsafe path, got %q", result)
				}
			} else {
				if result == "" {
					t.Errorf("expected non-empty result for safe path")
				}
			}
		})
	}
}

func TestReadFileLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `line 1
line 2
line 3
line 4
line 5
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startLine int
		endLine   int
		wantLines int
	}{
		{
			name:      "single line",
			startLine: 2,
			endLine:   2,
			wantLines: 1,
		},
		{
			name:      "multiple lines",
			startLine: 2,
			endLine:   4,
			wantLines: 3,
		},
		{
			name:      "first line",
			startLine: 1,
			endLine:   1,
			wantLines: 1,
		},
		{
			name:      "all lines",
			startLine: 1,
			endLine:   5,
			wantLines: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := readFileLines(testFile, tt.startLine, tt.endLine)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Count lines in result (each line has a line number prefix)
			lines := 0
			for _, c := range result {
				if c == '\n' {
					lines++
				}
			}
			if lines != tt.wantLines {
				t.Errorf("expected %d lines, got %d", tt.wantLines, lines)
			}
		})
	}
}

func TestReadFileAllLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `line 1
line 2
line 3
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := readFileAllLines(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "line 1" || lines[1] != "line 2" || lines[2] != "line 3" {
		t.Errorf("unexpected line content")
	}
}
