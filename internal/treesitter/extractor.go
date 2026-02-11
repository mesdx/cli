package treesitter

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/mesdx/cli/internal/symbols"
)

//go:embed queries/go.scm
var goQuery string

//go:embed queries/java.scm
var javaQuery string

//go:embed queries/rust.scm
var rustQuery string

//go:embed queries/python.scm
var pythonQuery string

//go:embed queries/typescript.scm
var typescriptQuery string

//go:embed queries/javascript.scm
var javascriptQuery string

// Extractor extracts symbols and references from parsed trees using queries.
type Extractor struct {
	lang      *Language
	query     *Query
	langName  string
}

// NewExtractor creates a new extractor for the given language.
func NewExtractor(langName string) (*Extractor, error) {
	lang, err := LoadLanguage(langName)
	if err != nil {
		return nil, fmt.Errorf("load language %s: %w", langName, err)
	}

	// Select query source
	var querySource string
	switch langName {
	case "go":
		querySource = goQuery
	case "java":
		querySource = javaQuery
	case "rust":
		querySource = rustQuery
	case "python":
		querySource = pythonQuery
	case "typescript":
		querySource = typescriptQuery
	case "javascript":
		querySource = javascriptQuery
	default:
		return nil, fmt.Errorf("no query defined for language %s", langName)
	}

	query, err := NewQuery(lang, querySource)
	if err != nil {
		return nil, fmt.Errorf("parse query for %s: %w", langName, err)
	}

	return &Extractor{
		lang:     lang,
		query:    query,
		langName: langName,
	}, nil
}

// Extract parses source code and extracts symbols and references.
func (e *Extractor) Extract(filename string, source []byte) (*symbols.FileResult, error) {
	// Create parser
	parser := NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(e.lang); err != nil {
		return nil, fmt.Errorf("set language: %w", err)
	}

	// Parse the source
	tree := parser.ParseString(nil, source)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse source")
	}
	defer tree.Close()

	// Execute query
	cursor := NewQueryCursor()
	defer cursor.Close()

	rootNode := tree.RootNode()
	cursor.ExecWithText(e.query, rootNode, source)

	result := &symbols.FileResult{
		Symbols: []symbols.Symbol{},
		Refs:    []symbols.Ref{},
	}

	// Build a map of capture names to IDs
	captureNames := make(map[uint32]string)
	for i := uint32(0); i < e.query.CaptureCount(); i++ {
		captureNames[i] = e.query.CaptureNameForID(i)
	}

	// Track seen symbols to avoid duplicates
	seenSymbols := make(map[string]bool)
	seenRefs := make(map[string]bool)

	// Container context for methods
	containerContext := make(map[string]string) // node pointer -> container name

	// Process matches
	for {
		match := cursor.NextMatch()
		if match == nil {
			break
		}

		// Group captures by name
		capturesByName := make(map[string][]QueryCapture)
		for _, cap := range match.Captures {
			name := captureNames[cap.Index]
			capturesByName[name] = append(capturesByName[name], cap)
		}

		// Extract container name if present
		containerName := ""
		if containers, ok := capturesByName["container.name"]; ok && len(containers) > 0 {
			containerName = containers[0].Node.Content(source)
		}

		// Process definition captures
		for capName, caps := range capturesByName {
			if !strings.HasPrefix(capName, "def.") {
				continue
			}

			for _, cap := range caps {
				node := cap.Node
				if node.IsNull() || !node.IsNamed() {
					continue
				}

				name := node.Content(source)
				if name == "" || isCommonKeyword(name) {
					continue
				}

				// Determine symbol kind
				kind := mapCaptureToKind(capName)
				if kind == symbols.KindUnknown {
					continue
				}

				// Python-specific: Detect constants by naming convention (UPPER_CASE)
				if e.langName == "python" && kind == symbols.KindVariable && isUpperCaseIdentifier(name) {
					kind = symbols.KindConstant
				}

				// Get span
				startPoint := node.StartPoint()
				endPoint := node.EndPoint()

				// For definitions, try to expand to the full declaration node
				parent := node.Parent()
				if !parent.IsNull() {
					endPoint = parent.EndPoint()
				}

				// Build unique key to avoid duplicates
				key := fmt.Sprintf("%s:%d:%d:%s", name, startPoint.Row, startPoint.Column, kind)
				if seenSymbols[key] {
					continue
				}
				seenSymbols[key] = true

				// Store container context for this node
				if containerName != "" {
					nodeKey := fmt.Sprintf("%d", node.StartByte())
					containerContext[nodeKey] = containerName
				}

				// Determine container (for methods)
				container := containerName
				if container == "" {
					// Try to find from parent context
					parentKey := fmt.Sprintf("%d", parent.StartByte())
					container = containerContext[parentKey]
				}

				sym := symbols.Symbol{
					Name:          name,
					Kind:          kind,
					ContainerName: container,
					StartLine:     int(startPoint.Row) + 1,
					StartCol:      int(startPoint.Column),
					EndLine:       int(endPoint.Row) + 1,
					EndCol:        int(startPoint.Column) + len(name),
				}

				result.Symbols = append(result.Symbols, sym)
			}
		}

		// Process reference captures
		for capName, caps := range capturesByName {
			if !strings.HasPrefix(capName, "ref.") {
				continue
			}

			for _, cap := range caps {
				node := cap.Node
				if node.IsNull() {
					continue
				}

				name := node.Content(source)
				if name == "" || len(name) <= 1 || isCommonKeyword(name) {
					continue
				}

				// Skip if this is a definition (avoid double-counting)
				startPoint := node.StartPoint()
				defKey := fmt.Sprintf("%s:%d:%d", name, startPoint.Row, startPoint.Column)
				if seenSymbols[defKey] {
					continue
				}

				// Build unique key for refs
				endPoint := node.EndPoint()
				refKey := fmt.Sprintf("%s:%d:%d", name, startPoint.Row, startPoint.Column)
				if seenRefs[refKey] {
					continue
				}
				seenRefs[refKey] = true

				refKind := mapCaptureToRefKind(capName)

				ref := symbols.Ref{
					Name:      name,
					Kind:      refKind,
					StartLine: int(startPoint.Row) + 1,
					StartCol:  int(startPoint.Column),
					EndLine:   int(endPoint.Row) + 1,
					EndCol:    int(endPoint.Column),
				}

				result.Refs = append(result.Refs, ref)
			}
		}
	}

	return result, nil
}

