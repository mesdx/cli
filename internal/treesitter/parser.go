package treesitter

import (
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// Parser wraps a tree-sitter parser.
type Parser struct {
	p *tree_sitter.Parser
}

// NewParser creates a new tree-sitter parser.
func NewParser() *Parser {
	return &Parser{
		p: tree_sitter.NewParser(),
	}
}

// SetLanguage sets the language for the parser.
func (p *Parser) SetLanguage(lang *Language) error {
	p.p.SetLanguage(lang.lang)
	return nil
}

// ParseString parses a source string and returns a tree.
func (p *Parser) ParseString(oldTree *Tree, source []byte) *Tree {
	var tsOldTree *tree_sitter.Tree
	if oldTree != nil {
		tsOldTree = oldTree.t
	}

	tsTree := p.p.Parse(source, tsOldTree)
	if tsTree == nil {
		return nil
	}

	return &Tree{t: tsTree}
}

// Close deletes the parser.
func (p *Parser) Close() {
	if p.p != nil {
		p.p.Close()
		p.p = nil
	}
}

// Tree wraps a tree-sitter tree.
type Tree struct {
	t *tree_sitter.Tree
}

// RootNode returns the root node of the tree.
func (t *Tree) RootNode() Node {
	return Node{n: t.t.RootNode()}
}

// Close deletes the tree.
func (t *Tree) Close() {
	if t.t != nil {
		t.t.Close()
		t.t = nil
	}
}

// Node wraps a tree-sitter node.
type Node struct {
	n *tree_sitter.Node
}

// IsNull returns true if the node is null.
func (n Node) IsNull() bool {
	return n.n == nil
}

// Type returns the node's type string.
func (n Node) Type() string {
	if n.n == nil {
		return ""
	}
	return n.n.Kind()
}

// IsNamed returns true if the node is named.
func (n Node) IsNamed() bool {
	if n.n == nil {
		return false
	}
	return n.n.IsNamed()
}

// StartByte returns the start byte offset.
func (n Node) StartByte() uint32 {
	if n.n == nil {
		return 0
	}
	return uint32(n.n.StartByte())
}

// EndByte returns the end byte offset.
func (n Node) EndByte() uint32 {
	if n.n == nil {
		return 0
	}
	return uint32(n.n.EndByte())
}

// StartPoint returns the start point (row, column).
func (n Node) StartPoint() Point {
	if n.n == nil {
		return Point{}
	}
	p := n.n.StartPosition()
	return Point{Row: uint32(p.Row), Column: uint32(p.Column)}
}

// EndPoint returns the end point (row, column).
func (n Node) EndPoint() Point {
	if n.n == nil {
		return Point{}
	}
	p := n.n.EndPosition()
	return Point{Row: uint32(p.Row), Column: uint32(p.Column)}
}

// ChildCount returns the number of children.
func (n Node) ChildCount() uint32 {
	if n.n == nil {
		return 0
	}
	return uint32(n.n.ChildCount())
}

// Child returns the child at the given index.
func (n Node) Child(index uint32) Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.Child(uint(index))}
}

// NamedChildCount returns the number of named children.
func (n Node) NamedChildCount() uint32 {
	if n.n == nil {
		return 0
	}
	return uint32(n.n.NamedChildCount())
}

// NamedChild returns the named child at the given index.
func (n Node) NamedChild(index uint32) Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.NamedChild(uint(index))}
}

// ChildByFieldName returns the child with the given field name.
func (n Node) ChildByFieldName(fieldName string) Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.ChildByFieldName(fieldName)}
}

// NextSibling returns the next sibling node.
func (n Node) NextSibling() Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.NextSibling()}
}

// NextNamedSibling returns the next named sibling node.
func (n Node) NextNamedSibling() Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.NextNamedSibling()}
}

// Parent returns the parent node.
func (n Node) Parent() Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.Parent()}
}

// Content extracts the node's text content from the source.
func (n Node) Content(source []byte) string {
	if n.n == nil {
		return ""
	}
	return n.n.Utf8Text(source)
}

// Point represents a position in the source.
type Point struct {
	Row    uint32
	Column uint32
}
