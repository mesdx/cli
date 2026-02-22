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
