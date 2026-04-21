package commands

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

// SkillsGenerateCommand returns a stub CLI command for generating agent skill files.
// The full implementation will be provided by T21.
func SkillsGenerateCommand(w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate agent skill files",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// TODO: T21 will implement this
			fmt.Fprintln(w, "Skills generation not yet implemented")
			return nil
		},
	}
}
