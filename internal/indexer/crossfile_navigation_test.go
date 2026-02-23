package indexer

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesdx/cli/internal/db"
)

// setupCrossFileTest indexes testdata and returns a Navigator.
func setupCrossFileTest(t *testing.T) (*Navigator, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "crossfile.db")
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
	return nav, func() { _ = d.Close() }
}

// crossFileCase describes a single cross-file go-to-definition test.
type crossFileCase struct {
	name string
	lang string

	// Call site: file containing the usage / call
	callFile string

	// Expected definition
	defFile     string
	defSymbol   string
	defKindLike string // substring of kind (e.g. "function", "method")
}

func TestCrossFileGoToDefinitionByName(t *testing.T) {
	nav, cleanup := setupCrossFileTest(t)
	defer cleanup()

	cases := []struct {
		name    string
		lang    string
		defFile string // substring of expected path
		defLine int
		kind    string
	}{
		// Python module-level functions
		{"process_user_data", "python", "services/user_service.py", 7, "function"},
		{"validate_email", "python", "services/user_service.py", 14, "function"},
		{"format_user_name", "python", "services/user_service.py", 19, "function"},

		// Go top-level functions
		{"ProcessUserData", "go", "services/user_service.go", 5, "function"},
		{"ValidateEmail", "go", "services/user_service.go", 12, "function"},
		{"FormatUserName", "go", "services/user_service.go", 21, "function"},

		// Rust free functions
		{"process_user_data", "rust", "services/user_service.rs", 3, "function"},
		{"validate_email", "rust", "services/user_service.rs", 7, "function"},
		{"format_user_name", "rust", "services/user_service.rs", 11, "function"},

		// TypeScript exported functions
		{"processUserData", "typescript", "services/userService.ts", 3, "function"},
		{"validateEmail", "typescript", "services/userService.ts", 7, "function"},

		// JavaScript functions
		{"processUserData", "javascript", "services/userService.js", 3, "function"},
		{"validateEmail", "javascript", "services/userService.js", 7, "function"},

		// Java static methods
		{"processUserData", "java", "services/UserService.java", 9, "method"},
		{"validateEmail", "java", "services/UserService.java", 13, "method"},
		{"formatUserName", "java", "services/UserService.java", 17, "method"},
	}

	for _, tt := range cases {
		t.Run(tt.name+"_"+tt.lang, func(t *testing.T) {
			results, err := nav.GoToDefinitionByName(tt.name, "", tt.lang)
			if err != nil {
				t.Fatalf("GoToDefinitionByName(%q, %q): %v", tt.name, tt.lang, err)
			}
			if len(results) == 0 {
				t.Fatalf("GoToDefinitionByName(%q, %q) returned 0 results", tt.name, tt.lang)
			}

			found := false
			for _, r := range results {
				if strings.HasSuffix(r.Location.Path, tt.defFile) &&
					r.Location.StartLine == tt.defLine &&
					r.Kind == tt.kind {
					found = true
					break
				}
			}
			if !found {
				for _, r := range results {
					t.Logf("  got: %s (%s) at %s:%d", r.Name, r.Kind, r.Location.Path, r.Location.StartLine)
				}
				t.Errorf("expected definition in %s:%d kind=%s; not found", tt.defFile, tt.defLine, tt.kind)
			}
		})
	}
}

