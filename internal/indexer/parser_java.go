package indexer

import (
	"regexp"
	"strings"

	"github.com/codeintelx/cli/internal/symbols"
)

// JavaParser extracts symbols from Java source files using regex patterns.
type JavaParser struct{}

var (
	javaClassRe       = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?(?:abstract\s+|final\s+|static\s+)*class\s+(\w+)`)
	javaInterfaceRe   = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?interface\s+(\w+)`)
	javaEnumRe        = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?enum\s+(\w+)`)
	javaMethodRe      = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?(?:abstract\s+|static\s+|final\s+|synchronized\s+|native\s+)*(?:[\w<>\[\],\s]+?)\s+(\w+)\s*\(([^)]*)\)`)
	javaFieldRe       = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?(?:static\s+|final\s+|volatile\s+|transient\s+)*(?:[\w<>\[\],]+)\s+(\w+)\s*[;=]`)
	javaConstructorRe = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?(\w+)\s*\(([^)]*)\)\s*(?:throws\s+[\w,\s]+)?\s*\{`)
)

func (p *JavaParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	content := string(src)
	lines := strings.Split(content, "\n")
	result := &symbols.FileResult{}

	// Track current class context for container names
	currentClass := ""

	// Classes
	for _, m := range javaClassRe.FindAllStringIndex(content, -1) {
		sub := javaClassRe.FindStringSubmatch(content[m[0]:m[1]])
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
	for _, m := range javaInterfaceRe.FindAllStringIndex(content, -1) {
		sub := javaInterfaceRe.FindStringSubmatch(content[m[0]:m[1]])
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

	// Enums
	for _, m := range javaEnumRe.FindAllStringIndex(content, -1) {
		sub := javaEnumRe.FindStringSubmatch(content[m[0]:m[1]])
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

	// Methods
	for _, m := range javaMethodRe.FindAllStringIndex(content, -1) {
		sub := javaMethodRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		// Skip if this is actually a keyword
		if isJavaKeyword(name) {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		sig := name + "(" + strings.TrimSpace(sub[2]) + ")"
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          symbols.KindMethod,
			ContainerName: currentClass,
			Signature:     sig,
			StartLine:     line,
			StartCol:      col,
			EndLine:       endLine,
			EndCol:        col + len(name),
		})
	}

	// Constructors
	for _, m := range javaConstructorRe.FindAllStringIndex(content, -1) {
		sub := javaConstructorRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		// Constructor name must match a known class
		if !isUpperFirst(name) {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          symbols.KindConstructor,
			ContainerName: name,
			StartLine:     line,
			StartCol:      col,
			EndLine:       endLine,
			EndCol:        col + len(name),
		})
	}

	// Fields (simple heuristic — only top-level-ish declarations)
	for _, m := range javaFieldRe.FindAllStringIndex(content, -1) {
		sub := javaFieldRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if isJavaKeyword(name) {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		endLine := findBlockEndLine(lines, line)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          symbols.KindField,
			ContainerName: currentClass,
			StartLine:     line,
			StartCol:      col,
			EndLine:       endLine,
			EndCol:        col + len(name),
		})
	}

	// References: identifiers in the body
	result.Refs = extractIdentRefs(lines, content)

	return result, nil
}

func isJavaKeyword(s string) bool {
	switch s {
	case "if", "else", "for", "while", "do", "switch", "case", "break",
		"continue", "return", "try", "catch", "finally", "throw", "throws",
		"new", "this", "super", "class", "interface", "enum", "extends",
		"implements", "import", "package", "void", "null", "true", "false",
		"public", "private", "protected", "static", "final", "abstract",
		"synchronized", "native", "volatile", "transient", "default",
		"instanceof", "assert":
		return true
	}
	return false
}

func isUpperFirst(s string) bool {
	return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z'
}
