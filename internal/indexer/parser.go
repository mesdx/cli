package indexer

import "github.com/codeintelx/cli/internal/symbols"

// Parser extracts symbols and references from source code.
type Parser interface {
	// Parse processes the source bytes and returns extracted symbols and references.
	Parse(filename string, src []byte) (*symbols.FileResult, error)
}

// parserRegistry maps languages to their parser implementations.
var parserRegistry = map[Lang]Parser{}

func init() {
	parserRegistry[LangGo] = &GoParser{}
	parserRegistry[LangJava] = &JavaParser{}
	parserRegistry[LangRust] = &RustParser{}
	parserRegistry[LangPython] = &PythonParser{}
	parserRegistry[LangTypeScript] = &TypeScriptParser{}
	parserRegistry[LangJavaScript] = &TypeScriptParser{} // JS uses the same regex patterns
}

// GetParser returns the parser for the given language, or nil if unsupported.
func GetParser(lang Lang) Parser {
	return parserRegistry[lang]
}