// TestCrossFileGoToDefinitionByPositionMidIdentifier simulates a user clicking
// in the MIDDLE of an identifier (not at start_col). This is the real-world
// scenario that triggers the "module-level function returns null" bug.
func TestCrossFileGoToDefinitionByPositionMidIdentifier(t *testing.T) {
	nav, cleanup := setupCrossFileTest(t)
	defer cleanup()

	type midCase struct {
		name      string
		lang      string
		callFile  string
		symbol    string
		defFile   string
		defKind   string
		colOffset int // offset into the identifier from start_col
	}

	cases := []midCase{
		// Click in the middle of "process_user_data" (offset +5 = "s" in "process")
		{"python_mid_process_user_data", "python", "python/views/user_views.py", "process_user_data", "services/user_service.py", "function", 5},
		{"go_mid_ProcessUserData", "go", "go/views/user_views.go", "ProcessUserData", "services/user_service.go", "function", 7},
		{"rust_mid_process_user_data", "rust", "rust/views/user_views.rs", "process_user_data", "services/user_service.rs", "function", 8},
		{"ts_mid_processUserData", "typescript", "ts/views/userViews.ts", "processUserData", "services/userService.ts", "function", 7},
		{"js_mid_processUserData", "javascript", "js/views/userViews.js", "processUserData", "services/userService.js", "function", 7},
		{"java_mid_processUserData", "java", "java/views/UserViews.java", "processUserData", "services/UserService.java", "method", 7},

		// Click on the last character of the identifier (offset = len-1)
		{"python_end_validate_email", "python", "python/views/user_views.py", "validate_email", "services/user_service.py", "function", 13},
		{"go_end_ValidateEmail", "go", "go/views/user_views.go", "ValidateEmail", "services/user_service.go", "function", 12},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Find the ref position in the call file
			refs, err := nav.FindUsagesByName(tt.symbol, "", tt.lang)
			if err != nil {
				t.Fatalf("FindUsagesByName(%q): %v", tt.symbol, err)
			}

			var refLine, refCol int
			var refPath string
			for _, r := range refs {
				if strings.Contains(r.Location.Path, tt.callFile) {
					refLine = r.Location.StartLine
					refCol = r.Location.StartCol
					refPath = r.Location.Path
					break
				}
			}
			if refPath == "" {
				t.Fatalf("no ref for %q found in %s", tt.symbol, tt.callFile)
			}

			// Apply offset to simulate mid-identifier cursor
			midCol := refCol + tt.colOffset

			defs, err := nav.GoToDefinitionByPosition(refPath, refLine, midCol, tt.lang)
			if err != nil {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d): %v", refPath, refLine, midCol, err)
			}
			if len(defs) == 0 {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d) returned 0 results (cursor at mid-identifier)",
					refPath, refLine, midCol)
			}

			found := false
			for _, d := range defs {
				if strings.HasSuffix(d.Location.Path, tt.defFile) && d.Name == tt.symbol {
					found = true
					if d.Kind != tt.defKind {
						t.Errorf("expected kind %q, got %q", tt.defKind, d.Kind)
					}
					break
				}
			}
			if !found {
				for _, d := range defs {
					t.Logf("  got: %s (%s) at %s:%d", d.Name, d.Kind, d.Location.Path, d.Location.StartLine)
				}
				t.Errorf("expected definition %q in %q; not found", tt.symbol, tt.defFile)
			}
		})
	}
}

// TestCrossFileGoToDefinitionByPositionHardcoded simulates the exact MCP scenario:
// a client provides a file path and line+col without using DB ref positions.
// This tests the full resolution chain including extractWordAtPosition fallback.
func TestCrossFileGoToDefinitionByPositionHardcoded(t *testing.T) {
	nav, cleanup := setupCrossFileTest(t)
	defer cleanup()

	// Python: views/user_views.py line 8 has:
	//   "        result = process_user_data(user_id, data)"
	// "process_user_data" starts at col 17, ends at col 35 (exclusive).
	// We'll try various cols within the identifier.
	pyFile := filepath.Join("python", "views", "user_views.py")
	pyDefFile := filepath.Join("python", "services", "user_service.py")

	pythonCases := []struct {
		desc   string
		line   int
		col    int
		symbol string
	}{
		{"start of process_user_data", 8, 17, "process_user_data"},
		{"mid process_user_data", 8, 25, "process_user_data"},
		{"end-1 process_user_data", 8, 33, "process_user_data"},
		{"validate_email call", 13, 15, "validate_email"},
		{"format_user_name in standalone", 17, 11, "format_user_name"},
	}

	for _, tt := range pythonCases {
		t.Run("python_"+tt.desc, func(t *testing.T) {
			defs, err := nav.GoToDefinitionByPosition(pyFile, tt.line, tt.col, "python")
			if err != nil {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d): %v", pyFile, tt.line, tt.col, err)
			}
			if len(defs) == 0 {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d) returned 0 results", pyFile, tt.line, tt.col)
			}
			found := false
			for _, d := range defs {
				if d.Name == tt.symbol && strings.HasSuffix(d.Location.Path, pyDefFile) {
					found = true
					break
				}
			}
			if !found {
				for _, d := range defs {
					t.Logf("  got: %s (%s) at %s:%d", d.Name, d.Kind, d.Location.Path, d.Location.StartLine)
				}
				t.Errorf("expected %q in %s", tt.symbol, pyDefFile)
			}
		})
	}

	// Go: views/user_views.go line 6 has:
	//   "\tresult := services.ProcessUserData(userID, data)"
	// "ProcessUserData" starts at col 20 (tab=1 char + "result := services." = 20)
	goFile := filepath.Join("go", "views", "user_views.go")
	goDefFile := filepath.Join("go", "services", "user_service.go")

	goCases := []struct {
		desc   string
		line   int
		col    int
		symbol string
	}{
		{"ProcessUserData qualified call", 6, 21, "ProcessUserData"},
		{"ValidateEmail qualified call", 11, 20, "ValidateEmail"},
	}

	for _, tt := range goCases {
		t.Run("go_"+tt.desc, func(t *testing.T) {
			defs, err := nav.GoToDefinitionByPosition(goFile, tt.line, tt.col, "go")
			if err != nil {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d): %v", goFile, tt.line, tt.col, err)
			}
			if len(defs) == 0 {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d) returned 0 results", goFile, tt.line, tt.col)
			}
			found := false
			for _, d := range defs {
				if d.Name == tt.symbol && strings.HasSuffix(d.Location.Path, goDefFile) {
					found = true
					break
				}
			}
			if !found {
				for _, d := range defs {
					t.Logf("  got: %s (%s) at %s:%d", d.Name, d.Kind, d.Location.Path, d.Location.StartLine)
				}
				t.Errorf("expected %q in %s", tt.symbol, goDefFile)
			}
		})
	}
}

