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
	"github.com/codeintelx/cli/internal/repo"
	"github.com/fsnotify/fsnotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// GoToDefArgs is the input for the goToDefinition MCP tool.
type GoToDefArgs struct {
	FilePath   string `json:"filePath,omitempty"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	SymbolName string `json:"symbolName,omitempty"`
	Language   string `json:"language,omitempty"`
}

// FindUsagesArgs is the input for the findUsages MCP tool.
type FindUsagesArgs struct {
	FilePath   string `json:"filePath,omitempty"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	SymbolName string `json:"symbolName,omitempty"`
	Language   string `json:"language,omitempty"`
}

func newMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server",
		Long:  "Start the Model Context Protocol server for Claude Code integration.",
		RunE:  runMcp,
	}

	return cmd
}

func runMcp(cmd *cobra.Command, args []string) error {
	// Find repo root and load config
	repoRoot, err := repo.FindRoot()
	if err != nil {
		return fmt.Errorf("failed to find repo root: %w", err)
	}

	codeintelxDir := repo.CodeintelxDir(repoRoot)

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
	defer d.Close()

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

	// Start file watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startFileWatcher(ctx, idx, cfg, repoRoot)

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
		Description: "Find the definition of a symbol. Provide either (filePath + line + column) for cursor-based lookup, or (symbolName) for name-based search. Returns definition locations with file path, line, column, kind, and signature.",
		InputSchema: mustSchema(GoToDefArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GoToDefArgs) (*mcp.CallToolResult, any, error) {
		var results []indexer.DefinitionResult
		var err error

		if args.SymbolName != "" {
			results, err = nav.GoToDefinitionByName(args.SymbolName, args.FilePath)
		} else if args.FilePath != "" && args.Line > 0 {
			// Convert absolute path to repo-relative if needed
			relPath := toRepoRelative(args.FilePath, repoRoot)
			results, err = nav.GoToDefinitionByPosition(relPath, args.Line, args.Column)
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
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, results, nil
	})

	// Register Find Usages tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.findUsages",
		Description: "Find all usage references of a symbol across the codebase. Provide either (filePath + line + column) for cursor-based lookup, or (symbolName) for name-based search. Returns reference locations with file path, line, column, and context.",
		InputSchema: mustSchema(FindUsagesArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args FindUsagesArgs) (*mcp.CallToolResult, any, error) {
		var results []indexer.UsageResult
		var err error

		if args.SymbolName != "" {
			results, err = nav.FindUsagesByName(args.SymbolName, args.FilePath)
		} else if args.FilePath != "" && args.Line > 0 {
			relPath := toRepoRelative(args.FilePath, repoRoot)
			results, err = nav.FindUsagesByPosition(relPath, args.Line, args.Column)
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

		text := indexer.FormatUsages(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, results, nil
	})

	// Run server over stdio
	return server.Run(context.Background(), &mcp.StdioTransport{})
}

// startFileWatcher watches source roots for file changes and incrementally updates the index.
func startFileWatcher(ctx context.Context, idx *indexer.Indexer, cfg *config.Config, repoRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("fsnotify: failed to create watcher: %v", err)
		return
	}
	defer watcher.Close()

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
			"description": "Optional language filter (go, java, rust, python, typescript, javascript)",
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
			"description": "Optional language filter (go, java, rust, python, typescript, javascript)",
		}
	}

	return schema
}
