package symbols

import "fmt"

// SymbolKind represents the type of a code symbol.
type SymbolKind int

const (
	KindUnknown     SymbolKind = 0
	KindPackage     SymbolKind = 1
	KindClass       SymbolKind = 2
	KindInterface   SymbolKind = 3
	KindStruct      SymbolKind = 4
	KindEnum        SymbolKind = 5
	KindFunction    SymbolKind = 6
	KindMethod      SymbolKind = 7
	KindProperty    SymbolKind = 8
	KindField       SymbolKind = 9
	KindVariable    SymbolKind = 10
	KindConstant    SymbolKind = 11
	KindConstructor SymbolKind = 12
	KindTypeAlias   SymbolKind = 13
	KindTrait       SymbolKind = 14
	KindModule      SymbolKind = 15
)

var kindNames = map[SymbolKind]string{
	KindUnknown:     "unknown",
	KindPackage:     "package",
	KindClass:       "class",
	KindInterface:   "interface",
	KindStruct:      "struct",
	KindEnum:        "enum",
	KindFunction:    "function",
	KindMethod:      "method",
	KindProperty:    "property",
	KindField:       "field",
	KindVariable:    "variable",
	KindConstant:    "constant",
	KindConstructor: "constructor",
	KindTypeAlias:   "type_alias",
	KindTrait:       "trait",
	KindModule:      "module",
}

var nameToKind map[string]SymbolKind

func init() {
	nameToKind = make(map[string]SymbolKind, len(kindNames))
	for k, v := range kindNames {
		nameToKind[v] = k
	}
}

// String returns the human-readable name of the kind.
func (k SymbolKind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("kind(%d)", int(k))
}

// ParseKind converts a string name to a SymbolKind.
// Returns KindUnknown if the name is not recognized.
func ParseKind(name string) SymbolKind {
	if k, ok := nameToKind[name]; ok {
		return k
	}
	return KindUnknown
}

// RefKind describes the type of a reference (usage).
type RefKind int

const (
	RefCall       RefKind = 1
	RefRead       RefKind = 2
	RefWrite      RefKind = 3
	RefImport     RefKind = 4
	RefTypeRef    RefKind = 5
	RefInherit    RefKind = 6
	RefAnnotation RefKind = 7
	RefOther      RefKind = 0
)

// Symbol is the in-memory representation of an extracted definition.
type Symbol struct {
	Name          string
	Kind          SymbolKind
	ContainerName string
	Signature     string
	IsExternal    bool // true when the symbol originates from an external package/module
	StartLine     int  // 1-based
	StartCol      int  // 0-based
	EndLine       int  // 1-based
	EndCol        int  // 0-based
}

var refKindNames = map[RefKind]string{
	RefOther:      "other",
	RefCall:       "call",
	RefRead:       "read",
	RefWrite:      "write",
	RefImport:     "import",
	RefTypeRef:    "type_ref",
	RefInherit:    "inherit",
	RefAnnotation: "annotation",
}

// String returns the human-readable name of the ref kind.
func (k RefKind) String() string {
	if s, ok := refKindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("ref_kind(%d)", int(k))
}

// Ref is the in-memory representation of an extracted usage reference.
type Ref struct {
	Name             string
	Kind             RefKind
	IsExternal       bool   // true when the ref targets an external package/module
	IsBuiltin        bool   // true when the ref targets a language builtin
	Relation         string // semantic relationship: "inherits", "implements", "overrides", "prototype", etc.
	ReceiverType     string // for method calls/field access, the receiver type if known
	TargetType       string // for type refs/inheritance, the target type name
	StartLine        int    // 1-based
	StartCol         int    // 0-based
	EndLine          int    // 1-based
	EndCol           int    // 0-based
	ContextContainer string // enclosing function/class name if available
}

// FileResult holds the parsing output for a single file.
type FileResult struct {
	Symbols []Symbol
	Refs    []Ref
}