// TestExtractWordAtPosition verifies the source-file fallback for identifier extraction.
func TestExtractWordAtPosition(t *testing.T) {
	repoRoot := testdataDir(t)

	cases := []struct {
		name     string
		relPath  string
		line     int
		col      int
		expected string
	}{
		// Python: "def process_user_data(user_id: int, data: dict) -> bool:"
		// "process_user_data" starts at col 4
		{"python_func_start", "python/services/user_service.py", 7, 4, "process_user_data"},
		{"python_func_mid", "python/services/user_service.py", 7, 12, "process_user_data"},
		// Go: "func ProcessUserData(userID int, data map[string]string) bool {"
		// "ProcessUserData" starts at col 5
		{"go_func_start", "go/services/user_service.go", 5, 5, "ProcessUserData"},
		{"go_func_mid", "go/services/user_service.go", 5, 12, "ProcessUserData"},
		// Cursor on whitespace should return empty
		{"whitespace", "python/services/user_service.py", 2, 0, ""},
		// Cursor past end of line
		{"past_end", "python/services/user_service.py", 7, 200, ""},
		// TypeScript: "export function processUserData(...)"
		{"ts_func", "ts/services/userService.ts", 3, 20, "processUserData"},
		// Rust: "pub fn process_user_data(user_id: u64, data: &str) -> bool {"
		{"rust_func", "rust/services/user_service.rs", 3, 10, "process_user_data"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			absPath := filepath.Join(repoRoot, tt.relPath)
			got, err := extractWordAtPosition(absPath, tt.line, tt.col)
			if err != nil {
				t.Fatalf("extractWordAtPosition: %v", err)
			}
			if got != tt.expected {
				t.Errorf("extractWordAtPosition(%s:%d:%d) = %q, want %q",
					tt.relPath, tt.line, tt.col, got, tt.expected)
			}
		})
	}
}

// TestIdentifierAtFallback ensures that identifierAt falls through to
// the source-file fallback when the DB has no matching ref/symbol at the cursor.
func TestIdentifierAtFallback(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fallback.db")
	if err := db.Initialize(dbPath); err != nil {
		t.Fatalf("db.Initialize: %v", err)
	}
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer func() { _ = d.Close() }()

	repoRoot := testdataDir(t)
	idx := New(d, repoRoot)
	if _, err := idx.FullIndex([]string{"."}); err != nil {
		t.Fatalf("FullIndex: %v", err)
	}

	nav := &Navigator{DB: d, ProjectID: idx.Store.ProjectID, RepoRoot: repoRoot}

	// Use a position on an identifier in an indexed file.
	// Even if DB happens to miss this particular position, the fallback should find it.
	pyFile := filepath.Join("python", "services", "user_service.py")
	name, err := nav.identifierAt(pyFile, 7, 4) // "def process_user_data(...)"
	if err != nil {
		t.Fatalf("identifierAt: %v", err)
	}
	if name != "process_user_data" {
		t.Errorf("identifierAt = %q, want %q", name, "process_user_data")
	}
}

