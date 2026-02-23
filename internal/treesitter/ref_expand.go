package treesitter

import (
	"regexp"

	"github.com/mesdx/cli/internal/symbols"
)

// expandedRef is a single derived reference extracted from within a captured
// node's text. Row and Col are 0-indexed (tree-sitter convention); the caller
// must add 1 to Row when converting to the 1-indexed symbols.Ref.StartLine.
type expandedRef struct {
	Name     string
	Kind     symbols.RefKind
	Relation string
	Row      uint32 // 0-indexed
	Col      uint32 // 0-indexed
	EndRow   uint32 // 0-indexed
	EndCol   uint32 // 0-indexed
}

// identRe matches Python identifiers (including leading underscores).
var identRe = regexp.MustCompile(`\b([_A-Za-z][_A-Za-z0-9]*)\b`)

// expandCapture converts a single tree-sitter capture into 0-N expanded refs
// when the captured node contains embedded type information that cannot be
// recovered from a single identifier node (e.g., Python string literals used
// as forward-reference type annotations).
//
// Returns (refs, true) when expansion was performed; the caller should skip its
// normal single-ref processing for this capture.
// Returns (nil, false) when no expansion is needed; the caller falls through to
// its normal logic.
func expandCapture(langName, capName string, node Node, source []byte) ([]expandedRef, bool) {
	if langName == "python" && capName == "ref.annotation" && node.Type() == "type" {
		return expandPythonTypeAnnotation(node, source), true
	}
	return nil, false
}

// expandPythonTypeAnnotation walks the subtree of a Python `type` annotation
// node and emits one expandedRef for every identifier found inside any string
// literal descendant.
//
// This handles all three quoted-annotation patterns:
//
//	Pattern 1 – direct quoted return type:
//	  def foo() -> "MyClass": ...
//	  tree: (type (string "\"MyClass\""))
//
//	Pattern 2 – quoted parameter type:
//	  def bar(x: "MyClass"): ...
//	  tree: (typed_parameter type: (type (string "\"MyClass\"")))
//
//	Pattern 3 – quoted element inside generic:
//	  def baz() -> List["MyClass"]: ...
//	  def qux() -> tuple["MyClass", bool]: ...
//	  tree: (type (subscript ... (string "\"MyClass\"")))
//	     or (type (generic_type ... (type_parameter (string "\"MyClass\""))))
//
// Plain identifiers inside type annotations (e.g., `MyClass` in `-> MyClass`)
// are already captured by the existing `(identifier) @ref.identifier` pattern;
// the deduplication pass in Extract() ensures they are not double-counted.
func expandPythonTypeAnnotation(node Node, source []byte) []expandedRef {
	stringNodes := collectStringDescendants(node)
	if len(stringNodes) == 0 {
		return nil
	}
	var refs []expandedRef
	for _, sn := range stringNodes {
		refs = append(refs, extractRefsFromAnnotationString(sn, source)...)
	}
	return refs
}

// collectStringDescendants returns all string-typed nodes in the subtree
// rooted at node. Recursion stops at string nodes (they never contain other
// strings as children in Python grammar).
func collectStringDescendants(node Node) []Node {
	if node.IsNull() {
		return nil
	}
	if node.Type() == "string" {
		return []Node{node}
	}
	// Iterate ALL children (not just named ones) to be robust against
	// grammar versions that differ in which children are named.
	count := node.ChildCount()
	var result []Node
	for i := uint32(0); i < count; i++ {
		child := node.Child(i)
		result = append(result, collectStringDescendants(child)...)
	}
	return result
}

// extractRefsFromAnnotationString parses a single Python string literal node
// that appears in a type annotation context and returns one expandedRef per
// identifier found inside the string's text (including identifiers buried
// inside complex type expressions like "tuple[MyClass, bool]").
//
// Quote characters are transparent to the regex — `\b` word boundaries handle
// the transition from `"` to the first letter of the identifier automatically.
// Position mapping uses offsetToPos so the reported column points to the
// identifier itself (inside the quotes), not the opening quote character.
func extractRefsFromAnnotationString(strNode Node, source []byte) []expandedRef {
	content := strNode.Content(source)
	if len(content) < 3 { // minimum valid: "x" (3 bytes including quotes)
		return nil
	}
	start := strNode.StartPoint()
	matches := identRe.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return nil
	}
	refs := make([]expandedRef, 0, len(matches))
	for _, m := range matches {
		name := content[m[0]:m[1]]
		row, col := offsetToPos(content, start.Row, start.Column, m[0])
		refs = append(refs, expandedRef{
			Name:     name,
			Kind:     symbols.RefAnnotation,
			Relation: "annotation",
			Row:      row,
			Col:      col,
			EndRow:   row,
			EndCol:   col + uint32(len(name)),
		})
	}
	return refs
}

// offsetToPos maps a byte offset within content to an absolute (row, col)
// source position, given the 0-indexed row and col of content[0].
func offsetToPos(content string, baseRow, baseCol uint32, offset int) (row, col uint32) {
	row = baseRow
	col = baseCol
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			row++
			col = 0
		} else {
			col++
		}
	}
	return row, col
}
