package cli

import (
	"context"
	"fmt"

	"github.com/codeintelx/cli/internal/config"
	"github.com/codeintelx/cli/internal/repo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

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
	cfg, err := config.Load(codeintelxDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "codeintelx",
		Version: "1.0.0",
	}, nil)

	// Register project info tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.projectInfo",
		Description: "Get project information including repo root, source roots, and database path",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		info := map[string]interface{}{
			"repoRoot":    cfg.RepoRoot,
			"sourceRoots": cfg.SourceRoots,
			"dbPath":      fmt.Sprintf("%s/index.db", codeintelxDir),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Repo Root: %s\nSource Roots: %v\nDatabase: %s",
						cfg.RepoRoot, cfg.SourceRoots, fmt.Sprintf("%s/index.db", codeintelxDir)),
				},
			},
		}, info, nil
	})

	// Run server over stdio
	return server.Run(context.Background(), &mcp.StdioTransport{})
}
