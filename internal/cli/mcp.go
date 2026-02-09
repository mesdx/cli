package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codeintelx/cli/internal/config"
	"github.com/codeintelx/cli/internal/db"
	"github.com/codeintelx/cli/internal/indexer"
	"github.com/codeintelx/cli/internal/mcpstate"
	"github.com/codeintelx/cli/internal/memory"
	"github.com/codeintelx/cli/internal/repo"
	"github.com/fsnotify/fsnotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var supportedLanguages = map[string]bool{
	"go":         true,
	"java":       true,
	"rust":       true,
	"python":     true,
	"typescript": true,
	"javascript": true,
}

func validateLanguage(lang string) error {
	if lang == "" {
		return fmt.Errorf("language parameter is required")
	}
	if !supportedLanguages[lang] {
		return fmt.Errorf("unsupported language %q; supported: go, java, rust, python, typescript, javascript", lang)
	}
	return nil
}

// GoToDefArgs is the input for the goToDefinition MCP tool.
type GoToDefArgs struct {
	FilePath     string `json:"filePath,omitempty"`
	Line         int    `json:"line,omitempty"`
	Column       int    `json:"column,omitempty"`
	SymbolName   string `json:"symbolName,omitempty"`
	Language     string `json:"language,omitempty"`
	FetchTheCode *bool  `json:"fetchTheCode,omitempty"`
}

// FindUsagesArgs is the input for the findUsages MCP tool.
type FindUsagesArgs struct {
	FilePath             string `json:"filePath,omitempty"`
	Line                 int    `json:"line,omitempty"`
	Column               int    `json:"column,omitempty"`
	SymbolName           string `json:"symbolName,omitempty"`
	Language             string `json:"language,omitempty"`
	FetchCodeLinesAround int    `json:"fetchCodeLinesAround,omitempty"`
}

// DependencyGraphArgs is the input for the dependencyGraph MCP tool.
type DependencyGraphArgs struct {
	FilePath   string  `json:"filePath,omitempty"`
	Line       int     `json:"line,omitempty"`
	Column     int     `json:"column,omitempty"`
	SymbolName string  `json:"symbolName,omitempty"`
	Language   string  `json:"language,omitempty"`
	MaxDepth   int     `json:"maxDepth,omitempty"`
	MinScore   float64 `json:"minScore,omitempty"`
	MaxUsages  int     `json:"maxUsages,omitempty"`
}

func newMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server",
		Long:  "Start the Model Context Protocol server for Agentic code integration.",
		RunE:  runMcp,
	}

	cmd.Flags().String("cwd", "", "Working directory (defaults to current directory)")

	return cmd
}