// Close closes the extractor's resources.
func (e *Extractor) Close() {
	if e.query != nil {
		e.query.Close()
	}
}

// mapCaptureToKind maps a capture name to a symbol kind.
func mapCaptureToKind(capName string) symbols.SymbolKind {
	switch capName {
	case "def.package":
		return symbols.KindPackage
	case "def.class":
		return symbols.KindClass
	case "def.interface":
		return symbols.KindInterface
	case "def.struct":
		return symbols.KindStruct
	case "def.enum":
		return symbols.KindEnum
	case "def.function":
		return symbols.KindFunction
	case "def.method":
		return symbols.KindMethod
	case "def.property":
		return symbols.KindProperty
	case "def.field":
		return symbols.KindField
	case "def.var":
		return symbols.KindVariable
	case "def.const":
		return symbols.KindConstant
	case "def.constructor":
		return symbols.KindConstructor
	case "def.typealias", "def.type":
		return symbols.KindTypeAlias
	case "def.trait":
		return symbols.KindTrait
	case "def.module":
		return symbols.KindModule
	case "def.parameter":
		return symbols.KindVariable
	default:
		return symbols.KindUnknown
	}
}

// mapCaptureToRefKind maps a capture name to a reference kind.
func mapCaptureToRefKind(capName string) symbols.RefKind {
	switch capName {
	case "ref.import":
		return symbols.RefImport
	default:
		return symbols.RefOther
	}
}

// isCommonKeyword filters out common language keywords.
func isCommonKeyword(s string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "break": true, "continue": true,
		"return": true, "try": true, "catch": true, "finally": true,
		"throw": true, "this": true, "super": true,
		"class": true, "extends": true, "implements": true, "import": true,
		"export": true, "from": true, "default": true, "void": true,
		"null": true, "undefined": true, "true": true, "false": true,
		"typeof": true, "instanceof": true, "in": true, "of": true,
		"as": true, "is": true, "let": true, "const": true, "var": true,
		"function": true, "async": true, "await": true, "yield": true,
		"public": true, "private": true, "protected": true, "static": true,
		"final": true, "abstract": true, "interface": true, "enum": true,
		"struct": true, "trait": true, "impl": true, "type": true,
		"fn": true, "pub": true, "mod": true, "use": true, "self": true,
		"Self": true, "crate": true, "match": true, "loop": true,
		"mut": true, "ref": true, "unsafe": true, "extern": true,
		"def": true, "lambda": true, "pass": true, "raise": true,
		"except": true, "with": true, "assert": true, "del": true,
		"not": true, "and": true, "or": true, "None": true, "True": true,
		"False": true, "package": true, "goto": true,
	}
	return keywords[s]
}

// isUpperCaseIdentifier checks if an identifier follows UPPER_CASE naming convention.
// Used to detect Python constants (e.g., MAX_RETRIES).
func isUpperCaseIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch >= 'a' && ch <= 'z' {
			return false
		}
	}
	// At least one uppercase letter or underscore
	hasUpper := false
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			hasUpper = true
			break
		}
	}
	return hasUpper
}
