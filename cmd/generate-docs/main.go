// generate-docs produces the repository-root SKILL.md file from the live
// Cobra and structcli command tree. The hand-written AGENTS.md files remain the
// source of project architecture and review guidance.
//
// Usage:
//
//	go run ./cmd/generate-docs/
//	make docs
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/leodido/structcli/generate"

	"github.com/major/marketsurge-agent/cmd"
)

// version is set via ldflags at build time, matching the main binary pattern.
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generate-docs: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outDir, err := projectRoot()
	if err != nil {
		return fmt.Errorf("finding project root: %w", err)
	}

	skill, err := generate.Skill(cmd.RootCommand(), generate.SkillOptions{
		Name:      "marketsurge-agent",
		Author:    "major",
		Version:   version,
		MCPServer: "marketsurge-agent --mcp",
	})
	if err != nil {
		return fmt.Errorf("generating SKILL.md: %w", err)
	}

	skill = labelSkillCodeFences(skill)
	path := filepath.Join(outDir, "SKILL.md")
	//nolint:gosec // G306: public documentation files should be world-readable
	if err := os.WriteFile(path, skill, 0o644); err != nil {
		return fmt.Errorf("writing SKILL.md: %w", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", path)

	return nil
}

func labelSkillCodeFences(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	inFence := false
	for i, line := range lines {
		if line != "```" {
			continue
		}

		if inFence {
			inFence = false
			continue
		}

		// structcli generates shell examples in SKILL.md with bare opening
		// fences. Add a language tag here so generated docs satisfy markdownlint
		// without forking or post-editing the generated file by hand.
		lines[i] = "```bash"
		inFence = true
	}

	return []byte(strings.Join(lines, "\n"))
}

// projectRoot returns the repository root by walking up from this source file's
// location. This works regardless of the working directory used to run the
// generator.
func projectRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not determine source file path")
	}

	// thisFile is cmd/generate-docs/main.go, so root is two levels up.
	return filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}
