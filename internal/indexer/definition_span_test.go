package indexer

import (
	"path/filepath"
	"testing"

	"github.com/mesdx/cli/internal/db"
)

// setupSpanTest creates a fully indexed test DB from testdata and returns a Navigator.
func setupSpanTest(t *testing.T) (*Navigator, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "span_test.db")
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
	return nav, func() { _ = d.Close() }
}

// ---------- Go span tests ----------

func TestGoStructSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("Config", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	// Should find Config in defs_a.go (struct)
	found := false
	for _, r := range results {
		if r.Kind == "struct" {
			found = true
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("Config struct: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			// Config struct spans lines 7-11 in defs_a.go
			if r.Location.StartLine != 7 {
				t.Errorf("Config struct StartLine: got %d, want 7", r.Location.StartLine)
			}
			if r.Location.EndLine != 11 {
				t.Errorf("Config struct EndLine: got %d, want 11", r.Location.EndLine)
			}
		}
	}
	if !found {
		t.Error("Config struct not found in results")
	}
}

func TestGoInterfaceSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("Formatter", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for Formatter")
	}
	r := results[0]
	if r.Kind != "interface" {
		t.Fatalf("expected interface, got %s", r.Kind)
	}
	// Formatter interface spans lines 25-28
	if r.Location.StartLine != 25 {
		t.Errorf("Formatter StartLine: got %d, want 25", r.Location.StartLine)
	}
	if r.Location.EndLine != 28 {
		t.Errorf("Formatter EndLine: got %d, want 28", r.Location.EndLine)
	}
}

func TestGoMethodSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("Validate", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for Validate")
	}
	r := results[0]
	if r.Kind != "method" {
		t.Fatalf("expected method, got %s", r.Kind)
	}
	// Validate method spans lines 14-22
	if r.Location.StartLine != 14 {
		t.Errorf("Validate StartLine: got %d, want 14", r.Location.StartLine)
	}
	if r.Location.EndLine != 22 {
		t.Errorf("Validate EndLine: got %d, want 22", r.Location.EndLine)
	}
}

func TestGoFunctionSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("NewConfig", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for NewConfig")
	}
	r := results[0]
	if r.Kind != "function" {
		t.Fatalf("expected function, got %s", r.Kind)
	}
	// NewConfig func spans lines 37-43
	if r.Location.StartLine != 37 {
		t.Errorf("NewConfig StartLine: got %d, want 37", r.Location.StartLine)
	}
	if r.Location.EndLine != 43 {
		t.Errorf("NewConfig EndLine: got %d, want 43", r.Location.EndLine)
	}
}

func TestGoGroupedTypeSpec(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// Endpoint is in a grouped type(...) block
	results, err := nav.GoToDefinitionByName("Endpoint", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for Endpoint")
	}
	r := results[0]
	if r.Kind != "struct" {
		t.Fatalf("expected struct, got %s", r.Kind)
	}
	// Endpoint struct spans lines 47-50 (single spec, not entire group)
	if r.Location.StartLine != 47 {
		t.Errorf("Endpoint StartLine: got %d, want 47", r.Location.StartLine)
	}
	if r.Location.EndLine != 50 {
		t.Errorf("Endpoint EndLine: got %d, want 50", r.Location.EndLine)
	}
}

func TestGoConstAndVarSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// DefaultPort is a single-line const
	results, err := nav.GoToDefinitionByName("DefaultPort", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for DefaultPort")
	}
	r := results[0]
	if r.Location.StartLine != 31 {
		t.Errorf("DefaultPort StartLine: got %d, want 31", r.Location.StartLine)
	}
	// Single line — EndLine = StartLine
	if r.Location.EndLine != 31 {
		t.Errorf("DefaultPort EndLine: got %d, want 31", r.Location.EndLine)
	}

	// AppName is a single-line var
	results, err = nav.GoToDefinitionByName("AppName", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppName")
	}
	r = results[0]
	if r.Location.StartLine != 34 {
		t.Errorf("AppName StartLine: got %d, want 34", r.Location.StartLine)
	}
	if r.Location.EndLine != 34 {
		t.Errorf("AppName EndLine: got %d, want 34", r.Location.EndLine)
	}
}

// ---------- Java span tests ----------

func TestJavaClassSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppConfig", "", "java")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppConfig")
	}
	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppConfig class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppConfig class not found in results")
}

func TestJavaMethodSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("getHost", "", "java")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for getHost")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("getHost method: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

func TestJavaEnumSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppStatus", "", "java")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppStatus")
	}
	for _, r := range results {
		if r.Kind == "enum" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppStatus enum: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppStatus enum not found in results")
}

// ---------- Rust span tests ----------

func TestRustStructSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppConfig", "", "rust")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppConfig")
	}
	for _, r := range results {
		if r.Kind == "struct" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppConfig struct: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppConfig struct not found in results")
}

func TestRustMethodSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("validate", "", "rust")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for validate")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("validate method: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

func TestRustEnumSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppStatus", "", "rust")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppStatus")
	}
	for _, r := range results {
		if r.Kind == "enum" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppStatus enum: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppStatus enum not found in results")
}

func TestRustTraitSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppFormatter", "", "rust")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppFormatter")
	}
	for _, r := range results {
		if r.Kind == "trait" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppFormatter trait: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppFormatter trait not found in results")
}

// ---------- Python span tests ----------

func TestPythonClassSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppConfig", "", "python")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppConfig")
	}
	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppConfig class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppConfig class not found in results")
}

func TestPythonFunctionSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("say_hello", "", "python")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for say_hello")
	}
	for _, r := range results {
		if r.Kind == "function" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("say_hello function: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("say_hello function not found in results")
}

func TestPythonMethodSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("validate", "", "python")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for validate")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("validate method: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

// ---------- TypeScript span tests ----------

func TestTypeScriptClassSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppConfig", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppConfig")
	}
	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppConfig class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppConfig class not found in results")
}

func TestTypeScriptFunctionSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("sayHelloApp", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for sayHelloApp")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("sayHelloApp function: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

func TestTypeScriptInterfaceSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppFormatter", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppFormatter")
	}
	for _, r := range results {
		if r.Kind == "interface" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppFormatter interface: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppFormatter interface not found in results")
}

func TestTypeScriptEnumSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppStatus", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppStatus")
	}
	for _, r := range results {
		if r.Kind == "enum" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppStatus enum: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppStatus enum not found in results")
}

func TestTypeScriptArrowFunctionSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("formatNameApp", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for formatNameApp")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("formatNameApp arrow: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

func TestTypeScriptNamespaceSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppUtils", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppUtils")
	}
	for _, r := range results {
		if r.Kind == "module" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppUtils namespace: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppUtils namespace not found in results")
}

// ---------- JavaScript span tests ----------

func TestJavaScriptClassSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("AppConfigJs", "", "javascript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppConfigJs")
	}
	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppConfigJs class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
			return
		}
	}
	t.Error("AppConfigJs class not found in results")
}

func TestJavaScriptFunctionSpan(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	results, err := nav.GoToDefinitionByName("sayHelloJs", "", "javascript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for sayHelloJs")
	}
	r := results[0]
	if r.Location.EndLine <= r.Location.StartLine {
		t.Errorf("sayHelloJs function: want EndLine > StartLine, got %d..%d",
			r.Location.StartLine, r.Location.EndLine)
	}
}

// ---------- Cross-file duplicate name tests ----------

func TestDuplicateNameAcrossFiles(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// "Processor" appears in both Go files: defs_b.go (struct) and
	// "Person" in sample.go. Let's use Processor from defs_b.go.
	results, err := nav.GoToDefinitionByName("Processor", "", "go")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}

	// Should find the struct
	found := false
	for _, r := range results {
		if r.Kind == "struct" {
			found = true
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("Processor struct: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
		}
	}
	if !found {
		t.Error("Processor struct not found")
	}
}

func TestDuplicateNameAcrossFilesRust(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// "AppProcessor" appears in defs_b.rs only as struct
	results, err := nav.GoToDefinitionByName("AppProcessor", "", "rust")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppProcessor")
	}

	for _, r := range results {
		if r.Kind == "struct" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppProcessor struct: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
		}
	}
}

func TestDuplicateNameAcrossFilesTS(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// "AppProcessor" appears in defs_b.ts as class
	results, err := nav.GoToDefinitionByName("AppProcessor", "", "typescript")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppProcessor")
	}

	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppProcessor class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
		}
	}
}

