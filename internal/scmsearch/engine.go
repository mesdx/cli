package scmsearch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mesdx/cli/internal/indexer"
	"github.com/mesdx/cli/internal/treesitter"
)

// Engine performs parallel Tree-sitter query searches over source files.
type Engine struct {
	RepoRoot    string
	SourceRoots []string
	Cache       *QueryCache
}

// NewEngine creates a search engine with the given config.
func NewEngine(repoRoot string, sourceRoots []string, cache *QueryCache) *Engine {
	if cache == nil {
		cache = NewQueryCache(64)
	}
	return &Engine{
		RepoRoot:    repoRoot,
		SourceRoots: sourceRoots,
		Cache:       cache,
	}
}

// Search executes a search request and returns matches.
func (e *Engine) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	start := time.Now()

	querySrc, err := e.resolveQuery(req)
	if err != nil {
		return nil, err
	}

	defaults(&req)

	files, err := e.enumerateFiles(req)
	if err != nil {
		return nil, fmt.Errorf("enumerate files: %w", err)
	}
	if req.MaxFiles > 0 && len(files) > req.MaxFiles {
		files = files[:req.MaxFiles]
	}

	numWorkers := req.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.GOMAXPROCS(0)
	}
	if numWorkers > len(files) {
		numWorkers = len(files)
	}
	if numWorkers < 1 {
		return &SearchResult{Summary: Summary{DurationMs: int(time.Since(start).Milliseconds())}}, nil
	}

	// Use a derived context so we can cancel workers when MaxMatches is reached
	// without cancelling the caller's context.
	searchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type fileResult struct {
		matches []Match
		matched bool
	}

	work := make(chan string, numWorkers*2)
	results := make(chan fileResult, numWorkers*2)

	var totalMatches atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			parser := treesitter.NewParser()
			defer parser.Close()

			for filePath := range work {
				if searchCtx.Err() != nil {
					return
				}
				if req.MaxMatches > 0 && int(totalMatches.Load()) >= req.MaxMatches {
					cancel()
					return
				}

				m, err := e.searchFile(searchCtx, parser, filePath, req.Language, querySrc, req)
				if err != nil {
					continue
				}
				if len(m) > 0 {
					totalMatches.Add(int64(len(m)))
					results <- fileResult{matches: m, matched: true}
				} else {
					results <- fileResult{}
				}
			}
		}()
	}

	go func() {
	loop:
		for _, f := range files {
			select {
			case work <- f:
			case <-searchCtx.Done():
				break loop
			}
		}
		close(work)
		wg.Wait()
		close(results)
	}()

	var allMatches []Match
	filesMatched := 0
	for fr := range results {
		if fr.matched {
			filesMatched++
		}
		allMatches = append(allMatches, fr.matches...)
	}

	sort.Slice(allMatches, func(i, j int) bool {
		a, b := allMatches[i], allMatches[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.StartLine != b.StartLine {
			return a.StartLine < b.StartLine
		}
		if a.StartCol != b.StartCol {
			return a.StartCol < b.StartCol
		}
		return a.CaptureName < b.CaptureName
	})

	if req.MaxMatches > 0 && len(allMatches) > req.MaxMatches {
		allMatches = allMatches[:req.MaxMatches]
	}

	hits, misses := e.Cache.Stats()

	return &SearchResult{
		Matches: allMatches,
		Summary: Summary{
			FilesScanned:    len(files),
			FilesMatched:    filesMatched,
			MatchesReturned: len(allMatches),
			DurationMs:      int(time.Since(start).Milliseconds()),
			CacheHits:       int(hits),
			CacheMisses:     int(misses),
		},
	}, nil
}

func (e *Engine) resolveQuery(req SearchRequest) (string, error) {
	if req.Query != "" {
		return req.Query, nil
	}
	if req.StubName != "" {
		stub := LookupStub(req.StubName)
		if stub == nil {
			return "", fmt.Errorf("unknown stub %q", req.StubName)
		}
		return stub.Render(req.Language, req.StubArgs)
	}
	return "", fmt.Errorf("either query or stubName is required")
}

func defaults(req *SearchRequest) {
	if req.ContextLines <= 0 {
		req.ContextLines = 2
	}
	if req.MaxFiles <= 0 {
		req.MaxFiles = 5000
	}
	if req.MaxMatches <= 0 {
		req.MaxMatches = 500
	}
	if req.MaxMatchesPerFile <= 0 {
		req.MaxMatchesPerFile = 100
	}
	if req.ASTParentDepth <= 0 {
		req.ASTParentDepth = 3
	}
	if req.NodeTextLimitBytes <= 0 {
		req.NodeTextLimitBytes = 512
	}
}

// excludedDirs mirrors the indexer's excluded directories.
var excludedDirs = map[string]bool{
	".git": true, ".mesdx": true, "node_modules": true,
	".venv": true, "venv": true, ".env": true, "vendor": true,
	"target": true, "build": true, "dist": true,
	".idea": true, ".vscode": true,
	"__pycache__": true, ".mypy_cache": true,
}

