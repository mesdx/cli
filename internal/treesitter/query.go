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

// Query wraps a tree-sitter query.
type Query struct {
	c    *C.TSQuery
	lang *Language
}

// NewQuery creates a new query from a query string.
func NewQuery(lang *Language, source string) (*Query, error) {
	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))

	var errorOffset C.uint32_t
	var errorType C.uint32_t

	cQuery := C.ts_query_new(
		lang.lang,
		cSource,
		C.uint32_t(len(source)),
		&errorOffset,
		&errorType,
	)

	if cQuery == nil {
		return nil, fmt.Errorf("query parse error at offset %d (type %d)", errorOffset, errorType)
	}

	return &Query{c: cQuery, lang: lang}, nil
}

// CaptureCount returns the number of captures in the query.
func (q *Query) CaptureCount() uint32 {
	return uint32(C.ts_query_capture_count(q.c))
}

// CaptureNameForID returns the name of the capture at the given index.
func (q *Query) CaptureNameForID(id uint32) string {
	var length C.uint32_t
	name := C.ts_query_capture_name_for_id(q.c, C.uint32_t(id), &length)
	return C.GoStringN(name, C.int(length))
}

// Close deletes the query.
func (q *Query) Close() {
	if q.c != nil {
		C.ts_query_delete(q.c)
		q.c = nil
	}
}

// QueryCursor wraps a tree-sitter query cursor.
type QueryCursor struct {
	c *C.TSQueryCursor
}

// NewQueryCursor creates a new query cursor.
func NewQueryCursor() *QueryCursor {
	return &QueryCursor{c: C.ts_query_cursor_new()}
}

// Exec executes a query on a node.
func (qc *QueryCursor) Exec(query *Query, node Node) {
	C.ts_query_cursor_exec(qc.c, query.c, node.c)
}

// NextMatch returns the next query match, or nil if no more matches.
func (qc *QueryCursor) NextMatch() *QueryMatch {
	var cMatch C.TSQueryMatch
	if !C.ts_query_cursor_next_match(qc.c, &cMatch) {
		return nil
	}

	// Convert captures
	captureCount := int(cMatch.capture_count)
	captures := make([]QueryCapture, captureCount)

	// Access the captures array
	cCaptures := (*[1 << 30]C.TSQueryCapture)(unsafe.Pointer(cMatch.captures))[:captureCount:captureCount]
	for i := 0; i < captureCount; i++ {
		captures[i] = QueryCapture{
			Index: uint32(cCaptures[i].index),
			Node:  Node{c: cCaptures[i].node},
		}
	}

	return &QueryMatch{
		ID:           uint32(cMatch.id),
		PatternIndex: uint16(cMatch.pattern_index),
		Captures:     captures,
	}
}

// Close deletes the query cursor.
func (qc *QueryCursor) Close() {
	if qc.c != nil {
		C.ts_query_cursor_delete(qc.c)
		qc.c = nil
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
