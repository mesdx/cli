package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mesdx/cli/internal/config"
	"github.com/mesdx/cli/internal/db"
	"github.com/mesdx/cli/internal/ignore"
	"github.com/mesdx/cli/internal/indexer"
	"github.com/mesdx/cli/internal/memory"
	"github.com/mesdx/cli/internal/repo"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize mesdx in the current repository",
		Long:  "Initialize mesdx by selecting source directories and setting up the index database.",
		RunE:  runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Verify parser libraries are available
	if err := indexer.VerifyParsersAvailable(); err != nil {
		return err
	}

	// Detect repo root
	repoRoot, err := repo.FindRoot()
	if err != nil {
		return fmt.Errorf("failed to find repo root: %w", err)
	}

	cmd.Printf("%s Initializing mesdx in: %s\n", infoStyle.Render("→"), repoRoot)

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
				Description("Where should mesdx store markdown memory files? (repo-relative, committed to VCS)").
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
	// Validate: not inside .mesdx
	if strings.HasPrefix(memoryDir, ".mesdx") {
		return fmt.Errorf("memory directory cannot be inside .mesdx/")
	}

	// Create memory directory
	memoryDirAbs := filepath.Join(repoRoot, memoryDir)
	if err := os.MkdirAll(memoryDirAbs, 0755); err != nil {
		return fmt.Errorf("failed to create memory directory: %w", err)
	}
	cmd.Printf("%s Memory directory: %s\n", successStyle.Render("✓"), memoryDir)

	// Remove existing .mesdx directory and recreate (bulk replace)
	mesdxDir := repo.MesdxDir(repoRoot)
	if err := os.RemoveAll(mesdxDir); err != nil {
		return fmt.Errorf("failed to remove existing .mesdx directory: %w", err)
	}
	if err := os.MkdirAll(mesdxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mesdx directory: %w", err)
	}

	// Save config
	cfg := &config.Config{
		RepoRoot:    repoRoot,
		SourceRoots: selectedDirs,
		MemoryDir:   memoryDir,
	}
	if err := config.Save(cfg, mesdxDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("%s Configuration saved to: %s\n", successStyle.Render("✓"), config.ConfigPath(mesdxDir))

	// Initialize database
	dbPath := db.DatabasePath(mesdxDir)
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

	// Prompt to create/update CLAUDE.md
	if err := promptAndUpdateClaudeMd(cmd, repoRoot); err != nil {
		// Non-fatal, just log
		cmd.Printf("%s Warning: failed to create/update CLAUDE.md: %v\n", infoStyle.Render("!"), err)
	}

	cmd.Printf("\n%s Initialization complete!\n", successStyle.Render("✓"))
	cmd.Println("Next steps:")
	cmd.Println("  - Run 'mesdx mcp' to start the MCP server")
	cmd.Println("  - Configure Claude Code to use this MCP server")

	return nil
}

const (
	claudeMdBeginMarker = "<!-- mesdx:begin -->"
	claudeMdEndMarker   = "<!-- mesdx:end -->"
)

// promptAndUpdateClaudeMd prompts the user to create/update CLAUDE.md with MesDX guidance.
func promptAndUpdateClaudeMd(cmd *cobra.Command, repoRoot string) error {
	claudeMdPath := filepath.Join(repoRoot, "CLAUDE.md")

	// Check if CLAUDE.md exists and has MesDX markers
	existingContent := ""
	_, err := os.Stat(claudeMdPath)
	claudeMdExists := err == nil
	hasMesdxSection := false

	if claudeMdExists {
		content, err := os.ReadFile(claudeMdPath)
		if err == nil {
			existingContent = string(content)
			hasMesdxSection = strings.Contains(existingContent, claudeMdBeginMarker)
		}
	}

	// Prompt user
	var shouldUpdate bool
	promptText := "Add MesDX MCP guidance to CLAUDE.md?"
	if hasMesdxSection {
		promptText = "Update MesDX MCP guidance in CLAUDE.md?"
	}

	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(promptText).
				Description("This will add/update a managed section with setup instructions and tool overview.").
				Value(&shouldUpdate),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	if !shouldUpdate {
		return nil
	}

	// Generate or update CLAUDE.md
	if err := updateClaudeMd(claudeMdPath, repoRoot, claudeMdExists); err != nil {
		return err
	}

	if hasMesdxSection {
		cmd.Printf("%s Updated MesDX guidance in CLAUDE.md\n", successStyle.Render("✓"))
	} else if claudeMdExists {
		cmd.Printf("%s Added MesDX guidance to CLAUDE.md\n", successStyle.Render("✓"))
	} else {
		cmd.Printf("%s Created CLAUDE.md with MesDX guidance\n", successStyle.Render("✓"))
	}

	return nil
}

