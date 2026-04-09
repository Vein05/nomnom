package ai

import (
	"fmt"
	"os"
	"testing"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"
)

func TestSendQuery(t *testing.T) {
	// Skip test if DEEPSEEK_API_KEY is not set
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping test")
	}

	// Set up test configuration
	config := configutils.Config{
		AI: configutils.AIConfig{
			APIKey: os.Getenv("DEEPSEEK_API_KEY"),
		},
	}

	// Create a test query with sample data
	testQuery := contentprocessors.Query{
		Prompt: "What is the title of this document? Only respond with the title and extension in snake case.",
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
					Context:      "This is a PowerPoint presentation about quarterly sales results for Q1 2024.",
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

	// Test the SendQuery function
	SendQueryWithDeepSeek(config, testQuery)

	// Verify that new names were assigned for all files
	for i, entry := range testQuery.Plan {
		if entry.SuggestedName == "" {
			t.Errorf("Expected SuggestedName to be set for file %s", entry.File.OriginalName)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, entry.File.OriginalName, entry.SuggestedName)
	}
}
