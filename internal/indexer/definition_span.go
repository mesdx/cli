package indexer

import "strings"

// findBlockEndLine finds the end line of a brace-delimited block starting
// from startLine (1-based). It looks for the first '{' on or after startLine,
// then finds the matching '}' using brace depth tracking.
// Returns the 1-based line number of the closing '}'.
// If no '{' is found, it looks for ';' to handle single-statement declarations.
// Falls back to startLine if neither is found.
func findBlockEndLine(lines []string, startLine int) int {
	if startLine < 1 || startLine > len(lines) {
		return startLine
	}

	depth := 0
	foundOpen := false
	for i := startLine - 1; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			switch ch {
			case '{':
				depth++
				foundOpen = true
			case '}':
				depth--
				if foundOpen && depth == 0 {
					return i + 1 // 1-based
				}
			}
		}
		// If we haven't found a '{' yet, check for ';' (single-line statement)
		if !foundOpen {
			trimmed := strings.TrimSpace(line)
			if strings.HasSuffix(trimmed, ";") {
				return i + 1 // 1-based
			}
		}
	}

	// Fallback
	return startLine
}

// pythonBlockEndLine finds the last line of a Python indented block.
// defLine is the 1-based line number of the def/class statement.
// Returns the 1-based line number of the last line in the block.
func pythonBlockEndLine(lines []string, defLine int) int {
	if defLine < 1 || defLine > len(lines) {
		return defLine
	}

	defIdx := defLine - 1
	defIndent := leadingWhitespaceCount(lines[defIdx])

	lastBodyLine := defLine
	for i := defIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Skip blank lines
		if trimmed == "" {
			continue
		}

		// If indented more than def line, it belongs to the body
		lineIndent := leadingWhitespaceCount(lines[i])
		if lineIndent > defIndent {
			lastBodyLine = i + 1 // 1-based
		} else {
			break
		}
	}

	return lastBodyLine
}

// leadingWhitespaceCount returns the number of leading whitespace characters.
func leadingWhitespaceCount(s string) int {
	for i, ch := range s {
		if ch != ' ' && ch != '\t' {
			return i
		}
	}
	return len(s)
}

// FindDocStartLine scans backward from declLine (1-based) to include
// contiguous documentation / comment / annotation / decorator lines.
// Returns the 1-based line of the first doc line, or declLine if none.
func FindDocStartLine(lines []string, declLine int, lang Lang) int {
	if declLine <= 1 || declLine > len(lines) {
		return declLine
	}

	start := declLine
	for i := declLine - 2; i >= 0; i-- { // 0-indexed, one line before declLine
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			break // blank line ends the doc block
		}
		if isDocLine(trimmed, lang) {
			start = i + 1 // convert back to 1-based
		} else {
			break
		}
	}
	return start
}

// isDocLine returns true if the trimmed line looks like a documentation /
// comment / annotation / decorator line for the given language.
func isDocLine(trimmed string, lang Lang) bool {
	switch lang {
	case LangGo:
		return strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasSuffix(trimmed, "*/")
	case LangJava:
		return strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "/**") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasSuffix(trimmed, "*/") ||
			strings.HasPrefix(trimmed, "@")
	case LangRust:
		return strings.HasPrefix(trimmed, "///") ||
			strings.HasPrefix(trimmed, "//!") ||
			strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "#[") ||
			strings.HasPrefix(trimmed, "#![")
	case LangPython:
		return strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "@")
	case LangTypeScript, LangJavaScript:
		return strings.HasPrefix(trimmed, "//") ||
			strings.HasPrefix(trimmed, "/*") ||
			strings.HasPrefix(trimmed, "/**") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasSuffix(trimmed, "*/") ||
			strings.HasPrefix(trimmed, "@")
	}
	return false
}