// updateClaudeMd creates or updates CLAUDE.md with the managed MesDX section.
func updateClaudeMd(path string, repoRoot string, exists bool) error {
	managedContent := generateMesdxGuidance(repoRoot)

	if !exists {
		// Create new file with only the managed section
		content := claudeMdBeginMarker + "\n" + managedContent + "\n" + claudeMdEndMarker + "\n"
		return os.WriteFile(path, []byte(content), 0644)
	}

	// Update existing file: replace managed section or append it
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read existing CLAUDE.md: %w", err)
	}

	updated := updateManagedSection(string(existing), managedContent)
	return os.WriteFile(path, []byte(updated), 0644)
}

// updateManagedSection replaces or appends the managed section in existing content.
func updateManagedSection(existing string, managedContent string) string {
	beginIdx := strings.Index(existing, claudeMdBeginMarker)
	endIdx := strings.Index(existing, claudeMdEndMarker)

	if beginIdx >= 0 && endIdx > beginIdx {
		// Replace existing managed section
		before := existing[:beginIdx]
		after := existing[endIdx+len(claudeMdEndMarker):]
		return before + claudeMdBeginMarker + "\n" + managedContent + "\n" + claudeMdEndMarker + after
	}

	// Append managed section
	var buf bytes.Buffer
	buf.WriteString(existing)
	if !strings.HasSuffix(existing, "\n") {
		buf.WriteString("\n")
	}
	buf.WriteString("\n")
	buf.WriteString(claudeMdBeginMarker)
	buf.WriteString("\n")
	buf.WriteString(managedContent)
	buf.WriteString("\n")
	buf.WriteString(claudeMdEndMarker)
	buf.WriteString("\n")
	return buf.String()
}

// generateMesdxGuidance generates the managed MesDX guidance content.
func generateMesdxGuidance(repoRoot string) string {
	_ = repoRoot // Not used in template, but kept for future extensibility
	return `## MesDX Code Intelligence

MesDX provides precise, local-first code navigation and analysis via the Model Context Protocol (MCP).

### Setup

Add MesDX to Claude Code with this command:

` + "```bash" + `
claude mcp add mesdx --transport stdio -- mesdx mcp --cwd <REPO_ROOT>
` + "```" + `

Replace ` + "`<REPO_ROOT>`" + ` with the absolute path to this repository.


### Available Tools

**Navigation & Definition Lookup**
- ` + "`mesdx.projectInfo`" + ` — Get repo root, source roots, and database path
- ` + "`mesdx.goToDefinition`" + ` — Find symbol definitions by cursor position or name
- ` + "`mesdx.findUsages`" + ` — Find all references to a symbol with dependency scoring

**Impact Analysis**
- ` + "`mesdx.dependencyGraph`" + ` — Analyze inbound/outbound dependencies for refactor risk assessment

**Memory (Persistent Context)**
- ` + "`mesdx.memoryAppend`" + ` — Create project or file-scoped markdown notes
- ` + "`mesdx.memoryRead`" + ` — Read or list memory elements
- ` + "`mesdx.memorySearch`" + ` — Full-text search across memories
- ` + "`mesdx.memoryUpdate`" + ` — Update existing memories
- ` + "`mesdx.memoryGrepReplace`" + ` — Regex find-and-replace in memories
- ` + "`mesdx.memoryDelete`" + ` — Soft-delete memories

### Workflow Skills

For detailed step-by-step guidance on specific workflows, use MCP prompts:

- ` + "`mesdx.skill.bugfix`" + ` — Navigate, analyze, and document bug fixes
- ` + "`mesdx.skill.refactoring`" + ` — Safe refactoring with impact analysis
- ` + "`mesdx.skill.feature_development`" + ` — Plan and implement new features
- ` + "`mesdx.skill.security_analysis`" + ` — Find and document security issues

**Supported languages:** Go, Java, Rust, Python, TypeScript, JavaScript
`
}
