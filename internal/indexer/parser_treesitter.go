package indexer

import (
	"fmt"
	"sync"

	"github.com/mesdx/cli/internal/symbols"
	"github.com/mesdx/cli/internal/treesitter"
)

// TreeSitterParser uses tree-sitter for parsing.
type TreeSitterParser struct {
	langName  string
	extractor *treesitter.Extractor
	once      sync.Once
	initErr   error
}

// NewTreeSitterParser creates a new tree-sitter parser for the given language.
func NewTreeSitterParser(langName string) *TreeSitterParser {
	return &TreeSitterParser{
		langName: langName,
	}
}

// Parse parses the source using tree-sitter.
func (p *TreeSitterParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	// Lazy initialization
	p.once.Do(func() {
		p.extractor, p.initErr = treesitter.NewExtractor(p.langName)
	})

	if p.initErr != nil {
		return nil, fmt.Errorf("tree-sitter parser init for %s: %w", p.langName, p.initErr)
	}

	return p.extractor.Extract(filename, src)
}
