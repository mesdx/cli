package indexer

import (
	"regexp"
	"strings"

	"github.com/codeintelx/cli/internal/symbols"
)

// RustParser extracts symbols from Rust source files using regex patterns.
type RustParser struct{}

var (
	rustFnRe     = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?(?:async\s+)?(?:unsafe\s+)?(?:extern\s+"[^"]*"\s+)?fn\s+(\w+)`)
	rustStructRe = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?struct\s+(\w+)`)
	rustEnumRe   = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?enum\s+(\w+)`)
	rustTraitRe  = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?(?:unsafe\s+)?trait\s+(\w+)`)
	rustImplRe   = regexp.MustCompile(`(?m)^impl(?:<[^>]*>)?\s+(?:(\w+)\s+for\s+)?(\w+)`)
	rustConstRe  = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?(?:const|static)\s+(\w+)`)
	rustTypeRe   = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?type\s+(\w+)`)
	rustModRe    = regexp.MustCompile(`(?m)^\s*(?:pub(?:\([\w:]+\))?\s+)?mod\s+(\w+)`)
	rustUseRe    = regexp.MustCompile(`(?m)^use\s+([\w:]+)`)
)

// implBlock tracks byte range and type name of an impl block.
type implBlock struct {
	typeName string
	startOff int
	endOff   int // approximation: next impl or EOF
}

func (p *RustParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	content := string(src)
	lines := strings.Split(content, "\n")
	result := &symbols.FileResult{}

	// Detect impl blocks and their approximate ranges.
	implBlocks := findImplBlocks(content)

	// Functions / methods
	for _, m := range rustFnRe.FindAllStringIndex(content, -1) {
		sub := rustFnRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		kind := symbols.KindFunction
		container := ""
		if ib := implBlockAt(implBlocks, m[0]); ib != nil {
			kind = symbols.KindMethod
			container = ib.typeName
		}
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:          name,
			Kind:          kind,
			ContainerName: container,
			StartLine:     line,
			StartCol:      col,
			EndLine:       line,
			EndCol:        col + len(name),
		})
	}

	// Structs
	for _, m := range rustStructRe.FindAllStringIndex(content, -1) {
		sub := rustStructRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindStruct,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Enums
	for _, m := range rustEnumRe.FindAllStringIndex(content, -1) {
		sub := rustEnumRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindEnum,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Traits
	for _, m := range rustTraitRe.FindAllStringIndex(content, -1) {
		sub := rustTraitRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindTrait,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Constants / statics
	for _, m := range rustConstRe.FindAllStringIndex(content, -1) {
		sub := rustConstRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		if name == "_" {
			continue
		}
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindConstant,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Type aliases
	for _, m := range rustTypeRe.FindAllStringIndex(content, -1) {
		sub := rustTypeRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindTypeAlias,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// Modules
	for _, m := range rustModRe.FindAllStringIndex(content, -1) {
		sub := rustModRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		name := sub[1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Symbols = append(result.Symbols, symbols.Symbol{
			Name:      name,
			Kind:      symbols.KindModule,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	// References
	result.Refs = extractIdentRefs(lines, content)

	// Use statements as import refs
	for _, m := range rustUseRe.FindAllStringIndex(content, -1) {
		sub := rustUseRe.FindStringSubmatch(content[m[0]:m[1]])
		if len(sub) < 2 {
			continue
		}
		path := sub[1]
		parts := strings.Split(path, "::")
		name := parts[len(parts)-1]
		line := lineAt(content, m[0])
		col := colAt(content, m[0], name)
		result.Refs = append(result.Refs, symbols.Ref{
			Name:      name,
			Kind:      symbols.RefImport,
			StartLine: line,
			StartCol:  col,
			EndLine:   line,
			EndCol:    col + len(name),
		})
	}

	return result, nil
}

// findImplBlocks locates all `impl ... {` blocks and approximates their ranges.
func findImplBlocks(content string) []implBlock {
	var blocks []implBlock
	for _, m := range rustImplRe.FindAllSubmatchIndex([]byte(content), -1) {
		if m[0] < 0 {
			continue
		}
		// m[2]:m[3] = group 1 (trait name, may be -1 if not "impl Trait for Type")
		// m[4]:m[5] = group 2 (type name)
		if m[4] < 0 {
			continue
		}
		typeName := content[m[4]:m[5]]

		startOff := m[0]
		// Approximate end: find matching closing brace by counting braces
		endOff := findClosingBrace(content, m[1])

		blocks = append(blocks, implBlock{
			typeName: typeName,
			startOff: startOff,
			endOff:   endOff,
		})
	}
	return blocks
}

// findClosingBrace finds the position of the matching '}' starting from offset.
func findClosingBrace(content string, offset int) int {
	depth := 0
	for i := offset; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return i + 1
			}
			depth--
		}
	}
	return len(content)
}

// implBlockAt returns the impl block containing the given offset, or nil.
func implBlockAt(blocks []implBlock, offset int) *implBlock {
	for i := range blocks {
		if offset >= blocks[i].startOff && offset < blocks[i].endOff {
			return &blocks[i]
		}
	}
	return nil
}
