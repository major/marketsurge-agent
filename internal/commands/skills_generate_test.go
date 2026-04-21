package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

type skillsGenerateEnvelope struct {
	Data struct {
		GeneratedFiles []string `json:"generated_files"`
		OutputDir      string   `json:"output_dir"`
	} `json:"data"`
	Metadata map[string]any `json:"metadata"`
}

func TestSkillsGenerateCommand(t *testing.T) {
	t.Run("generates skill files to default directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "skills")

		var buf bytes.Buffer
		cmd := SkillsGenerateCommand(&buf)

		// Create a CLI context with the output-dir flag
		ctx := context.Background()
		cliCmd := &cli.Command{
			Name: "test",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: outputDir,
				},
			},
		}

		// Manually set the flag value
		flagSet := cliCmd.Flags[0].(*cli.StringFlag)
		_ = flagSet

		// Execute the command
		err := cmd.Action(ctx, &cli.Command{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: outputDir,
				},
			},
		})

		require.NoError(t, err)
		envelope := decodeSkillsGenerateEnvelope(t, buf.Bytes())
		assert.Equal(t, outputDir, envelope.Data.OutputDir)
		assert.Len(t, envelope.Data.GeneratedFiles, 6)
		assert.NotEmpty(t, envelope.Metadata["timestamp"])

		// Verify output directory was created
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Errorf("Output directory was not created: %s", outputDir)
		}

		// Verify skill files were created
		expectedFiles := []string{"stock.md", "fundamental.md", "ownership.md", "rs-history.md", "chart.md", "catalog.md"}
		for _, filename := range expectedFiles {
			filepath := filepath.Join(outputDir, filename)
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("Skill file was not created: %s", filepath)
			}
		}
	})

	t.Run("generates files with expected content", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "skills")

		var buf bytes.Buffer
		cmd := SkillsGenerateCommand(&buf)

		ctx := context.Background()
		err := cmd.Action(ctx, &cli.Command{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: outputDir,
				},
			},
		})

		require.NoError(t, err)
		envelope := decodeSkillsGenerateEnvelope(t, buf.Bytes())
		assert.Equal(t, outputDir, envelope.Data.OutputDir)

		// Check stock.md content
		stockFile := filepath.Join(outputDir, "stock.md")
		content, err := os.ReadFile(stockFile)
		if err != nil {
			t.Fatalf("Failed to read stock.md: %v", err)
		}

		contentStr := string(content)

		// Verify expected sections exist
		expectedSections := []string{
			"# Stock Analysis Skill",
			"## Overview",
			"## Tools",
			"### get_stock",
			"### analyze_stock",
			"## Workflow Guidance",
		}

		for _, section := range expectedSections {
			if !bytes.Contains(content, []byte(section)) {
				t.Errorf("Expected section '%s' not found in stock.md", section)
			}
		}

		// Verify tool descriptions are present
		if !bytes.Contains(content, []byte("Fetch stock data including ratings")) {
			t.Error("Tool description not found in stock.md")
		}

		// Verify example invocations
		if !bytes.Contains(content, []byte("marketsurge-agent stock get AAPL")) {
			t.Error("Example invocation not found in stock.md")
		}

		// Verify expected output shape
		if !bytes.Contains(content, []byte("\"symbol\": \"AAPL\"")) {
			t.Error("Expected output shape not found in stock.md")
		}

		t.Logf("stock.md content length: %d bytes", len(contentStr))
	})

	t.Run("uses custom output directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		customDir := filepath.Join(tmpDir, "custom", "skills")

		var buf bytes.Buffer
		cmd := SkillsGenerateCommand(&buf)

		ctx := context.Background()
		err := cmd.Action(ctx, &cli.Command{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: customDir,
				},
			},
		})

		require.NoError(t, err)
		envelope := decodeSkillsGenerateEnvelope(t, buf.Bytes())
		assert.Equal(t, customDir, envelope.Data.OutputDir)

		// Verify custom directory was created
		if _, err := os.Stat(customDir); os.IsNotExist(err) {
			t.Errorf("Custom output directory was not created: %s", customDir)
		}

		// Verify files exist in custom directory
		stockFile := filepath.Join(customDir, "stock.md")
		if _, err := os.Stat(stockFile); os.IsNotExist(err) {
			t.Errorf("Skill file was not created in custom directory: %s", stockFile)
		}
	})

	t.Run("generates all command groups", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "skills")

		var buf bytes.Buffer
		cmd := SkillsGenerateCommand(&buf)

		ctx := context.Background()
		err := cmd.Action(ctx, &cli.Command{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: outputDir,
				},
			},
		})

		commandGroups := []string{"stock", "fundamental", "ownership", "rs-history", "chart", "catalog"}
		require.NoError(t, err)
		envelope := decodeSkillsGenerateEnvelope(t, buf.Bytes())
		assert.Len(t, envelope.Data.GeneratedFiles, len(commandGroups))

		// Verify all command groups have skill files
		for _, group := range commandGroups {
			filepath := filepath.Join(outputDir, group+".md")
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("Skill file for group '%s' was not created", group)
			}

			// Verify file has content
			content, err := os.ReadFile(filepath)
			if err != nil {
				t.Errorf("Failed to read skill file for group '%s': %v", group, err)
			}

			if len(content) == 0 {
				t.Errorf("Skill file for group '%s' is empty", group)
			}
		}
	})

	t.Run("outputs success message", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputDir := filepath.Join(tmpDir, "skills")

		var buf bytes.Buffer
		cmd := SkillsGenerateCommand(&buf)

		ctx := context.Background()
		err := cmd.Action(ctx, &cli.Command{
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output-dir",
					Value: outputDir,
				},
			},
		})

		require.NoError(t, err)
		envelope := decodeSkillsGenerateEnvelope(t, buf.Bytes())
		assert.Equal(t, outputDir, envelope.Data.OutputDir)
		assert.Len(t, envelope.Data.GeneratedFiles, 6)
		assert.NotEmpty(t, envelope.Metadata["timestamp"])
	})
}

func decodeSkillsGenerateEnvelope(t *testing.T, data []byte) skillsGenerateEnvelope {
	t.Helper()

	var envelope skillsGenerateEnvelope
	require.NoError(t, json.Unmarshal(data, &envelope))
	return envelope
}
