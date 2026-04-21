package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

// SkillsGenerateCommand returns a CLI command for generating agent skill files.
func SkillsGenerateCommand(w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate agent skill files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory for skill files",
				Value:   "skills/",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			outputDir := cmd.String("output-dir")

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintf(w, "Error creating output directory: %v\n", err)
				return err
			}

			// Generate skill files for each command group
			commandGroups := []string{"stock", "fundamental", "ownership", "rs-history", "chart", "catalog"}
			for _, group := range commandGroups {
				content, ok := SkillTemplates[group]
				if !ok {
					fmt.Fprintf(w, "Warning: No template found for group '%s'\n", group)
					continue
				}

				filename := filepath.Join(outputDir, group+".md")
				if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
					fmt.Fprintf(w, "Error writing skill file %s: %v\n", filename, err)
					return err
				}

				fmt.Fprintf(w, "Generated skill file: %s\n", filename)
			}

			fmt.Fprintf(w, "Successfully generated %d skill files to %s\n", len(commandGroups), outputDir)
			return nil
		},
	}
}
