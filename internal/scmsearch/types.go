package scmsearch

// SearchRequest describes a Tree-sitter SCM query search over source files.
type SearchRequest struct {
	Language string `json:"language"`

	// Raw Tree-sitter query (S-expression). Mutually exclusive with StubName.
	Query string `json:"query,omitempty"`

	// Predefined query stub. Mutually exclusive with Query.
	StubName string            `json:"stubName,omitempty"`
	StubArgs map[string]string `json:"stubArgs,omitempty"`

	// File filters (glob patterns relative to repo root).
	IncludeGlobs []string `json:"includeGlobs,omitempty"`
	ExcludeGlobs []string `json:"excludeGlobs,omitempty"`

	// Optional override for source roots; defaults to config source roots.
	SourceRoots []string `json:"sourceRoots,omitempty"`

	ContextLines          int `json:"contextLines,omitempty"`          // default 2
	MaxFiles              int `json:"maxFiles,omitempty"`              // default 5000
	MaxMatches            int `json:"maxMatches,omitempty"`            // default 500
	MaxMatchesPerFile     int `json:"maxMatchesPerFile,omitempty"`     // default 100
	ASTParentDepth        int `json:"astParentDepth,omitempty"`        // default 3
	NodeTextLimitBytes    int `json:"nodeTextLimitBytes,omitempty"`    // default 512
	Workers               int `json:"workers,omitempty"`               // default GOMAXPROCS
}

// SearchResult is the structured output of a search.
type SearchResult struct {
	Matches []Match `json:"matches"`
	Summary Summary `json:"summary"`
}

// Match is a single capture hit from a Tree-sitter query.
type Match struct {
	FilePath      string   `json:"filePath"`
	StartLine     int      `json:"startLine"`
	StartCol      int      `json:"startCol"`
	EndLine       int      `json:"endLine"`
	EndCol        int      `json:"endCol"`
	CaptureName   string   `json:"captureName"`
	NodeType      string   `json:"nodeType"`
	TextSnippet   string   `json:"textSnippet"`
	Line          string   `json:"line"`
	ContextBefore []string `json:"contextBefore,omitempty"`
	ContextAfter  []string `json:"contextAfter,omitempty"`
	ASTParents    []string `json:"astParents,omitempty"`
}

// Summary provides execution metadata.
type Summary struct {
	FilesScanned    int `json:"filesScanned"`
	FilesMatched    int `json:"filesMatched"`
	MatchesReturned int `json:"matchesReturned"`
	DurationMs      int `json:"durationMs"`
	CacheHits       int `json:"cacheHits"`
	CacheMisses     int `json:"cacheMisses"`
}
