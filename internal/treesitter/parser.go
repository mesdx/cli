package treesitter

/*
#cgo CFLAGS: -I${SRCDIR}
#include "tree_sitter_api.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Parser wraps a tree-sitter parser.
type Parser struct {
	c *C.TSParser
}

// NewParser creates a new tree-sitter parser.
func NewParser() *Parser {
	return &Parser{
		c: C.ts_parser_new(),
	}
}

// SetLanguage sets the language for the parser.
func (p *Parser) SetLanguage(lang *Language) error {
	if !C.ts_parser_set_language(p.c, lang.lang) {
		return fmt.Errorf("failed to set language: version mismatch or invalid language")
	}
	return nil
}

// ParseString parses a source string and returns a tree.
func (p *Parser) ParseString(oldTree *Tree, source []byte) *Tree {
	var cOldTree *C.TSTree
	if oldTree != nil {
		cOldTree = oldTree.c
	}

	cSource := (*C.char)(unsafe.Pointer(&source[0]))
	cLength := C.uint32_t(len(source))

	cTree := C.ts_parser_parse_string(p.c, cOldTree, cSource, cLength)
	if cTree == nil {
		return nil
	}

	return &Tree{c: cTree}
}

// Close deletes the parser.
func (p *Parser) Close() {
	if p.c != nil {
		C.ts_parser_delete(p.c)
		p.c = nil
	}
}

// Tree wraps a tree-sitter tree.
type Tree struct {
	c *C.TSTree
}

// RootNode returns the root node of the tree.
func (t *Tree) RootNode() Node {
	return Node{c: C.ts_tree_root_node(t.c)}
}

// Close deletes the tree.
func (t *Tree) Close() {
	if t.c != nil {
		C.ts_tree_delete(t.c)
		t.c = nil
	}
}

// Node wraps a tree-sitter node.
type Node struct {
	c C.TSNode
}

// IsNull returns true if the node is null.
func (n Node) IsNull() bool {
	return bool(C.ts_node_is_null(n.c))
}

// Type returns the node's type string.
func (n Node) Type() string {
	return C.GoString(C.ts_node_type(n.c))
}

// IsNamed returns true if the node is named.
func (n Node) IsNamed() bool {
	return bool(C.ts_node_is_named(n.c))
}

// StartByte returns the start byte offset.
func (n Node) StartByte() uint32 {
	return uint32(C.ts_node_start_byte(n.c))
}

// EndByte returns the end byte offset.
func (n Node) EndByte() uint32 {
	return uint32(C.ts_node_end_byte(n.c))
}

// StartPoint returns the start point (row, column).
func (n Node) StartPoint() Point {
	p := C.ts_node_start_point(n.c)
	return Point{Row: uint32(p.row), Column: uint32(p.column)}
}

// EndPoint returns the end point (row, column).
func (n Node) EndPoint() Point {
	p := C.ts_node_end_point(n.c)
	return Point{Row: uint32(p.row), Column: uint32(p.column)}
}

// ChildCount returns the number of children.
func (n Node) ChildCount() uint32 {
	return uint32(C.ts_node_child_count(n.c))
}

// Child returns the child at the given index.
func (n Node) Child(index uint32) Node {
	return Node{c: C.ts_node_child(n.c, C.uint32_t(index))}
}

// NamedChildCount returns the number of named children.
func (n Node) NamedChildCount() uint32 {
	return uint32(C.ts_node_named_child_count(n.c))
}

// NamedChild returns the named child at the given index.
func (n Node) NamedChild(index uint32) Node {
	return Node{c: C.ts_node_named_child(n.c, C.uint32_t(index))}
}

// ChildByFieldName returns the child with the given field name.
func (n Node) ChildByFieldName(fieldName string) Node {
	cFieldName := C.CString(fieldName)
	defer C.free(unsafe.Pointer(cFieldName))
	return Node{c: C.ts_node_child_by_field_name(n.c, cFieldName, C.uint32_t(len(fieldName)))}
}

// NextSibling returns the next sibling node.
func (n Node) NextSibling() Node {
	return Node{c: C.ts_node_next_sibling(n.c)}
}

// NextNamedSibling returns the next named sibling node.
func (n Node) NextNamedSibling() Node {
	return Node{c: C.ts_node_next_named_sibling(n.c)}
}

// Parent returns the parent node.
func (n Node) Parent() Node {
	return Node{c: C.ts_node_parent(n.c)}
}

// Content extracts the node's text content from the source.
func (n Node) Content(source []byte) string {
	start := n.StartByte()
	end := n.EndByte()
	if start >= uint32(len(source)) || end > uint32(len(source)) {
		return ""
	}
	return string(source[start:end])
}

// Point represents a position in the source.
type Point struct {
	Row    uint32
	Column uint32
}
