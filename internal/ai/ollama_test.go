package ai

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"
)

func isOllamaRunning() bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func TestSendQueryWithOllama(t *testing.T) {

	if !isOllamaRunning() {
		t.Skip("Ollama server is not running on localhost:11434")
	}

	// Set up test configuration
	config := configutils.Config{
		AI: configutils.AIConfig{
			Model:  "llama3.2-vision:11b",
			Prompt: "What is the name of this document or image attached to this request? If it's an image, analyse the contents to give out a proper name. Only respond with the name in snake case.",
			Vision: configutils.VisionConfig{
				Enabled: true,
			},
		},
		Performance: configutils.PerformanceConfig{
			AI: configutils.PerformanceAIConfig{
				Workers: 4,
				Timeout: "30",
				Retries: 3,
			},
		},
	}

	// Create a test query with sample data
	testQuery := contentprocessors.Query{
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
					SourcePath:   "../../demo/small/Image 484972979.jpg",
					RelativePath: "image.jpg",
					OriginalName: "image.jpg",
					Context:      "",
				},
			},
		},
	}

	// Test the SendQueryWithOllama function
	result, err := SendQueryWithOllama(config, testQuery)
	if err != nil {
		t.Fatalf("Failed to process query with Ollama: %v", err)
	}

	// Verify that new names were assigned for all files
	for i, entry := range result.Plan {
		if entry.SuggestedName == "" {
			t.Errorf("Expected SuggestedName to be set for file %s", entry.File.OriginalName)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, entry.File.OriginalName, entry.SuggestedName)
	}
}
