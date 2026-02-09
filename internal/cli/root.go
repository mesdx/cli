package cli

import (
	"github.com/mesdx/cli/internal/selfupdate"
	"github.com/spf13/cobra"
)

// Version is the version of mesdx CLI.
// Update this constant manually on every release.
const Version = "v0.1.3"

// NewRootCmd creates the root command for mesdx
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "mesdx",
		Short:   "Code intelligence CLI and MCP server",
		Long:    "MesDX is a CLI tool for indexing codebases and exposing code intelligence via MCP.",
		Version: Version,
		// SilenceErrors prevents Cobra from printing errors twice
		// (once by the command, once by the root)
		SilenceErrors: true,
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
	rootCmd.AddCommand(newMemoryCmd())

	return rootCmd
}
