package indexer

import (
	"github.com/mesdx/cli/internal/symbols"
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
