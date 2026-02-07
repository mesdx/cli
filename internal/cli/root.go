package cli

import (
	"github.com/codeintelx/cli/internal/selfupdate"
	"github.com/spf13/cobra"
)

// Version is the version of codeintelx CLI.
// Update this constant manually on every release.
const Version = "v0.1.1"

// NewRootCmd creates the root command for codeintelx
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "codeintelx",
		Short:   "Code intelligence CLI and MCP server",
		Long:    "Codeintelx is a CLI tool for indexing codebases and exposing code intelligence via MCP.",
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip self-update for the mcp command (daemon-like usage)
			if cmd.Name() == "mcp" {
				return nil
			}

			// Check and update (best-effort, never fails the command)
			_ = selfupdate.CheckAndUpdate(Version)
			return nil
		},
	}

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newMcpCmd())

	return rootCmd
}
