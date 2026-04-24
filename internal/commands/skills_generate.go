package commands

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/major/marketsurge-agent/internal/output"
	"github.com/urfave/cli/v3"
)

type skillsGenerateData struct {
	GeneratedFiles []string `json:"generated_files"`
	OutputDir      string   `json:"output_dir"`
}

// SkillsGenerateCommand returns a CLI command for generating agent skill files.
func SkillsGenerateCommand(w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate agent skill files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory for skill files",
				Value: "skills/",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			outputDir := cmd.String("output-dir")

			if err := os.MkdirAll(outputDir, 0o750); err != nil {
				return err
			}

			commandGroups := []string{"stock", "fundamental", "ownership", "rs-history", "chart", "catalog"}
			generatedFiles := make([]string, 0, len(commandGroups))
			for _, group := range commandGroups {
				content, ok := SkillTemplates[group]
				if !ok {
					continue
				}

				filename := filepath.Join(outputDir, group+".md")
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil { //nolint:gosec // generated docs are world-readable
					return err
				}
				generatedFiles = append(generatedFiles, filename)
			}

			return output.WriteSuccess(w, skillsGenerateData{
				GeneratedFiles: generatedFiles,
				OutputDir:      outputDir,
			}, map[string]any{"timestamp": time.Now().UTC().Format(time.RFC3339)})
		},
	}
}
