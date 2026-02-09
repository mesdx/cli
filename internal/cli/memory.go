package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/codeintelx/cli/internal/config"
	"github.com/codeintelx/cli/internal/db"
	"github.com/codeintelx/cli/internal/indexer"
	"github.com/codeintelx/cli/internal/mcpstate"
	"github.com/codeintelx/cli/internal/memory"
	"github.com/codeintelx/cli/internal/repo"
	"github.com/spf13/cobra"
)

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Memory management commands",
		Long:  "Manage project memory elements (markdown-backed documentation and context)",
	}

	cmd.AddCommand(newMemorySearchCmd())
	cmd.AddCommand(newMemoryReindexCmd())

	return cmd
}

// checkMCPNotRunning checks if MCP server is running and returns an error if it is
func checkMCPNotRunning(codeintelxDir string) error {
	running, state, err := mcpstate.IsRunning(codeintelxDir)
	if err != nil {
		return fmt.Errorf("failed to check MCP state: %w", err)
	}
	if running {
		return fmt.Errorf("memory index is currently locked by MCP server (PID: %d, started: %s)\n\nThe MCP server has exclusive access to the memory index. Please:\n  1. Stop the MCP server, or\n  2. Wait for MCP operations to complete\n\nThen try again",
			state.PID, state.StartedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func newMemorySearchCmd() *cobra.Command {
	var (
		scope  string
		file   string
		limit  int
		format string
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search memory elements using full-text search",
		Long: `Search memory elements using Bleve full-text search.

Memory files are indexed in chunks split by any markdown header (# through ######).
Returns ranked results. Deleted memories and memories referencing deleted files are excluded.

Examples:
  codeintelx memory search "authentication JWT tokens"
  codeintelx memory search "database schema" --scope project
  codeintelx memory search "main function" --scope file --file src/main.go
  codeintelx memory search "API endpoints" --limit 5
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			// Find repo root and load config
			repoRoot, err := repo.FindRoot()
			if err != nil {
				return fmt.Errorf("failed to find repo root: %w", err)
			}

			codeintelxDir := repo.CodeintelxDir(repoRoot)

			// Check if MCP server is running
			if err := checkMCPNotRunning(codeintelxDir); err != nil {
				return err
			}

			cfg, err := config.Load(codeintelxDir)
			if err != nil {
				return fmt.Errorf("failed to load config: %w (run 'codeintelx init' first)", err)
			}

			// Open DB
			dbPath := db.DatabasePath(codeintelxDir)
			d, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer func() { _ = d.Close() }()

			// Get project ID
			idx := indexer.New(d, repoRoot)
			if err := idx.Store.EnsureProject(repoRoot); err != nil {
				return fmt.Errorf("failed to ensure project: %w", err)
			}

			// Create memory manager
			memoryDirAbs := cfg.MemoryDir
			if !filepath.IsAbs(memoryDirAbs) {
				memoryDirAbs = filepath.Join(repoRoot, memoryDirAbs)
			}

			mgr, err := memory.NewManagerReadOnly(d, idx.Store.ProjectID, repoRoot, memoryDirAbs)
			if err != nil {
				// Check if it's a lock error
				if memory.IsIndexLockedError(err) {
					return fmt.Errorf("memory index is locked by MCP server. Please wait for MCP operations to complete, or run 'codeintelx memory reindex' to rebuild the index")
				}
				return fmt.Errorf("failed to open memory index: %w", err)
			}
			defer func() { _ = mgr.Close() }()

			// Perform search
			results, err := mgr.Search(query, scope, file, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Display results
			if len(results) == 0 {
				cmd.Println("No results found.")
				return nil
			}

			switch format {
			case "json":
				return printSearchResultsJSON(cmd, results)
			default:
				return printSearchResultsTable(cmd, results)
			}
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Filter by scope: 'project' or 'file'")
	cmd.Flags().StringVar(&file, "file", "", "Filter by repo-relative file path (requires --scope file)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of results")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: 'table' or 'json'")

	return cmd
}

func printSearchResultsTable(cmd *cobra.Command, results []memory.SearchResult) error {
	cmd.Printf("Found %d result(s):\n\n", len(results))
	for i, r := range results {
		cmd.Printf("%d. ", i+1)
		if r.Title != "" {
			cmd.Printf("%s", successStyle.Render(r.Title))
		} else {
			cmd.Printf("%s", r.MemoryUID)
		}
		cmd.Printf(" (score: %.2f)\n", r.Score)

		cmd.Printf("   Scope: %s", r.Scope)
		if r.FilePath != "" {
			cmd.Printf(" | File: %s", r.FilePath)
		}
		cmd.Printf("\n")

		cmd.Printf("   Path: %s\n", infoStyle.Render(r.MdRelPath))

		if r.Snippet != "" {
			cmd.Printf("   Snippet: %s\n", r.Snippet)
		}
		cmd.Println()
	}
	return nil
}

func printSearchResultsJSON(cmd *cobra.Command, results []memory.SearchResult) error {
	data := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		data = append(data, map[string]interface{}{
			"memoryId": r.MemoryUID,
			"scope":    r.Scope,
			"filePath": r.FilePath,
			"mdPath":   r.MdRelPath,
			"title":    r.Title,
			"score":    r.Score,
			"snippet":  r.Snippet,
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func newMemoryReindexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Reindex and reconcile all memory elements",
		Long: `Reindex all memory markdown files and reconcile their state.

This command performs a full reindex of all memory elements and then reconciles
their status (checking file references, symbol references, etc.). This is useful
when:
  - The memory index is in an inconsistent state
  - The MCP server has locked the index and you need to rebuild it
  - Memory files have been manually edited or moved

Examples:
  codeintelx memory reindex
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find repo root and load config
			repoRoot, err := repo.FindRoot()
			if err != nil {
				return fmt.Errorf("failed to find repo root: %w", err)
			}

			codeintelxDir := repo.CodeintelxDir(repoRoot)

			// Check if MCP server is running
			if err := checkMCPNotRunning(codeintelxDir); err != nil {
				return err
			}

			cfg, err := config.Load(codeintelxDir)
			if err != nil {
				return fmt.Errorf("failed to load config: %w (run 'codeintelx init' first)", err)
			}

			// Open DB
			dbPath := db.DatabasePath(codeintelxDir)
			d, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer func() { _ = d.Close() }()

			// Get project ID
			idx := indexer.New(d, repoRoot)
			if err := idx.Store.EnsureProject(repoRoot); err != nil {
				return fmt.Errorf("failed to ensure project: %w", err)
			}

			// Create memory manager (read-write mode for reindexing)
			memoryDirAbs := cfg.MemoryDir
			if !filepath.IsAbs(memoryDirAbs) {
				memoryDirAbs = filepath.Join(repoRoot, memoryDirAbs)
			}

			cmd.Println("Creating memory manager...")
			mgr, err := memory.NewManager(d, idx.Store.ProjectID, repoRoot, memoryDirAbs)
			if err != nil {
				return fmt.Errorf("failed to create memory manager: %w", err)
			}
			defer func() { _ = mgr.Close() }()

			// Perform bulk reindex
			cmd.Println("Reindexing all memory elements...")
			if err := mgr.BulkIndex(); err != nil {
				return fmt.Errorf("failed to reindex: %w", err)
			}

			// Reconcile
			cmd.Println("Reconciling memory state...")
			if err := mgr.Reconcile(); err != nil {
				return fmt.Errorf("failed to reconcile: %w", err)
			}

			cmd.Println(successStyle.Render("✓ Memory reindex and reconciliation completed successfully"))
			return nil
		},
	}

	return cmd
}
