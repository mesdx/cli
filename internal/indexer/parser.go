package indexer

import (
	"fmt"

	"github.com/mesdx/cli/internal/symbols"
	"github.com/mesdx/cli/internal/treesitter"
)

// Parser extracts symbols and references from source code.
type Parser interface {
	// Parse processes the source bytes and returns extracted symbols and references.
	Parse(filename string, src []byte) (*symbols.FileResult, error)
}

// parserRegistry maps languages to their parser implementations.
var parserRegistry = map[Lang]Parser{}

func init() {
	// Use tree-sitter parsers for all languages
	parserRegistry[LangGo] = NewTreeSitterParser("go")
	parserRegistry[LangJava] = NewTreeSitterParser("java")
	parserRegistry[LangRust] = NewTreeSitterParser("rust")
	parserRegistry[LangPython] = NewTreeSitterParser("python")
	parserRegistry[LangTypeScript] = NewTreeSitterParser("typescript")
	parserRegistry[LangJavaScript] = NewTreeSitterParser("javascript")
}

// GetParser returns the parser for the given language, or nil if unsupported.
func GetParser(lang Lang) Parser {
	return parserRegistry[lang]
}

// VerifyParsersAvailable checks that all required parser libraries are available.
// This should be called at startup to fail fast if libraries are missing.
func VerifyParsersAvailable() error {
	required := treesitter.RequiredLanguages()
	if err := treesitter.VerifyLanguages(required); err != nil {
		return fmt.Errorf("parser library verification failed: %w\n\nPlease install parser libraries. Run:\n  mesdx --help\nfor installation instructions", err)
	}
	return nil
}
