package treesitter

import (
	"testing"

	"github.com/mesdx/cli/internal/symbols"
)

func TestExtractorGo(t *testing.T) {
	// Skip if parsers aren't available
	if err := VerifyLanguages([]string{"go"}); err != nil {
		t.Skip("Go parser not available:", err)
	}

	extractor, err := NewExtractor("go")
	if err != nil {
		t.Fatal(err)
	}
	defer extractor.Close()

	source := []byte(`package main

const MaxRetries = 3

type Person struct {
	Name string
	Age  int
}

func NewPerson(name string) *Person {
	return &Person{Name: name}
}

func (p *Person) Greet() string {
	return "Hello"
}
`)

	result, err := extractor.Extract("test.go", source)
	if err != nil {
		t.Fatal(err)
	}

	// Check we got symbols
	if len(result.Symbols) == 0 {
		t.Error("Expected symbols, got none")
	} else {
		t.Logf("Found %d symbols:", len(result.Symbols))
		for i, sym := range result.Symbols {
			t.Logf("  %d. %s (%s)", i+1, sym.Name, sym.Kind)
		}
	}

	// Check we got refs
	if len(result.Refs) == 0 {
		t.Error("Expected refs, got none")
	} else {
		t.Logf("Found %d refs", len(result.Refs))
	}

	// Check for specific symbols
	foundPerson := false
	foundGreet := false
	for _, sym := range result.Symbols {
		if sym.Name == "Person" && sym.Kind == symbols.KindStruct {
			foundPerson = true
		}
		if sym.Name == "Greet" && sym.Kind == symbols.KindMethod {
			foundGreet = true
		}
	}

	if !foundPerson {
		t.Error("Expected to find Person struct")
	}
	if !foundGreet {
		t.Error("Expected to find Greet method")
	}
}

func TestExtractorPython(t *testing.T) {
	// Skip if parsers aren't available
	if err := VerifyLanguages([]string{"python"}); err != nil {
		t.Skip("Python parser not available:", err)
	}

	extractor, err := NewExtractor("python")
	if err != nil {
		t.Fatal(err)
	}
	defer extractor.Close()

	source := []byte(`class Person:
    def __init__(self, name):
        self.name = name
    
    def greet(self):
        return "Hello"

def say_hello():
    p = Person("World")
    return p.greet()
`)

	result, err := extractor.Extract("test.py", source)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Symbols) == 0 {
		t.Error("Expected symbols, got none")
	}

	// Check for class
	foundPerson := false
	for _, sym := range result.Symbols {
		if sym.Name == "Person" && sym.Kind == symbols.KindClass {
			foundPerson = true
		}
	}

	if !foundPerson {
		t.Error("Expected to find Person class")
	}
}

func TestExtractorTypeScript(t *testing.T) {
	// Skip if parsers aren't available
	if err := VerifyLanguages([]string{"typescript"}); err != nil {
		t.Skip("TypeScript parser not available:", err)
	}

	extractor, err := NewExtractor("typescript")
	if err != nil {
		t.Fatal(err)
	}
	defer extractor.Close()

	source := []byte(`interface Greeter {
    greet(name: string): string;
}

class Person implements Greeter {
    constructor(public name: string) {}
    
    greet(name: string): string {
        return "Hello";
    }
}

function sayHello(): void {
    const p = new Person("World");
}
`)

	result, err := extractor.Extract("test.ts", source)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Symbols) == 0 {
		t.Error("Expected symbols, got none")
	}

	// Check for interface and class
	foundInterface := false
	foundClass := false
	for _, sym := range result.Symbols {
		if sym.Name == "Greeter" && sym.Kind == symbols.KindInterface {
			foundInterface = true
		}
		if sym.Name == "Person" && sym.Kind == symbols.KindClass {
			foundClass = true
		}
	}

	if !foundInterface {
		t.Error("Expected to find Greeter interface")
	}
	if !foundClass {
		t.Error("Expected to find Person class")
	}
}