func (e *Engine) enumerateFiles(req SearchRequest) ([]string, error) {
	roots := e.SourceRoots
	if len(req.SourceRoots) > 0 {
		roots = req.SourceRoots
	}

	var files []string
	for _, root := range roots {
		absRoot := root
		if root == "." {
			absRoot = e.RepoRoot
		} else if !filepath.IsAbs(root) {
			absRoot = filepath.Join(e.RepoRoot, root)
		}

		_ = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if info != nil && info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.IsDir() {
				if excludedDirs[info.Name()] || (strings.HasPrefix(info.Name(), ".") && info.Name() != ".") {
					return filepath.SkipDir
				}
				return nil
			}

			lang := indexer.DetectLang(path)
			if string(lang) != req.Language {
				return nil
			}

			relPath, _ := filepath.Rel(e.RepoRoot, path)
			if len(req.IncludeGlobs) > 0 && !matchesAny(relPath, req.IncludeGlobs) {
				return nil
			}
			if matchesAny(relPath, req.ExcludeGlobs) {
				return nil
			}

			files = append(files, path)
			return nil
		})
	}
	return files, nil
}

func matchesAny(path string, patterns []string) bool {
	for _, p := range patterns {
		if ok, _ := filepath.Match(p, path); ok {
			return true
		}
		if ok, _ := filepath.Match(p, filepath.Base(path)); ok {
			return true
		}
	}
	return false
}

func (e *Engine) searchFile(
	ctx context.Context,
	parser *treesitter.Parser,
	filePath, language, querySrc string,
	req SearchRequest,
) ([]Match, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lang, err := treesitter.LoadLanguage(language)
	if err != nil {
		return nil, err
	}

	if err := parser.SetLanguage(lang); err != nil {
		return nil, err
	}

	tree := parser.ParseString(nil, src)
	if tree == nil {
		return nil, fmt.Errorf("parse failed")
	}
	defer tree.Close()

	q := e.Cache.Get(language, querySrc)
	if q == nil {
		compiled, err := treesitter.NewQuery(lang, querySrc)
		if err != nil {
			return nil, fmt.Errorf("compile query: %w", err)
		}
		e.Cache.Put(language, querySrc, lang, compiled)
		q = compiled
	}

	cursor := treesitter.NewQueryCursor()
	defer cursor.Close()

	rootNode := tree.RootNode()
	cursor.ExecWithText(q, rootNode, src)

	captureNames := make(map[uint32]string)
	for i := uint32(0); i < q.CaptureCount(); i++ {
		captureNames[i] = q.CaptureNameForID(i)
	}

	lines := buildLineIndex(src)

	relPath, _ := filepath.Rel(e.RepoRoot, filePath)
	var matches []Match

	for ctx.Err() == nil {
		if len(matches) >= req.MaxMatchesPerFile {
			break
		}

		qm := cursor.NextMatch()
		if qm == nil {
			break
		}

		for _, cap := range qm.Captures {
			if len(matches) >= req.MaxMatchesPerFile {
				break
			}

			capName := captureNames[cap.Index]
			node := cap.Node
			if node.IsNull() {
				continue
			}

			startPoint := node.StartPoint()
			endPoint := node.EndPoint()

			text := node.Content(src)
			if req.NodeTextLimitBytes > 0 && len(text) > req.NodeTextLimitBytes {
				text = text[:req.NodeTextLimitBytes] + "..."
			}

			startLine := int(startPoint.Row) + 1
			endLine := int(endPoint.Row) + 1

			m := Match{
				FilePath:    relPath,
				StartLine:   startLine,
				StartCol:    int(startPoint.Column),
				EndLine:     endLine,
				EndCol:      int(endPoint.Column),
				CaptureName: capName,
				NodeType:    node.Type(),
				TextSnippet: text,
				Line:        getLine(lines, int(startPoint.Row)),
			}

			if req.ContextLines > 0 {
				m.ContextBefore = getLineRange(lines, int(startPoint.Row)-req.ContextLines, int(startPoint.Row)-1)
				m.ContextAfter = getLineRange(lines, int(endPoint.Row)+1, int(endPoint.Row)+req.ContextLines)
			}

			if req.ASTParentDepth > 0 {
				m.ASTParents = collectParents(node, req.ASTParentDepth)
			}

			matches = append(matches, m)
		}
	}

	return matches, nil
}

// buildLineIndex splits source into lines for O(1) access.
func buildLineIndex(src []byte) []string {
	s := string(src)
	return strings.Split(s, "\n")
}

func getLine(lines []string, row int) string {
	if row < 0 || row >= len(lines) {
		return ""
	}
	return lines[row]
}

func getLineRange(lines []string, from, to int) []string {
	if from < 0 {
		from = 0
	}
	if to >= len(lines) {
		to = len(lines) - 1
	}
	if from > to {
		return nil
	}
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, lines[i])
	}
	return out
}

func collectParents(node treesitter.Node, depth int) []string {
	var parents []string
	current := node.Parent()
	for i := 0; i < depth; i++ {
		if current.IsNull() {
			break
		}
		parents = append(parents, current.Type())
		current = current.Parent()
	}
	return parents
}