// TestEmptyResultsAreNotNil verifies that queries returning no rows produce
// an empty slice (marshals to "[]") rather than nil (marshals to "null").
func TestEmptyResultsAreNotNil(t *testing.T) {
	nav, cleanup := setupCrossFileTest(t)
	defer cleanup()

	// Query a symbol name that doesn't exist
	defs, err := nav.GoToDefinitionByName("CompletelyNonExistentSymbol99999", "", "python")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if defs == nil {
		t.Error("GoToDefinitionByName returned nil, want empty slice")
	}
	if len(defs) != 0 {
		t.Errorf("GoToDefinitionByName returned %d results, want 0", len(defs))
	}

	usages, err := nav.FindUsagesByName("CompletelyNonExistentSymbol99999", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}
	if usages == nil {
		t.Error("FindUsagesByName returned nil, want empty slice")
	}
	if len(usages) != 0 {
		t.Errorf("FindUsagesByName returned %d results, want 0", len(usages))
	}

	// Verify JSON marshaling produces "[]" not "null"
	toJSON := func(v interface{}) string {
		data, _ := json.Marshal(v)
		return string(data)
	}

	defsJSON := toJSON(defs)
	if defsJSON != "[]" {
		t.Errorf("definitions JSON = %s, want []", defsJSON)
	}
	usagesJSON := toJSON(usages)
	if usagesJSON != "[]" {
		t.Errorf("usages JSON = %s, want []", usagesJSON)
	}
}

