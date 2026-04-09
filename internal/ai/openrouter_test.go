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
						Context: "This is a PowerPoint presentation about quarterly sales results for Q1 2024. ",
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

	// Test the SendQueryWithOpenRouter function
	result, err := SendQueryWithOpenRouter(config, testQuery)
	if err != nil {
		t.Fatalf("SendQueryWithOpenRouter() error = %v", err)
	}

	// Verify that new names were assigned for all files
	for i, file := range result.Folders[0].FileList {
		if file.NewName == "" {
			t.Errorf("Expected NewName to be set for file %s", file.Name)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, file.Name, file.NewName)
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
