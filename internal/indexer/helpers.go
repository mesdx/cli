package indexer

import (
	"regexp"
	"strings"

	"github.com/mesdx/cli/internal/symbols"
)

// lineAt returns the 1-based line number for a byte offset in content.
func lineAt(content string, offset int) int {
	return strings.Count(content[:offset], "\n") + 1
}

// colAt returns the 0-based column of the first occurrence of name in the
// matched region starting at offset.
func colAt(content string, offset int, name string) int {
	// Find the start of the line
	lineStart := strings.LastIndex(content[:offset], "\n")
	if lineStart < 0 {
		lineStart = 0
	} else {
		lineStart++ // skip the \n
	}
	// Find name within the matched region
	region := content[offset:]
	idx := strings.Index(region, name)
	if idx < 0 {
		return 0
	}
	return (offset + idx) - lineStart
}

// identRe matches word-boundary identifiers (2+ chars, starting with a letter or _).
var identRe = regexp.MustCompile(`\b([A-Za-z_]\w{1,})\b`)

// extractIdentRefs scans lines and returns simple identifier references.
// It filters out very common keywords and single-character identifiers.
func extractIdentRefs(lines []string, _ string) []symbols.Ref {
	var refs []symbols.Ref
	for i, line := range lines {
		// Skip comment-only lines (simple heuristic)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		for _, loc := range identRe.FindAllStringIndex(line, -1) {
			name := line[loc[0]:loc[1]]
			if isCommonKeyword(name) {
				continue
			}
			refs = append(refs, symbols.Ref{
				Name:      name,
				Kind:      symbols.RefOther,
				StartLine: i + 1,
				StartCol:  loc[0],
				EndLine:   i + 1,
				EndCol:    loc[1],
			})
		}
	}
	return refs
}

// isCommonKeyword returns true for keywords/builtins shared across many languages.
func isCommonKeyword(s string) bool {
	switch s {
	case "if", "else", "for", "while", "do", "switch", "case", "break",
		"continue", "return", "try", "catch", "finally", "throw",
		"new", "this", "super", "null", "true", "false", "void",
		"import", "export", "from", "package", "class", "interface",
		"struct", "enum", "fn", "func", "def", "let", "var", "const",
		"type", "pub", "public", "private", "protected", "static",
		"final", "abstract", "async", "await", "yield", "self",
		"Self", "use", "mod", "crate", "extern", "impl", "trait",
		"where", "mut", "ref", "unsafe", "dyn", "move",
		"match", "loop", "in", "of", "as", "is",
		"not", "and", "or", "with", "pass", "raise", "except",
		"None", "True", "False", "lambda", "nonlocal", "global",
		"elif", "assert", "del", "println", "print":
		return true
	}
	return false
}
