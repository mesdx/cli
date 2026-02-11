package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mesdx/cli/internal/symbols"
)

// ---------- Go tricky fixture ----------

func TestGoTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "go", "tricky.go"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("go")
	result, err := parser.Parse("tricky.go", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "GlobalTimeout", symbols.KindConstant)
	expectSymbol(t, result, "ShadowExample", symbols.KindFunction)
	expectSymbol(t, result, "Animal", symbols.KindStruct)
	expectSymbol(t, result, "Vehicle", symbols.KindStruct)
	expectSymbol(t, result, "Outer", symbols.KindStruct)
	expectSymbol(t, result, "Reader", symbols.KindInterface)
	expectSymbol(t, result, "ReadCloser", symbols.KindInterface)
	expectSymbol(t, result, "UseBuiltins", symbols.KindFunction)
	expectSymbol(t, result, "UseExternalStdlib", symbols.KindFunction)
	expectSymbol(t, result, "MyString", symbols.KindTypeAlias)
	expectSymbol(t, result, "Color", symbols.KindTypeAlias)
	expectSymbol(t, result, "Red", symbols.KindConstant)
	expectSymbol(t, result, "Green", symbols.KindConstant)
	expectSymbol(t, result, "Blue", symbols.KindConstant)
	expectSymbol(t, result, "TransformFunc", symbols.KindTypeAlias)

	// Both Animal and Vehicle should have a String method
	stringMethods := 0
	for _, s := range result.Symbols {
		if s.Name == "String" && s.Kind == symbols.KindMethod {
			stringMethods++
		}
	}
	if stringMethods < 2 {
		t.Errorf("expected at least 2 String methods (Animal + Vehicle), got %d", stringMethods)
	}

	// --- References ---
	expectRef(t, result, "fmt")
	expectRef(t, result, "os")
	expectRef(t, result, "strings")

	// --- Builtins should be present as refs ---
	expectRef(t, result, "make")
	expectRef(t, result, "append")
	expectRef(t, result, "len")
	expectRef(t, result, "cap")
	expectRef(t, result, "delete")
	expectRef(t, result, "close")
	expectRef(t, result, "println")

	// --- Builtin refs should be marked as builtin ---
	expectRefBuiltin(t, result, "make", true)
	expectRefBuiltin(t, result, "append", true)
	expectRefBuiltin(t, result, "len", true)

	// --- Import refs should be marked as external ---
	expectRefExternal(t, result, "os", true)
	expectRefExternal(t, result, "strings", true)

	// --- Def should NOT be double-counted as ref ---
	expectNoDuplicateDefRef(t, result, "ShadowExample")
	expectNoDuplicateDefRef(t, result, "Animal")
}

// ---------- Java tricky fixture ----------

func TestJavaTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "java", "Tricky.java"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("java")
	result, err := parser.Parse("Tricky.java", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "BaseEntity", symbols.KindClass)
	expectSymbol(t, result, "User", symbols.KindClass)
	expectSymbol(t, result, "LegacyProcessor", symbols.KindClass)
	expectSymbol(t, result, "Repository", symbols.KindClass)
	expectSymbol(t, result, "AppConstants", symbols.KindClass)
	expectSymbol(t, result, "Priority", symbols.KindEnum)
	expectSymbol(t, result, "BuiltinUsage", symbols.KindClass)

	// Methods
	expectSymbol(t, result, "getId", symbols.KindMethod)
	expectSymbol(t, result, "validate", symbols.KindMethod)
	expectSymbol(t, result, "save", symbols.KindMethod)
	expectSymbol(t, result, "findById", symbols.KindMethod)
	expectSymbol(t, result, "getAppInfo", symbols.KindMethod)
	expectSymbol(t, result, "getValue", symbols.KindMethod)

	// Constructors
	expectSymbol(t, result, "User", symbols.KindConstructor)

	// Inner class
	expectSymbol(t, result, "Permissions", symbols.KindClass)

	// Method overloading: "format" should appear more than once
	formatMethods := 0
	for _, s := range result.Symbols {
		if s.Name == "format" && s.Kind == symbols.KindMethod {
			formatMethods++
		}
	}
	if formatMethods < 2 {
		t.Errorf("expected at least 2 format methods (overloaded), got %d", formatMethods)
	}

	// --- References ---
	// Import refs should be external
	expectRef(t, result, "HashMap")
	expectRef(t, result, "Optional")
	expectRefExternal(t, result, "HashMap", true)

	// Builtins (java.lang) should be marked builtin
	expectRefBuiltin(t, result, "String", true)
	expectRefBuiltin(t, result, "Integer", true)
	expectRefBuiltin(t, result, "System", true)
	expectRefBuiltin(t, result, "Object", true)
	expectRefBuiltin(t, result, "Math", true)
	expectRefBuiltin(t, result, "Exception", true)
	expectRefBuiltin(t, result, "RuntimeException", true)

	// --- Dedup ---
	expectNoDuplicateDefRef(t, result, "BaseEntity")
	expectNoDuplicateDefRef(t, result, "User")
}

