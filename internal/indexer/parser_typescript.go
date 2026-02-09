package indexer

import (
	"regexp"
	"strings"

	"github.com/mesdx/cli/internal/symbols"
)

// TypeScriptParser extracts symbols from TypeScript and JavaScript source files.
type TypeScriptParser struct{}

var (
	tsClassRe     = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:default\s+)?(?:abstract\s+)?class\s+(\w+)`)
	tsInterfaceRe = regexp.MustCompile(`(?m)^\s*(?:export\s+)?interface\s+(\w+)`)
	tsTypeRe      = regexp.MustCompile(`(?m)^\s*(?:export\s+)?type\s+(\w+)`)
	tsEnumRe      = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const\s+)?enum\s+(\w+)`)
	tsFuncRe      = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*[\(<]`)
	tsArrowRe     = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*(?::\s*[^=]+?)?\s*=\s*(?:async\s+)?\([^)]*\)[^=]*=>`)
	tsMethodRe    = regexp.MustCompile(`(?m)^\s+(?:public\s+|private\s+|protected\s+)?(?:static\s+)?(?:async\s+)?(?:get\s+|set\s+)?(\w+)\s*[\(<]`)
	tsVarRe       = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)`)
	tsModuleRe    = regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:declare\s+)?(?:namespace|module)\s+(\w+)`)
)

func (p *TypeScriptParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	content := string(src)
	lines := strings.Split(content, "\n")
	result := &symbols.FileResult{}

	currentClass := ""

	// Track which names are already captured as arrow functions to avoid duplicating as variables
	arrowNames := map[string]bool{}

	// Classes
	for _, m := range tsClassRe.FindAllStringIndex(content, -1) {
		sub := tsClassRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		currentClass = name
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindClass,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Interfaces
	for _, m := range tsInterfaceRe.FindAllStringIndex(content, -1) {
		sub := tsInterfaceRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindInterface,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Type aliases
	for _, m := range tsTypeRe.FindAllStringIndex(content, -1) {
		sub := tsTypeRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindTypeAlias,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Enums
	for _, m := range tsEnumRe.FindAllStringIndex(content, -1) {
		sub := tsEnumRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindEnum,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Functions
	for _, m := range tsFuncRe.FindAllStringIndex(content, -1) {
		sub := tsFuncRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if isTSKeyword(name) {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindFunction,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Arrow functions assigned to const/let/var
	for _, m := range tsArrowRe.FindAllStringIndex(content, -1) {
		sub := tsArrowRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if isTSKeyword(name) {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		arrowNames[name] = true
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindFunction,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Methods (inside classes)
	for _, m := range tsMethodRe.FindAllStringIndex(content, -1) {
		sub := tsMethodRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if isTSKeyword(name) || name == "constructor" {
			if name == "constructor" {
				line := lineAt(content, m[0])
				col := colAt(content, m[0], name)
				endLine := findBlockEndLine(lines, line)
				result.Symbols = append(result.Symbols, symbols.Symbol{
					Name:          name,
					Kind:          symbols.KindConstructor,
					ContainerName: currentClass,
					StartLine:     line,
					StartCol:      col,
					EndLine:       endLine,
					EndCol:        col + len(name),
				})
			}
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          symbols.KindMethod,
			ContainerName: currentClass,
			StartLine:     line,
			StartCol:      col,
			EndLine:       endLine,
			EndCol:        col + len(name),
		})
	}

	// Module-level variables (excluding arrow functions already captured)
	for _, m := range tsVarRe.FindAllStringIndex(content, -1) {
		sub := tsVarRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if isTSKeyword(name) || arrowNames[name] {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindVariable,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// Namespace/module
	for _, m := range tsModuleRe.FindAllStringIndex(content, -1) {
		sub := tsModuleRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindModule,
			StartLine: line,
			StartCol:  col,
			EndLine:   endLine,
			EndCol:    col + len(name),
		})
	}

	// General references
	result.Refs = extractIdentRefs(lines, content)

	return result, nil
}

func isTSKeyword(s string) bool {
	switch s {
	case "if", "else", "for", "while", "do", "switch", "case", "break",
		"continue", "return", "try", "catch", "finally", "throw",
		"new", "this", "super", "class", "extends", "implements",
		"import", "export", "from", "default", "void", "null",
		"undefined", "true", "false", "typeof", "instanceof",
		"in", "of", "as", "is", "keyof", "readonly", "declare",
		"abstract", "async", "await", "yield", "delete",
		"function", "const", "let", "var", "type", "interface",
		"enum", "namespace", "module", "get", "set",
		"public", "private", "protected", "static", "override":
		return true
	}
	return false
}
