package cli

import (
	"github.com/spf13/cobra"
)

var version = "dev"

// NewRootCmd creates the root command for codeintelx
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "codeintelx",
		Short: "Code intelligence CLI and MCP server",
		Long:  "Codeintelx is a CLI tool for indexing codebases and exposing code intelligence via MCP.",
		Version: version,
	}

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newMcpCmd())

	return rootCmd
}
