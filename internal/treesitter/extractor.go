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

// pendingCapture holds a captured node awaiting processing in the two-pass extraction.
type pendingCapture struct {
	capName       string
	node          Node
	containerName string
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

	// ---------- Pass 1: collect all captures ----------
	var defCaptures []pendingCapture
	var refCaptures []pendingCapture

	for {
		match := cursor.NextMatch()
		if match == nil {
			break
		}

		// Group captures by name for this match
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

		for capName, caps := range capturesByName {
			for _, cap := range caps {
				if strings.HasPrefix(capName, "def.") {
					defCaptures = append(defCaptures, pendingCapture{
						capName:       capName,
						node:          cap.Node,
						containerName: containerName,
					})
				} else if strings.HasPrefix(capName, "ref.") {
					refCaptures = append(refCaptures, pendingCapture{
						capName:       capName,
						node:          cap.Node,
						containerName: containerName,
					})
				}
			}
		}
	}

	// ---------- Pass 2: process definitions ----------
	seenSymbols := make(map[string]bool)
	defPositions := make(map[string]bool)
	containerContext := make(map[string]string)

	for _, dc := range defCaptures {
		node := dc.node
		if node.IsNull() || !node.IsNamed() {
			continue
		}

		name := node.Content(source)
		if name == "" || isCommonKeyword(name) {
			continue
		}

		kind := mapCaptureToKind(dc.capName)
		if kind == symbols.KindUnknown {
			continue
		}

		// Python-specific: Detect constants by naming convention (UPPER_CASE)
		if e.langName == "python" && kind == symbols.KindVariable && isUpperCaseIdentifier(name) {
			kind = symbols.KindConstant
		}

		startPoint := node.StartPoint()
		endPoint := node.EndPoint()

		parent := node.Parent()
		if !parent.IsNull() {
			endPoint = parent.EndPoint()
		}

		symKey := fmt.Sprintf("%s:%d:%d:%s", name, startPoint.Row, startPoint.Column, kind)
		if seenSymbols[symKey] {
			continue
		}
		seenSymbols[symKey] = true

		posKey := fmt.Sprintf("%s:%d:%d", name, startPoint.Row, startPoint.Column)
		defPositions[posKey] = true

		if dc.containerName != "" {
			nodeKey := fmt.Sprintf("%d", node.StartByte())
			containerContext[nodeKey] = dc.containerName
		}

		container := dc.containerName
		if container == "" {
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

	// ---------- Pass 3: process references ----------
	// We use semantic deduplication: same position, keep higher priority capture.
	seenRefs := make(map[string]symbols.Ref) // key = "name:row:col"
	importNames := make(map[string]bool)

	for _, rc := range refCaptures {
		node := rc.node
		if node.IsNull() {
			continue
		}

		name := node.Content(source)
		if name == "" || len(name) <= 1 || isCommonKeyword(name) {
			continue
		}

		startPoint := node.StartPoint()

		// Skip if this position is a definition (avoid double-counting)
		posKey := fmt.Sprintf("%s:%d:%d", name, startPoint.Row, startPoint.Column)
		if defPositions[posKey] {
			continue
		}

		refKey := fmt.Sprintf("%s:%d:%d", name, startPoint.Row, startPoint.Column)
		refKind := mapCaptureToRefKind(rc.capName)
		relation := mapCaptureToRelation(rc.capName)

		// Semantic deduplication: if we've seen this ref position, keep the higher priority one
		if existing, seen := seenRefs[refKey]; seen {
			existingPriority := refSemanticPriority(existing.Kind, existing.Relation)
			newPriority := refSemanticPriority(refKind, relation)
			if newPriority <= existingPriority {
				continue // Keep existing higher-priority ref
			}
			// Otherwise, replace with new higher-priority ref
		}

		// Go-specific: import paths are string literals with quotes;
		// strip them and extract the package name (last path segment).
		if refKind == symbols.RefImport && e.langName == "go" {
			name = strings.Trim(name, "\"'`")
			if idx := strings.LastIndex(name, "/"); idx >= 0 {
				name = name[idx+1:]
			}
			if name == "" || len(name) <= 1 {
				continue
			}
		}

		// Track import names
		if refKind == symbols.RefImport {
			importNames[name] = true
		}

		// Classify builtin
		isBuiltin := IsBuiltin(e.langName, name)

		// Classify external: import refs are always external;
		// non-builtin refs that match a previously seen import name are external too.
		isExternal := refKind == symbols.RefImport
		if !isExternal && !isBuiltin && importNames[name] {
			isExternal = true
		}

		endPoint := node.EndPoint()

		ref := symbols.Ref{
			Name:             name,
			Kind:             refKind,
			IsExternal:       isExternal,
			IsBuiltin:        isBuiltin,
			Relation:         relation,
			ReceiverType:     "", // TODO: extract from parent node context if needed
			TargetType:       "", // TODO: for type refs, could extract target type
			StartLine:        int(startPoint.Row) + 1,
			StartCol:         int(startPoint.Column),
			EndLine:          int(endPoint.Row) + 1,
			EndCol:           int(endPoint.Column),
			ContextContainer: rc.containerName,
		}

		seenRefs[refKey] = ref
	}

	// Collect all refs (deduplicated)
	for _, ref := range seenRefs {
		result.Refs = append(result.Refs, ref)
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

// mapCaptureToRelation maps a capture name to a semantic relation string.
func mapCaptureToRelation(capName string) string {
	switch capName {
	case "ref.inherit":
		return "inherits"
	case "ref.implements":
		return "implements"
	case "ref.annotation":
		return "annotation"
	case "ref.prototype":
		return "prototype"
	default:
		return ""
	}
}

// refSemanticPriority returns priority for deduplication (higher = more specific).
// When multiple captures target the same position, keep the most semantic one.
func refSemanticPriority(kind symbols.RefKind, relation string) int {
	// Most semantic: inheritance, interface implementation, override
	if kind == symbols.RefInherit {
		if relation == "implements" {
			return 100
		}
		return 90 // inherits
	}
	if kind == symbols.RefAnnotation {
		return 80
	}
	if kind == symbols.RefCall {
		return 70
	}
	if kind == symbols.RefWrite {
		return 60
	}
	if kind == symbols.RefImport {
		return 50
	}
	if kind == symbols.RefTypeRef {
		return 40
	}
	if kind == symbols.RefRead {
		return 30
	}
	// Fallback/generic
	return 10
}

// mapCaptureToRefKind maps a capture name to a reference kind.
func mapCaptureToRefKind(capName string) symbols.RefKind {
	switch capName {
	case "ref.import":
		return symbols.RefImport
	case "ref.type":
		return symbols.RefTypeRef
	case "ref.inherit", "ref.implements":
		return symbols.RefInherit
	case "ref.attribute", "ref.annotation":
		return symbols.RefAnnotation
	case "ref.call":
		return symbols.RefCall
	case "ref.write":
		return symbols.RefWrite
	case "ref.field", "ref.property", "ref.prototype":
		return symbols.RefRead
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
		"struct": true, "trait": true, "impl": true,
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
