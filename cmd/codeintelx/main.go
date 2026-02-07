package main

import (
	"github.com/codeintelx/cli/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cli.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}
