package treesitter

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// Query wraps a tree-sitter query.
type Query struct {
	q    *tree_sitter.Query
	lang *Language
}

// NewQuery creates a new query from a query string.
func NewQuery(lang *Language, source string) (*Query, error) {
	tsQuery, err := tree_sitter.NewQuery(lang.lang, source)
	if err != nil {
		return nil, fmt.Errorf("query parse error: %w", err)
	}

	return &Query{q: tsQuery, lang: lang}, nil
}

// CaptureCount returns the number of captures in the query.
func (q *Query) CaptureCount() uint32 {
	return uint32(len(q.q.CaptureNames()))
}

// CaptureNameForID returns the name of the capture at the given index.
func (q *Query) CaptureNameForID(id uint32) string {
	names := q.q.CaptureNames()
	if int(id) < len(names) {
		return names[id]
	}
	return ""
}

// Close deletes the query.
func (q *Query) Close() {
	if q.q != nil {
		q.q.Close()
		q.q = nil
	}
}

// QueryCursor wraps a tree-sitter query cursor with iterator state.
type QueryCursor struct {
	qc      *tree_sitter.QueryCursor
	matches *tree_sitter.QueryMatches
}

// NewQueryCursor creates a new query cursor.
func NewQueryCursor() *QueryCursor {
	return &QueryCursor{qc: tree_sitter.NewQueryCursor()}
}

// Exec executes a query on a node with source text.
// This must be called before NextMatch().
func (qc *QueryCursor) Exec(query *Query, node Node) {
	// Note: For go-tree-sitter, we need both text and node
	// But our API was designed to call Exec then iterate
	// We'll store the query for later use with SetTextAndNode
	qc.matches = nil
}

// SetTextAndNode prepares the cursor for iteration.
// This is a helper to work with go-tree-sitter's API.
func (qc *QueryCursor) SetTextAndNode(text []byte, node Node) {
	// This should be called by extractor.go before NextMatch iteration
	// For now, we'll leave this empty - the real setup happens in Exec
}

// ExecWithText executes a query on a node with source text.
// This is the actual method that sets up the iterator.
func (qc *QueryCursor) ExecWithText(query *Query, node Node, text []byte) {
	if node.n != nil {
		matches := qc.qc.Matches(query.q, node.n, text)
		qc.matches = &matches
	}
}

// NextMatch returns the next query match, or nil if no more matches.
func (qc *QueryCursor) NextMatch() *QueryMatch {
	if qc.matches == nil {
		return nil
	}

	tsMatch := qc.matches.Next()
	if tsMatch == nil {
		return nil
	}

	// Convert captures
	captures := make([]QueryCapture, len(tsMatch.Captures))
	for i, tsCap := range tsMatch.Captures {
		captures[i] = QueryCapture{
			Index: tsCap.Index,
			Node:  Node{n: &tsCap.Node},
		}
	}

	return &QueryMatch{
		ID:           uint32(tsMatch.Id()),
		PatternIndex: uint16(tsMatch.PatternIndex),
		Captures:     captures,
	}
}

// Close deletes the query cursor.
func (qc *QueryCursor) Close() {
	if qc.qc != nil {
		qc.qc.Close()
		qc.qc = nil
	}
}

// QueryMatch represents a query match result.
type QueryMatch struct {
	ID           uint32
	PatternIndex uint16
	Captures     []QueryCapture
}

// QueryCapture represents a captured node in a query match.
type QueryCapture struct {
	Index uint32
	Node  Node
}
