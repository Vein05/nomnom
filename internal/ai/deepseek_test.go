package nomnom

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
		Folders: []contentprocessors.FolderType{
			{
				Name:       "TestFolder",
				FolderPath: "/test/path",
				FileList: []contentprocessors.File{
					{
						Name:    "test_file.txt",
						Path:    "/test/path/test_file.txt",
						Context: "This is a test file containing important information about a game called Rain World.",
					},
					{
						Name:    "presentation.ppt",
						Path:    "/test/path/presentation.ppt",
						Context: "This is a PowerPoint presentation about quarterly sales results for Q1 2024.",
					},
					{
						Name:    "report.pdf",
						Path:    "/test/path/report.pdf",
						Context: "This is the annual financial report for 2023 fiscal year with detailed analysis.",
					},
				},
			},
		},
	}

	// Test the SendQuery function
	SendQueryWithDeepSeek(config, testQuery)

	// Verify that new names were assigned for all files
	for i, file := range testQuery.Folders[0].FileList {
		if file.NewName == "" {
			t.Errorf("Expected NewName to be set for file %s", file.Name)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, file.Name, file.NewName)
	}
}