func runMcp(cmd *cobra.Command, args []string) error {
	// Change working directory if --cwd is provided
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return fmt.Errorf("failed to get cwd flag: %w", err)
	}
	if cwd != "" {
		// Validate that the directory exists
		info, err := os.Stat(cwd)
		if err != nil {
			return fmt.Errorf("failed to access cwd directory %q: %w", cwd, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("cwd path %q is not a directory", cwd)
		}
		// Change to the specified directory
		if err := os.Chdir(cwd); err != nil {
			return fmt.Errorf("failed to change to directory %q: %w", cwd, err)
		}
	}

	// Find repo root and load config
	repoRoot, err := repo.FindRoot()
	if err != nil {
		return fmt.Errorf("failed to find repo root: %w", err)
	}

	codeintelxDir := repo.CodeintelxDir(repoRoot)

	// Create MCP state file to indicate server is running
	if err := mcpstate.CreateStateFile(codeintelxDir); err != nil {
		return fmt.Errorf("failed to create MCP state file: %w", err)
	}
	// Ensure state file is removed on exit
	defer func() {
		if err := mcpstate.RemoveStateFile(codeintelxDir); err != nil {
			log.Printf("warning: failed to remove MCP state file: %v", err)
		}
	}()

	// Redirect all logging to .codeintelx/mcp.log so nothing leaks into
	// the stdio JSON-RPC transport.
	if err := initMCPLog(codeintelxDir); err != nil {
		return fmt.Errorf("failed to initialize mcp log: %w", err)
	}

	cfg, err := config.Load(codeintelxDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Open and migrate DB
	dbPath := db.DatabasePath(codeintelxDir)
	d, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = d.Close() }()

	if err := db.Migrate(d); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create indexer and run initial reconcile
	idx := indexer.New(d, repoRoot)
	if err := idx.Store.EnsureProject(repoRoot); err != nil {
		return fmt.Errorf("failed to ensure project: %w", err)
	}

	stats, err := idx.Reconcile(cfg.SourceRoots)
	if err != nil {
		log.Printf("reconcile error: %v", err)
	} else if stats.Indexed > 0 || stats.Deleted > 0 || stats.Errors > 0 {
		log.Printf("reconcile: indexed=%d deleted=%d errors=%d", stats.Indexed, stats.Deleted, stats.Errors)
	}

	// Create navigator
	nav := &indexer.Navigator{
		DB:        d,
		ProjectID: idx.Store.ProjectID,
	}

	// Create memory manager (if memory dir is configured)
	var memMgr *memory.Manager
	if cfg.MemoryDir != "" {
		memDirAbs := cfg.MemoryDir
		if !filepath.IsAbs(memDirAbs) {
			memDirAbs = filepath.Join(repoRoot, memDirAbs)
		}
		var memErr error
		memMgr, memErr = memory.NewManager(d, idx.Store.ProjectID, repoRoot, memDirAbs)
		if memErr != nil {
			log.Printf("memory manager init error: %v", memErr)
			memMgr = nil
		}

		// Bulk-index existing memory files
		if memMgr != nil {
			if err := memMgr.BulkIndex(); err != nil {
				log.Printf("memory bulk index error: %v", err)
			}
			// Reconcile file/symbol statuses
			if err := memMgr.Reconcile(); err != nil {
				log.Printf("memory reconcile error: %v", err)
			}
		}
	}
	defer func() {
		if memMgr != nil {
			if err := memMgr.Close(); err != nil {
				log.Printf("failed to close memory manager: %v", err)
			}
		}
	}()

	// Start file watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startFileWatcher(ctx, idx, cfg, repoRoot, memMgr)

	// Start memory dir watcher in background (if configured)
	if memMgr != nil {
		go startMemoryWatcher(ctx, memMgr)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "codeintelx",
		Version: Version,
	}, nil)

	// Register project info tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.projectInfo",
		Description: "Get project information including repo root, source roots, and database path",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		info := map[string]interface{}{
			"repoRoot":    cfg.RepoRoot,
			"sourceRoots": cfg.SourceRoots,
			"dbPath":      dbPath,
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Repo Root: %s\nSource Roots: %v\nDatabase: %s",
						cfg.RepoRoot, cfg.SourceRoots, dbPath),
				},
			},
		}, info, nil
	})

	// Register Go To Definition tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.goToDefinition",
		Description: "Find the definition of a symbol. Provide either (filePath + line + column) for cursor-based lookup, or (symbolName) for name-based search. Returns definition locations with file path, line, column, kind, signature, and optionally the code. The language parameter is required.",
		InputSchema: mustSchema(GoToDefArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GoToDefArgs) (*mcp.CallToolResult, any, error) {
		// Validate language
		if err := validateLanguage(args.Language); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		var results []indexer.DefinitionResult
		var err error

		if args.SymbolName != "" {
			results, err = nav.GoToDefinitionByName(args.SymbolName, args.FilePath, args.Language)
		} else if args.FilePath != "" && args.Line > 0 {
			// Convert absolute path to repo-relative if needed
			relPath := toRepoRelative(args.FilePath, repoRoot)
			// Check if file's language matches requested language
			detectedLang := indexer.DetectLang(relPath)
			if string(detectedLang) != args.Language {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: no identifier found at %s:%d:%d", relPath, args.Line, args.Column)},
					},
					IsError: true,
				}, nil, nil
			}
			results, err = nav.GoToDefinitionByPosition(relPath, args.Line, args.Column, args.Language)
		} else {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Error: provide either symbolName, or filePath + line + column",
					},
				},
				IsError: true,
			}, nil, nil
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		text := indexer.FormatDefinitions(results)
		structuredContent := map[string]interface{}{
			"definitions": results,
		}

		// Handle fetchTheCode (default true)
		fetchTheCode := args.FetchTheCode == nil || *args.FetchTheCode
		if fetchTheCode && len(results) > 0 {
			code, err := fetchDefinitionsCode(repoRoot, results)
			if err != nil {
				log.Printf("failed to fetch code for definitions: %v", err)
			} else if code != "" {
				structuredContent["code"] = code
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, structuredContent, nil
	})

	// Register Find Usages tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.findUsages",
		Description: "Find all usage references of a symbol across the codebase. Provide either (filePath + line + column) for cursor-based lookup, or (symbolName) for name-based search. Returns reference locations with file path, line, column, context, and a dependencyScore (0-1) indicating confidence that the usage truly depends on the intended definition. Results are sorted by score (descending) while keeping adjacent usages grouped. The language parameter is required. For fetchCodeLinesAround: prefer 0 (or more) for better context; use -1 only when context is limited.",
		InputSchema: mustSchema(FindUsagesArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindUsagesArgs) (*mcp.CallToolResult, any, error) {
		// Validate language
		if err := validateLanguage(args.Language); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		var results []indexer.UsageResult
		var err error
		var symbolName string
		var filterFile string
		var primaryDef *indexer.DefinitionResult

		if args.SymbolName != "" {
			symbolName = args.SymbolName
			filterFile = args.FilePath

			// Ambiguity check: name-based lookups must resolve to exactly one definition.
			ambigCandidates, ambigErr := nav.GoToDefinitionByName(args.SymbolName, args.FilePath, args.Language)
			if ambigErr != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: %v", ambigErr)},
					},
					IsError: true,
				}, nil, nil
			}
			if len(ambigCandidates) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: no definitions found for %q", args.SymbolName)},
					},
					IsError: true,
				}, nil, nil
			}
			if len(ambigCandidates) > 1 {
				text := formatAmbiguousFindUsages(args.SymbolName, ambigCandidates)
				return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: text},
						},
						IsError: true,
					}, map[string]interface{}{
						"ambiguous":            true,
						"definitionCandidates": ambigCandidates,
						"hint":                 "Supply a filePath filter or use cursor-based lookup (filePath + line + column) to disambiguate.",
					}, nil
			}
			primaryDef = &ambigCandidates[0]

			results, err = nav.FindUsagesByName(args.SymbolName, args.FilePath, args.Language)
		} else if args.FilePath != "" && args.Line > 0 {
			relPath := toRepoRelative(args.FilePath, repoRoot)
			filterFile = relPath
			// Check if file's language matches requested language
			detectedLang := indexer.DetectLang(relPath)
			if string(detectedLang) != args.Language {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: no identifier found at %s:%d:%d", relPath, args.Line, args.Column)},
					},
					IsError: true,
				}, nil, nil
			}
			results, err = nav.FindUsagesByPosition(relPath, args.Line, args.Column, args.Language)
			// Resolve primary definition for cursor-based lookup.
			if err == nil && len(results) > 0 {
				symbolName = results[0].Name
				defs, defErr := nav.GoToDefinitionByPosition(relPath, args.Line, args.Column, args.Language)
				if defErr == nil && len(defs) > 0 {
					primaryDef = &defs[0]
				}
			}
		} else {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Error: provide either symbolName, or filePath + line + column",
					},
				},
				IsError: true,
			}, nil, nil
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		// Look up candidate definitions for scoring.
		candidates, _ := nav.GoToDefinitionByName(symbolName, filterFile, args.Language)

		// If no primary def was set (name-based lookup), use the top candidate.
		if primaryDef == nil && len(candidates) > 0 {
			primaryDef = &candidates[0]
		}

		// Score usages against candidate definitions.
		scored := indexer.ScoreUsages(results, candidates, primaryDef, repoRoot)

		// Group adjacent usages and sort by score descending.
		scored = indexer.GroupAndSortUsages(scored, 3)

		// Write scores back to UsageResult for formatting/output.
		scoredResults := make([]indexer.UsageResult, len(scored))
		for i, su := range scored {
			scoredResults[i] = su.UsageResult
			scoredResults[i].DependencyScore = su.DependencyScore
		}

		text := indexer.FormatScoredUsages(scored)
		structuredContent := map[string]interface{}{
			"usages":               scored,
			"primaryDefinition":    primaryDef,
			"definitionCandidates": candidates,
		}

		// Handle fetchCodeLinesAround (default -1, clamped to [-1, 50])
		linesAround := args.FetchCodeLinesAround
		if linesAround < -1 {
			linesAround = -1
		} else if linesAround > 50 {
			linesAround = 50
		}

		if linesAround >= 0 && len(scoredResults) > 0 {
			code, err := fetchUsagesCode(repoRoot, scoredResults, linesAround)
			if err != nil {
				log.Printf("failed to fetch code for usages: %v", err)
			} else if code != "" {
				structuredContent["code"] = code
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, structuredContent, nil
	})

	// Register Dependency Graph tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.dependencyGraph",
		Description: "Extract a dependency graph for a symbol, showing inbound usages (who depends on it) and outbound dependencies (what it depends on). Returns a symbol-level graph, a collapsed file-level graph, scored usages, and candidate definitions. Use this for risk analysis when renaming or removing a symbol. The language parameter is required.",
		InputSchema: mustSchema(DependencyGraphArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DependencyGraphArgs) (*mcp.CallToolResult, any, error) {
		// Validate language
		if err := validateLanguage(args.Language); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		// Resolve symbol name
		var symbolName string
		var filterFile string
		var primaryDef *indexer.DefinitionResult

		if args.SymbolName != "" {
			symbolName = args.SymbolName
			filterFile = args.FilePath
		} else if args.FilePath != "" && args.Line > 0 {
			relPath := toRepoRelative(args.FilePath, repoRoot)
			filterFile = relPath
			detectedLang := indexer.DetectLang(relPath)
			if string(detectedLang) != args.Language {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: no identifier found at %s:%d:%d", relPath, args.Line, args.Column)},
					},
					IsError: true,
				}, nil, nil
			}
			defs, err := nav.GoToDefinitionByPosition(relPath, args.Line, args.Column, args.Language)
			if err != nil || len(defs) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error: no definition found at %s:%d:%d", relPath, args.Line, args.Column)},
					},
					IsError: true,
				}, nil, nil
			}
			symbolName = defs[0].Name
			primaryDef = &defs[0]
		} else {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Error: provide either symbolName, or filePath + line + column",
					},
				},
				IsError: true,
			}, nil, nil
		}

		// Look up candidate definitions
		candidates, err := nav.GoToDefinitionByName(symbolName, filterFile, args.Language)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		if primaryDef == nil && len(candidates) > 0 {
			primaryDef = &candidates[0]
		}

		// Apply defaults and caps
		maxDepth := args.MaxDepth
		if maxDepth <= 0 {
			maxDepth = 1
		}
		if maxDepth > 2 {
			maxDepth = 2
		}
		minScore := args.MinScore
		if minScore <= 0 {
			minScore = 0.2
		}
		maxUsages := args.MaxUsages
		if maxUsages <= 0 {
			maxUsages = 500
		}

		// Build the dependency graph
		graph, err := indexer.BuildDependencyGraph(
			nav, primaryDef, candidates, args.Language,
			repoRoot, maxDepth, minScore, maxUsages,
		)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		// Format human-readable summary
		text := formatDependencyGraph(graph)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, graph, nil
	})

	// Register memory tools (if memory dir is configured)
	if memMgr != nil {
		registerMemoryTools(server, memMgr)
	}

	// Run server over stdio
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// startFileWatcher watches source roots for file changes and incrementally updates the index.
func startFileWatcher(ctx context.Context, idx *indexer.Indexer, cfg *config.Config, repoRoot string, memMgr *memory.Manager) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("fsnotify: failed to create watcher: %v", err)
		return
	}
	defer func() { _ = watcher.Close() }()

	// Add source root directories (recursively)
	for _, root := range cfg.SourceRoots {
		absRoot := root
		if root == "." {
			absRoot = repoRoot
		} else if !filepath.IsAbs(root) {
			absRoot = filepath.Join(repoRoot, root)
		}
		if err := addWatchRecursive(watcher, absRoot); err != nil {
			log.Printf("fsnotify: failed to watch %s: %v", absRoot, err)
		}
	}

	log.Printf("watcher started")

	// Debounce timer
	var mu sync.Mutex
	pending := map[string]struct{}{}
	var timer *time.Timer

	flush := func() {
		mu.Lock()
		files := pending
		pending = map[string]struct{}{}
		mu.Unlock()

		for path := range files {
			info, err := os.Stat(path)
			if err != nil {
				// File was deleted
				if err := idx.RemoveSingleFile(path); err != nil {
					log.Printf("failed to remove %s: %v", path, err)
				}
				if memMgr != nil {
					if rel, err := filepath.Rel(repoRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
						_ = memMgr.ReconcileFileRef(filepath.ToSlash(rel))
					}
				}
				continue
			}
			if info.IsDir() {
				// New directory — add to watcher
				_ = addWatchRecursive(watcher, path)
				continue
			}
			if _, err := idx.IndexSingleFile(path); err != nil {
				log.Printf("failed to index %s: %v", path, err)
			}
			if memMgr != nil {
				if rel, err := filepath.Rel(repoRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
					_ = memMgr.ReconcileFileRef(filepath.ToSlash(rel))
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			mu.Lock()
			pending[event.Name] = struct{}{}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(200*time.Millisecond, flush)
			mu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// addWatchRecursive adds a directory and all its subdirectories to the watcher.
func addWatchRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" ||
			name == "target" || name == "build" || name == "dist" || name == "__pycache__" {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

const mcpLogFileName = "mcp.log"

// initMCPLog opens (or creates) .codeintelx/mcp.log and redirects the
// standard log package output there. The file is truncated on each startup
// so it never grows unbounded between runs.
func initMCPLog(codeintelxDir string) error {
	logPath := filepath.Join(codeintelxDir, mcpLogFileName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("mcp server starting (log: %s)", logPath)
	return nil
}

// formatDependencyGraph formats a dependency graph as Mermaid syntax.
func formatDependencyGraph(g *indexer.DependencyGraph) string {
	var b strings.Builder

	// Summary section
	if g.PrimaryDefinition != nil {
		fmt.Fprintf(&b, "# Dependency Graph for %s (%s)\n\n", g.PrimaryDefinition.Name, g.PrimaryDefinition.Kind)
		fmt.Fprintf(&b, "**Location**: %s:%d:%d\n", g.PrimaryDefinition.Location.Path,
			g.PrimaryDefinition.Location.StartLine, g.PrimaryDefinition.Location.StartCol)
		if g.PrimaryDefinition.Signature != "" {
			fmt.Fprintf(&b, "**Signature**: `%s`\n", g.PrimaryDefinition.Signature)
		}
	} else {
		b.WriteString("# Dependency Graph\n\n")
	}

	fmt.Fprintf(&b, "\n**Definition Candidates**: %d\n", len(g.DefinitionCandidates))
	for i, d := range g.DefinitionCandidates {
		if i >= 5 {
			fmt.Fprintf(&b, "- ... and %d more\n", len(g.DefinitionCandidates)-5)
			break
		}
		fmt.Fprintf(&b, "- [%d] %s (%s) at %s:%d\n", i+1, d.Name, d.Kind,
			d.Location.Path, d.Location.StartLine)
	}

	// Symbol graph stats
	inboundCount := 0
	outboundCount := 0
	for _, e := range g.SymbolGraph.Edges {
		if e.Kind == "inbound" {
			inboundCount++
		} else {
			outboundCount++
		}
	}

	// Mermaid graph for file dependencies
	if len(g.FileGraph) > 0 {
		b.WriteString("\n## File Dependency Graph\n\n")
		b.WriteString("```mermaid\n")
		b.WriteString("graph LR\n")

		// Sanitize file paths for Mermaid node IDs
		fileToID := make(map[string]string)
		idCounter := 0
		getFileID := func(path string) string {
			if id, ok := fileToID[path]; ok {
				return id
			}
			id := fmt.Sprintf("F%d", idCounter)
			idCounter++
			fileToID[path] = id
			return id
		}

		// Define nodes with shortened labels
		nodesDefined := make(map[string]bool)
		for _, fe := range g.FileGraph {
			fromID := getFileID(fe.From)
			toID := getFileID(fe.To)
			if !nodesDefined[fromID] {
				fmt.Fprintf(&b, "    %s[\"%s\"]\n", fromID, shortenPath(fe.From, 40))
				nodesDefined[fromID] = true
			}
			if !nodesDefined[toID] {
				fmt.Fprintf(&b, "    %s[\"%s\"]\n", toID, shortenPath(fe.To, 40))
				nodesDefined[toID] = true
			}
		}

		// Draw edges with labels
		for _, fe := range g.FileGraph {
			fromID := getFileID(fe.From)
			toID := getFileID(fe.To)
			fmt.Fprintf(&b, "    %s -->|\"%.2f (%d)\"| %s\n", fromID, fe.Score, fe.Count, toID)
		}

		b.WriteString("```\n")
	}

	// Mermaid graph for symbol dependencies (top N only for readability)
	if len(g.SymbolGraph.Edges) > 0 {
		b.WriteString("\n## Symbol Dependency Graph\n\n")
		fmt.Fprintf(&b, "**Nodes**: %d | **Edges**: %d (%d inbound, %d outbound)\n\n",
			len(g.SymbolGraph.Nodes), len(g.SymbolGraph.Edges), inboundCount, outboundCount)

		b.WriteString("```mermaid\n")
		b.WriteString("graph TD\n")

		// Define primary node
		if g.PrimaryDefinition != nil {
			primaryID := sanitizeMermaidID(g.PrimaryDefinition.Name + "@" + g.PrimaryDefinition.Location.Path)
			fmt.Fprintf(&b, "    %s[[\"%s (%s)\"]]\n", primaryID, g.PrimaryDefinition.Name, g.PrimaryDefinition.Kind)
			fmt.Fprintf(&b, "    style %s fill:#f9f,stroke:#333,stroke-width:3px\n", primaryID)

			// Inbound edges (limit to top 15 by score)
			inboundEdges := []indexer.DepGraphEdge{}
			for _, e := range g.SymbolGraph.Edges {
				if e.Kind == "inbound" {
					inboundEdges = append(inboundEdges, e)
				}
			}
			// Sort by score descending
			for i := 0; i < len(inboundEdges); i++ {
				for j := i + 1; j < len(inboundEdges); j++ {
					if inboundEdges[j].Score > inboundEdges[i].Score {
						inboundEdges[i], inboundEdges[j] = inboundEdges[j], inboundEdges[i]
					}
				}
			}
			displayLimit := 15
			for i, e := range inboundEdges {
				if i >= displayLimit {
					if len(inboundEdges) > displayLimit {
						moreID := sanitizeMermaidID("more_inbound")
						fmt.Fprintf(&b, "    %s[\"... %d more files\"]\n", moreID, len(inboundEdges)-displayLimit)
						fmt.Fprintf(&b, "    %s -.-> %s\n", moreID, primaryID)
					}
					break
				}
				fromID := sanitizeMermaidID(e.FilePath)
				fmt.Fprintf(&b, "    %s[\"%s\"]\n", fromID, shortenPath(e.FilePath, 30))
				fmt.Fprintf(&b, "    %s -->|\"%.2f\"| %s\n", fromID, e.Score, primaryID)
			}

			// Outbound edges (limit to top 15)
			outboundEdges := []indexer.DepGraphEdge{}
			for _, e := range g.SymbolGraph.Edges {
				if e.Kind == "outbound" {
					outboundEdges = append(outboundEdges, e)
				}
			}
			for i, e := range outboundEdges {
				if i >= displayLimit {
					if len(outboundEdges) > displayLimit {
						moreID := sanitizeMermaidID("more_outbound")
						fmt.Fprintf(&b, "    %s[\"... %d more symbols\"]\n", moreID, len(outboundEdges)-displayLimit)
						fmt.Fprintf(&b, "    %s -.-> %s\n", primaryID, moreID)
					}
					break
				}
				// Parse To as "path:name:line"
				toID := sanitizeMermaidID(e.To)
				// Extract symbol name from node ID
				symbolName := extractSymbolFromNodeID(e.To)
				fmt.Fprintf(&b, "    %s[\"%s\"]\n", toID, symbolName)
				fmt.Fprintf(&b, "    %s -->|\"%.2f\"| %s\n", primaryID, e.Score, toID)
			}
		}

		b.WriteString("```\n")
	}

	// Scored usages summary
	if len(g.Usages) > 0 {
		b.WriteString("\n## Top Scored Usages\n\n")
		fmt.Fprintf(&b, "**Total**: %d usages\n\n", len(g.Usages))
		for i, u := range g.Usages {
			if i >= 20 {
				fmt.Fprintf(&b, "\n*... and %d more usages*\n", len(g.Usages)-20)
				break
			}
			fmt.Fprintf(&b, "%d. **%s** at `%s:%d` (score: %.4f)\n", i+1, u.Name,
				u.Location.Path, u.Location.StartLine, u.DependencyScore)
			if u.ContextContainer != "" {
				fmt.Fprintf(&b, "   - Context: %s\n", u.ContextContainer)
			}
		}
	}

	return b.String()
}

// shortenPath shortens a file path for display (keeps start and end if too long).
func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	// Keep first 15 and last 20 chars with "..." in between
	start := maxLen / 3
	end := maxLen - start - 3
	if end <= 0 {
		return path[:maxLen]
	}
	return path[:start] + "..." + path[len(path)-end:]
}

// sanitizeMermaidID converts a string into a valid Mermaid node ID.
func sanitizeMermaidID(s string) string {
	// Replace special chars with underscores
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	id := string(result)
	// Ensure it starts with a letter
	if len(id) > 0 && id[0] >= '0' && id[0] <= '9' {
		id = "n" + id
	}
	if len(id) == 0 {
		id = "node"
	}
	return id
}

// extractSymbolFromNodeID extracts the symbol name from a node ID "path:name:line".
func extractSymbolFromNodeID(nodeID string) string {
	// Find the last two colons
	lastColon := -1
	secondLastColon := -1
	for i := len(nodeID) - 1; i >= 0; i-- {
		if nodeID[i] == ':' {
			if lastColon == -1 {
				lastColon = i
			} else if secondLastColon == -1 {
				secondLastColon = i
				break
			}
		}
	}
	if secondLastColon >= 0 && lastColon > secondLastColon {
		return nodeID[secondLastColon+1 : lastColon]
	}
	return nodeID
}

// toRepoRelative converts an absolute path to a repo-relative path.
func toRepoRelative(path, repoRoot string) string {
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(repoRoot, path); err == nil {
			return rel
		}
	}
	return path
}

// mustSchema builds a JSON Schema from the struct using reflection.
func mustSchema(v interface{}) json.RawMessage {
	t := fmt.Sprintf("%T", v)
	_ = t
	// Simple schema generation based on struct tags.
	schema := buildSchema(v)
	data, _ := json.Marshal(schema)
	return data
}

// buildSchema creates a minimal JSON Schema from a struct's json tags.
func buildSchema(v interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{"language"},
	}

	// Use reflect to inspect the struct
	props := schema["properties"].(map[string]interface{})

	switch v.(type) {
	case GoToDefArgs:
		props["filePath"] = map[string]interface{}{
			"type":        "string",
			"description": "Path to the source file (absolute or repo-relative)",
		}
		props["line"] = map[string]interface{}{
			"type":        "integer",
			"description": "1-based line number of the cursor position",
		}
		props["column"] = map[string]interface{}{
			"type":        "integer",
			"description": "0-based column number of the cursor position",
		}
		props["symbolName"] = map[string]interface{}{
			"type":        "string",
			"description": "Name of the symbol to look up (alternative to cursor-based lookup)",
		}
		props["language"] = map[string]interface{}{
			"type":        "string",
			"description": "Programming language filter (required): go, java, rust, python, typescript, javascript",
		}
		props["fetchTheCode"] = map[string]interface{}{
			"type":        "boolean",
			"description": "Whether to include the definition source code in the response (default: true)",
			"default":     true,
		}
	case FindUsagesArgs:
		props["filePath"] = map[string]interface{}{
			"type":        "string",
			"description": "Path to the source file (absolute or repo-relative)",
		}
		props["line"] = map[string]interface{}{
			"type":        "integer",
			"description": "1-based line number of the cursor position",
		}
		props["column"] = map[string]interface{}{
			"type":        "integer",
			"description": "0-based column number of the cursor position",
		}
		props["symbolName"] = map[string]interface{}{
			"type":        "string",
			"description": "Name of the symbol to look up (alternative to cursor-based lookup)",
		}
		props["language"] = map[string]interface{}{
			"type":        "string",
			"description": "Programming language filter (required): go, java, rust, python, typescript, javascript",
		}
		props["fetchCodeLinesAround"] = map[string]interface{}{
			"type":        "integer",
			"description": "Number of context lines around each usage: -1=no code, 0=usage only, N>0=N lines before+after (default: -1, max: 50). Prefer 0+ for better context.",
			"default":     -1,
			"minimum":     -1,
			"maximum":     50,
		}
	case DependencyGraphArgs:
		props["filePath"] = map[string]interface{}{
			"type":        "string",
			"description": "Path to the source file (absolute or repo-relative)",
		}
		props["line"] = map[string]interface{}{
			"type":        "integer",
			"description": "1-based line number of the cursor position",
		}
		props["column"] = map[string]interface{}{
			"type":        "integer",
			"description": "0-based column number of the cursor position",
		}
		props["symbolName"] = map[string]interface{}{
			"type":        "string",
			"description": "Name of the symbol to look up (alternative to cursor-based lookup)",
		}
		props["language"] = map[string]interface{}{
			"type":        "string",
			"description": "Programming language filter (required): go, java, rust, python, typescript, javascript",
		}
		props["maxDepth"] = map[string]interface{}{
			"type":        "integer",
			"description": "Maximum depth for outbound dependency traversal (default: 1, max: 2)",
			"default":     1,
			"minimum":     0,
			"maximum":     2,
		}
		props["minScore"] = map[string]interface{}{
			"type":        "number",
			"description": "Minimum dependency score threshold for including usages (default: 0.2)",
			"default":     0.2,
			"minimum":     0,
			"maximum":     1,
		}
		props["maxUsages"] = map[string]interface{}{
			"type":        "integer",
			"description": "Maximum number of usages to include (default: 500)",
			"default":     500,
			"minimum":     1,
			"maximum":     5000,
		}
	}

	return schema
}

// startMemoryWatcher watches the memory directory for markdown file changes
// and incrementally updates the memory index.
func startMemoryWatcher(ctx context.Context, mgr *memory.Manager) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("memory fsnotify: failed to create watcher: %v", err)
		return
	}
	defer func() { _ = watcher.Close() }()

	if err := addWatchRecursive(watcher, mgr.MemoryDir); err != nil {
		log.Printf("memory fsnotify: failed to watch %s: %v", mgr.MemoryDir, err)
		return
	}

	log.Printf("memory watcher started for %s", mgr.MemoryDir)

	var mu sync.Mutex
	pending := map[string]struct{}{}
	var timer *time.Timer

	flush := func() {
		mu.Lock()
		files := pending
		pending = map[string]struct{}{}
		mu.Unlock()

		for path := range files {
			if !strings.HasSuffix(strings.ToLower(path), ".md") {
				continue
			}
			info, statErr := os.Stat(path)
			if statErr != nil {
				// File was deleted
				if err := mgr.RemoveFile(path); err != nil {
					log.Printf("memory: failed to remove %s: %v", path, err)
				}
				continue
			}
			if info.IsDir() {
				_ = addWatchRecursive(watcher, path)
				continue
			}
			if err := mgr.IndexFile(path); err != nil {
				log.Printf("memory: failed to index %s: %v", path, err)
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			mu.Lock()
			pending[event.Name] = struct{}{}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(200*time.Millisecond, flush)
			mu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("memory watcher error: %v", err)
		}
	}
}

// formatAmbiguousFindUsages formats an error message for ambiguous name-based findUsages.
func formatAmbiguousFindUsages(symbolName string, candidates []indexer.DefinitionResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: ambiguous symbol %q — %d definition candidates found. Provide a filePath filter or use cursor-based lookup (filePath + line + column) to disambiguate.\n\n", symbolName, len(candidates))
	for i, d := range candidates {
		fmt.Fprintf(&b, "%d. %s (%s) at %s:%d\n", i+1, d.Name, d.Kind, d.Location.Path, d.Location.StartLine)
		if d.Signature != "" {
			fmt.Fprintf(&b, "   Signature: %s\n", d.Signature)
		}
	}
	return b.String()
}