// ---------- Python tricky fixture ----------

func TestPythonTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "python", "tricky.py"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("python")
	result, err := parser.Parse("tricky.py", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "MAX_RETRIES", symbols.KindConstant)
	expectSymbol(t, result, "shadow_example", symbols.KindFunction)
	expectSymbol(t, result, "Serializable", symbols.KindClass)
	expectSymbol(t, result, "Printable", symbols.KindClass)
	expectSymbol(t, result, "Document", symbols.KindClass)
	expectSymbol(t, result, "Config", symbols.KindClass)
	expectSymbol(t, result, "Point", symbols.KindClass)
	expectSymbol(t, result, "Container", symbols.KindClass)
	expectSymbol(t, result, "use_builtins", symbols.KindFunction)
	expectSymbol(t, result, "use_external_stdlib", symbols.KindFunction)

	// Properties
	expectSymbol(t, result, "host", symbols.KindProperty)

	// Methods
	expectSymbol(t, result, "serialize", symbols.KindMethod)
	expectSymbol(t, result, "display", symbols.KindMethod)
	expectSymbol(t, result, "distance_to", symbols.KindMethod)

	// Dunder methods
	expectSymbol(t, result, "__len__", symbols.KindMethod)
	expectSymbol(t, result, "__getitem__", symbols.KindMethod)
	expectSymbol(t, result, "__contains__", symbols.KindMethod)

	// --- References ---
	// Import refs should be external
	expectRef(t, result, "os")
	expectRef(t, result, "json")
	expectRef(t, result, "defaultdict")
	expectRefExternal(t, result, "os", true)
	expectRefExternal(t, result, "json", true)

	// Builtins
	expectRefBuiltin(t, result, "len", true)
	expectRefBuiltin(t, result, "range", true)
	expectRefBuiltin(t, result, "list", true)
	expectRefBuiltin(t, result, "str", true)
	expectRefBuiltin(t, result, "int", true)
	expectRefBuiltin(t, result, "float", true)
	expectRefBuiltin(t, result, "dict", true)
	expectRefBuiltin(t, result, "set", true)
	expectRefBuiltin(t, result, "tuple", true)
	expectRefBuiltin(t, result, "isinstance", true)
	expectRefBuiltin(t, result, "print", true)
	expectRefBuiltin(t, result, "type", true)
	expectRefBuiltin(t, result, "sorted", true)
	expectRefBuiltin(t, result, "ValueError", true)

	// --- Dedup ---
	expectNoDuplicateDefRef(t, result, "Document")
	expectNoDuplicateDefRef(t, result, "Config")
}

// ---------- TypeScript tricky fixture ----------

func TestTypeScriptTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "ts", "tricky.ts"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("typescript")
	result, err := parser.Parse("tricky.ts", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "TIMEOUT", symbols.KindVariable)
	expectSymbol(t, result, "shadowExample", symbols.KindFunction)
	expectSymbol(t, result, "Serializable", symbols.KindInterface)
	expectSymbol(t, result, "Printable", symbols.KindInterface)
	expectSymbol(t, result, "BaseEntity", symbols.KindClass)
	expectSymbol(t, result, "UserEntity", symbols.KindClass)
	expectSymbol(t, result, "InMemoryRepository", symbols.KindClass)
	expectSymbol(t, result, "HttpStatus", symbols.KindEnum)
	expectSymbol(t, result, "Direction", symbols.KindEnum)
	expectSymbol(t, result, "App", symbols.KindClass)

	// Type aliases
	expectSymbol(t, result, "StringOrNumber", symbols.KindTypeAlias)
	expectSymbol(t, result, "UserDTO", symbols.KindTypeAlias)
	expectSymbol(t, result, "ReadonlyUser", symbols.KindTypeAlias)
	expectSymbol(t, result, "NamedAged", symbols.KindTypeAlias)

	// Interfaces
	expectSymbol(t, result, "Repository", symbols.KindInterface)

	// --- References ---
	// Import refs should be external
	expectRef(t, result, "EventEmitter")
	expectRefExternal(t, result, "EventEmitter", true)

	// Builtins (JS/TS globals)
	expectRefBuiltin(t, result, "Array", true)
	expectRefBuiltin(t, result, "Map", true)
	expectRefBuiltin(t, result, "Set", true)
	expectRefBuiltin(t, result, "Promise", true)
	expectRefBuiltin(t, result, "Date", true)
	expectRefBuiltin(t, result, "RegExp", true)
	expectRefBuiltin(t, result, "Error", true)
	expectRefBuiltin(t, result, "JSON", true)
	expectRefBuiltin(t, result, "Math", true)
	expectRefBuiltin(t, result, "console", true)
	expectRefBuiltin(t, result, "setTimeout", true)
	expectRefBuiltin(t, result, "clearTimeout", true)

	// --- Dedup ---
	expectNoDuplicateDefRef(t, result, "UserEntity")
	expectNoDuplicateDefRef(t, result, "BaseEntity")
}

// ---------- JavaScript tricky fixture ----------

func TestJavaScriptTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "js", "tricky.js"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("javascript")
	result, err := parser.Parse("tricky.js", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "TIMEOUT", symbols.KindVariable)
	expectSymbol(t, result, "shadowExample", symbols.KindFunction)
	expectSymbol(t, result, "Animal", symbols.KindFunction) // function-based constructor
	expectSymbol(t, result, "BaseEntity", symbols.KindClass)
	expectSymbol(t, result, "User", symbols.KindClass)
	expectSymbol(t, result, "processConfig", symbols.KindFunction)
	expectSymbol(t, result, "useBuiltins", symbols.KindFunction)
	expectSymbol(t, result, "DynamicProps", symbols.KindClass)

	// Methods
	expectSymbol(t, result, "validate", symbols.KindMethod)
	expectSymbol(t, result, "constructor", symbols.KindConstructor)

	// --- References ---
	// Builtins
	expectRefBuiltin(t, result, "Array", true)
	expectRefBuiltin(t, result, "Map", true)
	expectRefBuiltin(t, result, "Set", true)
	expectRefBuiltin(t, result, "Promise", true)
	expectRefBuiltin(t, result, "Date", true)
	expectRefBuiltin(t, result, "JSON", true)
	expectRefBuiltin(t, result, "Math", true)
	expectRefBuiltin(t, result, "console", true)
	expectRefBuiltin(t, result, "setTimeout", true)

	// --- Dedup ---
	expectNoDuplicateDefRef(t, result, "BaseEntity")
	expectNoDuplicateDefRef(t, result, "User")
}

// ---------- Rust tricky fixture ----------

