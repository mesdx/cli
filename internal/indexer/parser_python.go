package indexer

import (
	"regexp"
	"strings"

	"github.com/codeintelx/cli/internal/symbols"
)

// PythonParser extracts symbols from Python source files using regex patterns.
type PythonParser struct{}

var (
	pyClassRe    = regexp.MustCompile(`(?m)^class\s+(\w+)`)
	pyFuncRe     = regexp.MustCompile(`(?m)^(\s*)(?:async\s+)?def\s+(\w+)\s*\(([^)]*)\)`)
	pyVarRe      = regexp.MustCompile(`(?m)^(\w+)\s*(?::\s*\w[\w\[\], ]*\s*)?=`)
	pyImportRe   = regexp.MustCompile(`(?m)^(?:from\s+([\w.]+)\s+)?import\s+(.+)`)
	pyPropertyRe = regexp.MustCompile(`(?m)^[ \t]+@property`)
)

func (p *PythonParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	content := string(src)
	lines := strings.Split(content, "\n")
	result := &symbols.FileResult{}

	currentClass := ""

	// Classes — also record their line numbers for scope tracking
	var classes []pyClassInfo

	for _, m := range pyClassRe.FindAllSubmatchIndex([]byte(content), -1) {
		if m[2] < 0 {
			continue
		}
		name := content[m[2]:m[3]]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		currentClass = name
		classes = append(classes, pyClassInfo{name: name, line: line})
		endLine := pythonBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindClass,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Find property decorators (line numbers)
	propertyLines := map[int]bool{}
	for _, m := range pyPropertyRe.FindAllStringIndex(content, -1) {
		line := lineAt(content, m[0])
		propertyLines[line] = true
	}

	// Functions/methods — use FindAllSubmatchIndex for accurate capture positions
	for _, m := range pyFuncRe.FindAllSubmatchIndex([]byte(content), -1) {
		if m[0] < 0 {
			continue
		}
		// m[2]:m[3] = group 1 (indent)
		// m[4]:m[5] = group 2 (name)
		// m[6]:m[7] = group 3 (params)
		if m[4] < 0 {
			continue
		}

		indent := ""
		if m[2] >= 0 && m[3] > m[2] {
			indent = content[m[2]:m[3]]
		}
		// Strip newlines — (\s*) can capture \n in multiline mode
		indent = strings.ReplaceAll(indent, "\n", "")
		indent = strings.ReplaceAll(indent, "\r", "")

		name := content[m[4]:m[5]]
		params := ""
		if m[6] >= 0 {
			params = content[m[6]:m[7]]
		}
		line := lineAt(content, m[0])
		// Compute col from the line start
		lineStart := strings.LastIndex(content[:m[4]], "\n") + 1
		col := m[4] - lineStart

		kind := symbols.KindFunction
		container := ""

		if len(indent) > 0 {
			// Indented — likely a method inside a class
			kind = symbols.KindMethod
			// Find which class this belongs to
			container = classForLine(classes, line)
			if container == "" {
				container = currentClass
			}
			// Check if preceded by @property
			if propertyLines[line-1] {
				kind = symbols.KindProperty
			}
		}

		sig := "def " + name + "(" + strings.TrimSpace(params) + ")"
		endLine := pythonBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          kind,
			ContainerName: container,
			Signature:     sig,
			StartLine:     line,
			StartCol:      col,
			EndLine:       endLine,
			EndCol:        col + len(name),
		})
	}

	// Module-level variables / constants
	for _, m := range pyVarRe.FindAllSubmatchIndex([]byte(content), -1) {
		if m[0] < 0 || m[2] < 0 {
			continue
		}
		name := content[m[2]:m[3]]
		if name == "_" || strings.HasPrefix(name, "__") {
			continue
		}
		line := lineAt(content, m[0])
		lineStart := strings.LastIndex(content[:m[2]], "\n") + 1
		col := m[2] - lineStart
		kind := symbols.KindVariable
		if name == strings.ToUpper(name) && len(name) > 1 {
			kind = symbols.KindConstant
		}
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      kind,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Import references
	for _, m := range pyImportRe.FindAllStringIndex(content, -1) {
		sub := pyImportRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 3 {
			continue
		}
		importStr := strings.TrimSpace(sub[2])
		for _, part := range strings.Split(importStr, ",") {
			part = strings.TrimSpace(part)
			if idx := strings.Index(part, " as "); idx >= 0 {
				part = part[:idx]
			}
			part = strings.TrimSpace(part)
			if part == "" || part == "*" {
				continue
			}
			line := lineAt(content, m[0])
			result.Refs = append(result.Refs, symbols.Ref{
				Name:      part,
				Kind:      symbols.RefImport,
				StartLine: line,
				StartCol:  0,
				EndLine:   line,
				EndCol:    len(part),
			})
		}
	}

	// General references
	result.Refs = append(result.Refs, extractIdentRefs(lines, content)...)

	return result, nil
}

// pyClassInfo is used for tracking class scope in Python files.
type pyClassInfo struct {
	name string
	line int
}

// classForLine returns the class name that owns the given line (last class defined before this line).
func classForLine(classes []pyClassInfo, line int) string {
	best := ""
	for _, c := range classes {
		if c.line < line {
			best = c.name
		}
	}
	return best
}
