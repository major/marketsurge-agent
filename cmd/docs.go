// Package cmd provides the cobra command tree for marketsurge-agent.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(newDocsCmd())
}

func newDocsCmd() *cobra.Command {
	var outputDir string
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate markdown documentation for all commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(outputDir, 0o750); err != nil {
				return err
			}
			return doc.GenMarkdownTree(rootCmd, outputDir)
		},
	}
	cmd.Flags().StringVarP(&outputDir, "output", "o", "./docs", "Output directory for generated docs")
	return cmd
}