func TestCrossFileGoToDefinitionByPosition(t *testing.T) {
	nav, cleanup := setupCrossFileTest(t)
	defer cleanup()

	// For each call site, we first look up the ref from the DB to get exact position,
	// then test GoToDefinitionByPosition from that position.
	cases := []crossFileCase{
		// Python: call sites in views/user_views.py
		{
			name:        "python_process_user_data_call",
			lang:        "python",
			callFile:    "python/views/user_views.py",
			defFile:     "services/user_service.py",
			defSymbol:   "process_user_data",
			defKindLike: "function",
		},
		{
			name:        "python_validate_email_call",
			lang:        "python",
			callFile:    "python/views/user_views.py",
			defFile:     "services/user_service.py",
			defSymbol:   "validate_email",
			defKindLike: "function",
		},
		{
			name:        "python_format_user_name_call",
			lang:        "python",
			callFile:    "python/views/user_views.py",
			defFile:     "services/user_service.py",
			defSymbol:   "format_user_name",
			defKindLike: "function",
		},

		// Go: call sites in views/user_views.go
		{
			name:        "go_ProcessUserData_call",
			lang:        "go",
			callFile:    "go/views/user_views.go",
			defFile:     "services/user_service.go",
			defSymbol:   "ProcessUserData",
			defKindLike: "function",
		},
		{
			name:        "go_ValidateEmail_call",
			lang:        "go",
			callFile:    "go/views/user_views.go",
			defFile:     "services/user_service.go",
			defSymbol:   "ValidateEmail",
			defKindLike: "function",
		},
		{
			name:        "go_FormatUserName_call",
			lang:        "go",
			callFile:    "go/views/user_views.go",
			defFile:     "services/user_service.go",
			defSymbol:   "FormatUserName",
			defKindLike: "function",
		},

		// Rust: call sites in views/user_views.rs
		{
			name:        "rust_process_user_data_call",
			lang:        "rust",
			callFile:    "rust/views/user_views.rs",
			defFile:     "services/user_service.rs",
			defSymbol:   "process_user_data",
			defKindLike: "function",
		},
		{
			name:        "rust_validate_email_call",
			lang:        "rust",
			callFile:    "rust/views/user_views.rs",
			defFile:     "services/user_service.rs",
			defSymbol:   "validate_email",
			defKindLike: "function",
		},
		{
			name:        "rust_format_user_name_call",
			lang:        "rust",
			callFile:    "rust/views/user_views.rs",
			defFile:     "services/user_service.rs",
			defSymbol:   "format_user_name",
			defKindLike: "function",
		},

		// TypeScript: call sites in views/userViews.ts
		{
			name:        "ts_processUserData_call",
			lang:        "typescript",
			callFile:    "ts/views/userViews.ts",
			defFile:     "services/userService.ts",
			defSymbol:   "processUserData",
			defKindLike: "function",
		},
		{
			name:        "ts_validateEmail_call",
			lang:        "typescript",
			callFile:    "ts/views/userViews.ts",
			defFile:     "services/userService.ts",
			defSymbol:   "validateEmail",
			defKindLike: "function",
		},
		{
			name:        "ts_formatUserName_call",
			lang:        "typescript",
			callFile:    "ts/views/userViews.ts",
			defFile:     "services/userService.ts",
			defSymbol:   "formatUserName",
			defKindLike: "function",
		},

		// JavaScript: call sites in views/userViews.js
		{
			name:        "js_processUserData_call",
			lang:        "javascript",
			callFile:    "js/views/userViews.js",
			defFile:     "services/userService.js",
			defSymbol:   "processUserData",
			defKindLike: "function",
		},
		{
			name:        "js_validateEmail_call",
			lang:        "javascript",
			callFile:    "js/views/userViews.js",
			defFile:     "services/userService.js",
			defSymbol:   "validateEmail",
			defKindLike: "function",
		},
		{
			name:        "js_formatUserName_call",
			lang:        "javascript",
			callFile:    "js/views/userViews.js",
			defFile:     "services/userService.js",
			defSymbol:   "formatUserName",
			defKindLike: "function",
		},

		// Java: call sites in views/UserViews.java
		{
			name:        "java_processUserData_call",
			lang:        "java",
			callFile:    "java/views/UserViews.java",
			defFile:     "services/UserService.java",
			defSymbol:   "processUserData",
			defKindLike: "method",
		},
		{
			name:        "java_validateEmail_call",
			lang:        "java",
			callFile:    "java/views/UserViews.java",
			defFile:     "services/UserService.java",
			defSymbol:   "validateEmail",
			defKindLike: "method",
		},
		{
			name:        "java_formatUserName_call",
			lang:        "java",
			callFile:    "java/views/UserViews.java",
			defFile:     "services/UserService.java",
			defSymbol:   "formatUserName",
			defKindLike: "method",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Find the ref position for the symbol in the call file.
			refResults, err := nav.FindUsagesByName(tt.defSymbol, "", tt.lang)
			if err != nil {
				t.Fatalf("FindUsagesByName(%q): %v", tt.defSymbol, err)
			}

			var refLine, refCol int
			var refPath string
			for _, r := range refResults {
				if strings.Contains(r.Location.Path, tt.callFile) && r.Kind == "call" {
					refLine = r.Location.StartLine
					refCol = r.Location.StartCol
					refPath = r.Location.Path
					break
				}
			}
			if refPath == "" {
				// Try with any ref kind (some languages use "identifier" rather than "call")
				for _, r := range refResults {
					if strings.Contains(r.Location.Path, tt.callFile) {
						refLine = r.Location.StartLine
						refCol = r.Location.StartCol
						refPath = r.Location.Path
						break
					}
				}
			}
			if refPath == "" {
				t.Fatalf("no ref for %q found in %s (total refs=%d)", tt.defSymbol, tt.callFile, len(refResults))
			}

			// Step 2: GoToDefinitionByPosition from the ref position.
			defs, err := nav.GoToDefinitionByPosition(refPath, refLine, refCol, tt.lang)
			if err != nil {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d, %s): %v", refPath, refLine, refCol, tt.lang, err)
			}
			if len(defs) == 0 {
				t.Fatalf("GoToDefinitionByPosition(%s:%d:%d, %s) returned 0 results", refPath, refLine, refCol, tt.lang)
			}

			// Step 3: Verify the definition resolves to the service file.
			found := false
			for _, d := range defs {
				if strings.HasSuffix(d.Location.Path, tt.defFile) && d.Name == tt.defSymbol {
					found = true
					if !strings.Contains(d.Kind, tt.defKindLike) {
						t.Errorf("expected kind containing %q, got %q", tt.defKindLike, d.Kind)
					}
					break
				}
			}
			if !found {
				for _, d := range defs {
					t.Logf("  got: %s (%s) at %s:%d", d.Name, d.Kind, d.Location.Path, d.Location.StartLine)
				}
				t.Errorf("expected definition %q in file matching %q; not found among %d results",
					tt.defSymbol, tt.defFile, len(defs))
			}
		})
	}
}