// ---------------------------------------------------------------------------
// Python type-annotation coverage tests
// ---------------------------------------------------------------------------

// TestPythonQuotedForwardRef verifies that identifiers inside Python quoted
// type annotations (forward references) are emitted as annotation refs.
func TestPythonQuotedForwardRef(t *testing.T) {
	if err := VerifyLanguages([]string{"python"}); err != nil {
		t.Skip("Python parser not available:", err)
	}

	extractor, err := NewExtractor("python")
	if err != nil {
		t.Fatal(err)
	}
	defer extractor.Close()

	source := []byte(`from typing import List, Optional

class DataModel:
    pass

# Pattern 1: plain quoted return type
def get_one() -> "DataModel":
    return DataModel()

# Pattern 2: quoted element inside generic
def get_list() -> List["DataModel"]:
    return []

# Pattern 3: quoted tuple element
def get_pair() -> tuple["DataModel", bool]:
    return (DataModel(), True)

# Pattern 4: parameter is a quoted type
def accept(m: "DataModel") -> bool:
    return True

# Pattern 5: entire complex type is quoted
def get_nested() -> "Optional[DataModel]":
    return DataModel()

# Negative: string NOT in annotation position — must NOT produce a ref
description = "DataModel is a class"
`)

	result, err := extractor.Extract("test.py", source)
	if err != nil {
		t.Fatal(err)
	}

	// Verify DataModel appears as a ref (from the quoted annotations).
	foundAnnotationRef := false
	for _, r := range result.Refs {
		if r.Name == "DataModel" && r.Kind == RefAnnotation {
			foundAnnotationRef = true
		}
	}
	if !foundAnnotationRef {
		t.Error("expected at least one RefAnnotation ref for DataModel from quoted type annotations")
	}

	// Verify the raw negative-control string does NOT produce a DataModel annotation ref.
	// The only source of DataModel annotation refs should be the type annotations.
	annotLines := map[int]bool{}
	for _, r := range result.Refs {
		if r.Name == "DataModel" && r.Kind == RefAnnotation {
			annotLines[r.StartLine] = true
			t.Logf("  annotation ref: DataModel at line %d col %d", r.StartLine, r.StartCol)
		}
	}
	// The plain string assignment is on line 27; that line must not produce an annotation ref.
	if annotLines[27] {
		t.Error("line 27 (plain string assignment) should not produce an annotation ref for DataModel")
	}
}

// TestPythonQuotedAnnotation_Positions verifies column positions are correct
// for identifiers found inside quoted type strings.
func TestPythonQuotedAnnotation_Positions(t *testing.T) {
	if err := VerifyLanguages([]string{"python"}); err != nil {
		t.Skip("Python parser not available:", err)
	}

	extractor, err := NewExtractor("python")
	if err != nil {
		t.Fatal(err)
	}
	defer extractor.Close()

	// Line 1 (1-indexed): def foo() -> "MyClass": pass
	// "MyClass" starts at col 14 (0-indexed): `def foo() -> "` = 14 chars, then M at 14.
	source := []byte(`def foo() -> "MyClass": pass
`)

	result, err := extractor.Extract("pos_test.py", source)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range result.Refs {
		if r.Name == "MyClass" && r.Kind == RefAnnotation {
			found = true
			// "def foo() -> \"" = 14 chars, so MyClass col should be 14
			if r.StartCol != 14 {
				t.Errorf("expected MyClass at col 14, got col %d", r.StartCol)
			}
			if r.StartLine != 1 {
				t.Errorf("expected MyClass at line 1, got line %d", r.StartLine)
			}
		}
	}
	if !found {
		t.Error("expected MyClass annotation ref; not found")
	}
}

// RefAnnotation is a helper alias so extractor_test can reference it without
// importing the symbols package directly (it's already in scope via extractor.go).
const RefAnnotation = 7 // symbols.RefAnnotation
