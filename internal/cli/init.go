package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/codeintelx/cli/internal/config"
	"github.com/codeintelx/cli/internal/db"
	"github.com/codeintelx/cli/internal/ignore"
	"github.com/codeintelx/cli/internal/indexer"
	"github.com/codeintelx/cli/internal/memory"
	"github.com/codeintelx/cli/internal/repo"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize codeintelx in the current repository",
		Long:  "Initialize codeintelx by selecting source directories and setting up the index database.",
		RunE:  runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Detect repo root
	repoRoot, err := repo.FindRoot()
	if err != nil {
		return fmt.Errorf("failed to find repo root: %w", err)
	}

	cmd.Printf("%s Initializing codeintelx in: %s\n", infoStyle.Render("→"), repoRoot)

	// Discover all directories recursively for selection
	allDirs, err := repo.DiscoverAllDirs(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to discover directories: %w", err)
	}

	// Build options list: include root directory and all subdirectories
	var options []huh.Option[string]

	// Add root directory option (represented as ".")
	options = append(options, huh.NewOption(". (repository root)", "."))

	// Add all discovered directories
	for _, dir := range allDirs {
		// Display with forward slashes for consistency (even on Windows)
		displayPath := strings.ReplaceAll(dir, string(filepath.Separator), "/")
		options = append(options, huh.NewOption(displayPath, dir))
	}

	if len(options) == 0 {
		cmd.Println("No directories found to index.")
		return nil
	}

	// Interactive multi-select for directories
	var selectedDirs []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select directories to index").
				Description("Use arrow keys to navigate, space to select, enter to confirm. Note: Selected directories cannot be parent/child of each other.").
				Options(options...).
				Value(&selectedDirs),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("interactive prompt failed: %w", err)
	}

	if len(selectedDirs) == 0 {
		cmd.Println("No directories selected. Exiting.")
		return nil
	}

	// Validate selected directories: no duplicates, no parent/child relationships
	if err := repo.ValidateSelectedDirs(repoRoot, selectedDirs); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Prompt for memory directory
	memoryDir := config.DefaultMemoryDir
	memoryDirForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Memory directory").
				Description("Where should codeintelx store markdown memory files? (repo-relative, committed to VCS)").
				Placeholder(config.DefaultMemoryDir).
				Value(&memoryDir),
		),
	)
	if err := memoryDirForm.Run(); err != nil {
		return fmt.Errorf("interactive prompt failed: %w", err)
	}
	if memoryDir == "" {
		memoryDir = config.DefaultMemoryDir
	}
	// Validate: not inside .codeintelx
	if strings.HasPrefix(memoryDir, ".codeintelx") {
		return fmt.Errorf("memory directory cannot be inside .codeintelx/")
	}

	// Create memory directory
	memoryDirAbs := filepath.Join(repoRoot, memoryDir)
	if err := os.MkdirAll(memoryDirAbs, 0755); err != nil {
		return fmt.Errorf("failed to create memory directory: %w", err)
	}
	cmd.Printf("%s Memory directory: %s\n", successStyle.Render("✓"), memoryDir)

	// Remove existing .codeintelx directory and recreate (bulk replace)
	codeintelxDir := repo.CodeintelxDir(repoRoot)
	if err := os.RemoveAll(codeintelxDir); err != nil {
		return fmt.Errorf("failed to remove existing .codeintelx directory: %w", err)
	}
	if err := os.MkdirAll(codeintelxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .codeintelx directory: %w", err)
	}

	// Save config
	cfg := &config.Config{
		RepoRoot:    repoRoot,
		SourceRoots: selectedDirs,
		MemoryDir:   memoryDir,
	}
	if err := config.Save(cfg, codeintelxDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("%s Configuration saved to: %s\n", successStyle.Render("✓"), config.ConfigPath(codeintelxDir))

	// Initialize database
	dbPath := db.DatabasePath(codeintelxDir)
	if err := db.Initialize(dbPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	cmd.Printf("%s Database initialized at: %s\n", successStyle.Render("✓"), dbPath)

	// Run bulk indexing on source files
	cmd.Printf("%s Indexing source files...\n", infoStyle.Render("→"))

	d, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = d.Close() }()

	idx := indexer.New(d, repoRoot)
	stats, err := idx.FullIndex(selectedDirs)
	if err != nil {
		return fmt.Errorf("failed to index: %w", err)
	}

	cmd.Printf("%s Indexed %d files (%d symbols, %d references)\n",
		successStyle.Render("✓"), stats.Indexed, stats.Symbols, stats.Refs)
	if stats.Errors > 0 {
		cmd.Printf("%s %d files had errors during indexing\n", infoStyle.Render("!"), stats.Errors)
	}

	// Bulk-index existing memory files (if any exist in the memory dir)
	cmd.Printf("%s Indexing memory files...\n", infoStyle.Render("→"))
	memMgr, err := memory.NewManager(d, idx.Store.ProjectID, repoRoot, memoryDirAbs)
	if err != nil {
		return fmt.Errorf("failed to initialize memory manager: %w", err)
	}
	defer func() { _ = memMgr.Close() }()
	if err := memMgr.BulkIndex(); err != nil {
		cmd.Printf("%s Warning: failed to bulk-index memories: %v\n", infoStyle.Render("!"), err)
	}
	// Run initial reconcile on memories (file/symbol status)
	if err := memMgr.Reconcile(); err != nil {
		cmd.Printf("%s Warning: failed to reconcile memories: %v\n", infoStyle.Render("!"), err)
	}
	cmd.Printf("%s Memory indexing complete\n", successStyle.Render("✓"))

	// Handle .gitignore and .dockerignore prompts
	if err := ignore.HandleIgnoreFiles(repoRoot, cmd); err != nil {
		// Non-fatal, just log
		cmd.Printf("%s Warning: failed to update ignore files: %v\n", infoStyle.Render("!"), err)
	}

	cmd.Printf("\n%s Initialization complete!\n", successStyle.Render("✓"))
	cmd.Println("Next steps:")
	cmd.Println("  - Run 'codeintelx mcp' to start the MCP server")
	cmd.Println("  - Configure Claude Code to use this MCP server")

	return nil
}