func TestDuplicateNameAcrossFilesPython(t *testing.T) {
	nav, cleanup := setupSpanTest(t)
	defer cleanup()

	// "AppProcessor" appears in defs_b.py as class
	results, err := nav.GoToDefinitionByName("AppProcessor", "", "python")
	if err != nil {
		t.Fatalf("GoToDefinitionByName: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for AppProcessor")
	}

	for _, r := range results {
		if r.Kind == "class" {
			if r.Location.EndLine <= r.Location.StartLine {
				t.Errorf("AppProcessor class: want EndLine > StartLine, got %d..%d",
					r.Location.StartLine, r.Location.EndLine)
			}
		}
	}
}

// ---------- Helper span unit tests ----------

func TestFindBlockEndLine(t *testing.T) {
	lines := []string{
		"type Config struct {",    // line 1
		"    Host string",         // line 2
		"    Port int",            // line 3
		"}",                       // line 4
		"",                        // line 5
		"func main() {",           // line 6
		"    fmt.Println(\"hi\")", // line 7
		"}",                       // line 8
	}

	// Struct: starts at line 1, should end at line 4
	if end := findBlockEndLine(lines, 1); end != 4 {
		t.Errorf("struct block end: got %d, want 4", end)
	}

	// Function: starts at line 6, should end at line 8
	if end := findBlockEndLine(lines, 6); end != 8 {
		t.Errorf("func block end: got %d, want 8", end)
	}

	// Single-line with semicolon
	semicolonLines := []string{
		"const MAX_RETRIES: u32 = 3;", // line 1
	}
	if end := findBlockEndLine(semicolonLines, 1); end != 1 {
		t.Errorf("semicolon end: got %d, want 1", end)
	}
}

func TestPythonBlockEndLine(t *testing.T) {
	lines := []string{
		"class Foo:",             // line 1
		"    def bar(self):",     // line 2
		"        print('hello')", // line 3
		"        print('world')", // line 4
		"",                       // line 5
		"    def baz(self):",     // line 6
		"        pass",           // line 7
		"",                       // line 8
		"def top_level():",       // line 9
		"    return 42",          // line 10
	}

	// Class: starts at line 1, body extends through line 7
	if end := pythonBlockEndLine(lines, 1); end != 7 {
		t.Errorf("class end: got %d, want 7", end)
	}

	// Method bar: starts at line 2, body extends through line 4
	if end := pythonBlockEndLine(lines, 2); end != 4 {
		t.Errorf("method bar end: got %d, want 4", end)
	}

	// Method baz: starts at line 6, body is line 7
	if end := pythonBlockEndLine(lines, 6); end != 7 {
		t.Errorf("method baz end: got %d, want 7", end)
	}

	// Function top_level: starts at line 9, body is line 10
	if end := pythonBlockEndLine(lines, 9); end != 10 {
		t.Errorf("func top_level end: got %d, want 10", end)
	}
}

func TestFindDocStartLine(t *testing.T) {
	goLines := []string{
		"package main",              // line 1
		"",                          // line 2
		"// Config holds config.",   // line 3
		"// It supports many envs.", // line 4
		"type Config struct {",      // line 5
	}
	if start := FindDocStartLine(goLines, 5, LangGo); start != 3 {
		t.Errorf("Go doc start: got %d, want 3", start)
	}

	pyLines := []string{
		"",                     // line 1
		"# A helper decorator", // line 2
		"@my_decorator",        // line 3
		"def my_func():",       // line 4
		"    pass",             // line 5
	}
	if start := FindDocStartLine(pyLines, 4, LangPython); start != 2 {
		t.Errorf("Python doc start: got %d, want 2", start)
	}

	rustLines := []string{
		"",                             // line 1
		"/// AppConfig documentation.", // line 2
		"/// Second line.",             // line 3
		"#[derive(Debug)]",             // line 4
		"pub struct AppConfig {",       // line 5
	}
	if start := FindDocStartLine(rustLines, 5, LangRust); start != 2 {
		t.Errorf("Rust doc start: got %d, want 2", start)
	}

	javaLines := []string{
		"",                        // line 1
		"/**",                     // line 2
		" * Javadoc for MyClass.", // line 3
		" */",                     // line 4
		"@Deprecated",             // line 5
		"public class MyClass {",  // line 6
	}
	if start := FindDocStartLine(javaLines, 6, LangJava); start != 2 {
		t.Errorf("Java doc start: got %d, want 2", start)
	}

	// No doc comment — should return declLine unchanged
	noDocs := []string{
		"",                   // line 1
		"",                   // line 2
		"type Foo struct {}", // line 3
	}
	if start := FindDocStartLine(noDocs, 3, LangGo); start != 3 {
		t.Errorf("no doc start: got %d, want 3", start)
	}
}
