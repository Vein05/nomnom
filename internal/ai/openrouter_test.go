package ai

import (
	"fmt"
	"os"
	"testing"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"
)

func TestSendQueryWithOpenRouter(t *testing.T) {
	// Skip test if OPENROUTER_API_KEY is not set
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping test")
	}
	if os.Getenv("OPENROUTER_MODEL") == "" {
		t.Skip("OPENROUTER_MODEL not set, skipping test")
	}

	config := configutils.Config{
		AI: configutils.AIConfig{
			Provider: "openrouter",
			APIKey:   os.Getenv("OPENROUTER_API_KEY"),
			Model:    os.Getenv("OPENROUTER_MODEL"),
			Prompt:   "Rename the file from the content. Return only the filename with extension in snake case.",
		},
		Performance: configutils.PerformanceConfig{
			AI: configutils.PerformanceAIConfig{
				Workers: 1,
				Timeout: "30s",
				Retries: 1,
			},
		},
	}

	// Create a test query with sample data
	testQuery := contentprocessors.Query{
		Prompt: config.AI.Prompt,
		Scan: contentprocessors.ScanResult{
			RootDir: "/test/path",
			Files: []contentprocessors.ScannedFile{
				{
					SourcePath:   "/test/path/test_file.txt",
					RelativePath: "test_file.txt",
					OriginalName: "test_file.txt",
					Context:      "This is a test file containing important information about a game called Rain World.",
				},
				{
					SourcePath:   "/test/path/presentation.ppt",
					RelativePath: "presentation.ppt",
					OriginalName: "presentation.ppt",
					Context:      "This is a PowerPoint presentation about quarterly sales results for Q1 2024. ",
				},
				{
					SourcePath:   "/test/path/report.pdf",
					RelativePath: "report.pdf",
					OriginalName: "report.pdf",
					Context:      "This is the annual financial report for 2023 fiscal year with detailed analysis.",
				},
			},
		},
	}

	// Test the SendQueryWithOpenRouter function
	result, err := SendQueryWithOpenRouter(config, testQuery)
	if err != nil {
		t.Fatalf("SendQueryWithOpenRouter() error = %v", err)
	}

	// Verify that new names were assigned for all files
	for i, entry := range result.Plan {
		if entry.SuggestedName == "" {
			t.Errorf("Expected SuggestedName to be set for file %s", entry.File.OriginalName)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, entry.File.OriginalName, entry.SuggestedName)
	}
}

func TestSendQueryWithOpenRouterNoKey(t *testing.T) {
	config := configutils.Config{
		AI: configutils.AIConfig{
			Provider: "openrouter",
		},
	}

	_, err := SendQueryWithOpenRouter(config, contentprocessors.Query{})
	if err == nil {
		t.Errorf("Expected error when no API key is provided, got nil")
	}
}