func TestRustTrickyFixture(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "rust", "tricky.rs"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("rust")
	result, err := parser.Parse("tricky.rs", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// --- Definitions ---
	expectSymbol(t, result, "MAX_RETRIES", symbols.KindConstant)
	expectSymbol(t, result, "shadow_example", symbols.KindFunction)
	expectSymbol(t, result, "Serializable", symbols.KindTrait)
	expectSymbol(t, result, "Printable", symbols.KindTrait)
	expectSymbol(t, result, "Document", symbols.KindStruct)
	expectSymbol(t, result, "Shape", symbols.KindEnum)
	expectSymbol(t, result, "Repository", symbols.KindStruct)
	expectSymbol(t, result, "Borrowed", symbols.KindStruct)
	expectSymbol(t, result, "APP_NAME", symbols.KindConstant)
	expectSymbol(t, result, "VERSION", symbols.KindConstant)

	// Type aliases
	expectSymbol(t, result, "UserId", symbols.KindTypeAlias)
	expectSymbol(t, result, "UserMap", symbols.KindTypeAlias)
	expectSymbol(t, result, "Transformer", symbols.KindTypeAlias)

	// Methods
	expectSymbol(t, result, "area", symbols.KindMethod)
	expectSymbol(t, result, "serialize", symbols.KindMethod)
	expectSymbol(t, result, "save", symbols.KindMethod)
	expectSymbol(t, result, "find", symbols.KindMethod)
	expectSymbol(t, result, "double", symbols.KindFunction)
	expectSymbol(t, result, "apply", symbols.KindFunction)
	expectSymbol(t, result, "main", symbols.KindFunction)

	// --- References ---
	// Import refs should be external
	expectRef(t, result, "HashMap")
	expectRef(t, result, "fmt")
	expectRefExternal(t, result, "HashMap", true)

	// Prelude / builtins
	expectRefBuiltin(t, result, "Option", true)
	expectRefBuiltin(t, result, "Result", true)
	expectRefBuiltin(t, result, "Vec", true)
	expectRefBuiltin(t, result, "String", true)
	expectRefBuiltin(t, result, "Box", true)
	expectRefBuiltin(t, result, "Some", true)
	expectRefBuiltin(t, result, "Ok", true)

	// --- Dedup ---
	expectNoDuplicateDefRef(t, result, "Document")
	expectNoDuplicateDefRef(t, result, "Shape")
}

// ---------- helper assertions ----------

// expectRefBuiltin asserts that at least one ref with the given name has IsBuiltin == expected.
func expectRefBuiltin(t *testing.T, result *symbols.FileResult, name string, expected bool) {
	t.Helper()
	for _, r := range result.Refs {
		if r.Name == name && r.IsBuiltin == expected {
			return
		}
	}
	if expected {
		t.Errorf("ref %q: expected IsBuiltin=true, but none found (or all IsBuiltin=false)", name)
	} else {
		t.Errorf("ref %q: expected IsBuiltin=false, but all were IsBuiltin=true", name)
	}
}

// expectRefExternal asserts that at least one ref with the given name has IsExternal == expected.
func expectRefExternal(t *testing.T, result *symbols.FileResult, name string, expected bool) {
	t.Helper()
	for _, r := range result.Refs {
		if r.Name == name && r.IsExternal == expected {
			return
		}
	}
	if expected {
		t.Errorf("ref %q: expected IsExternal=true, but none found (or all IsExternal=false)", name)
	} else {
		t.Errorf("ref %q: expected IsExternal=false, but all were IsExternal=true", name)
	}
}

// expectNoDuplicateDefRef verifies that if a name is a definition at position (line, col),
// it does NOT also appear as a ref at the same (line, col).
func expectNoDuplicateDefRef(t *testing.T, result *symbols.FileResult, name string) {
	t.Helper()
	defPositions := map[string]bool{}
	for _, s := range result.Symbols {
		if s.Name == name {
			key := posKey(s.StartLine, s.StartCol)
			defPositions[key] = true
		}
	}
	for _, r := range result.Refs {
		if r.Name == name {
			key := posKey(r.StartLine, r.StartCol)
			if defPositions[key] {
				t.Errorf("name %q at %s is both a definition and a reference (double-counted)", name, key)
			}
		}
	}
}

func posKey(line, col int) string {
	return string(rune('0'+line/100)) + string(rune('0'+(line%100)/10)) + string(rune('0'+line%10)) +
		":" +
		string(rune('0'+col/100)) + string(rune('0'+(col%100)/10)) + string(rune('0'+col%10))
}
