package nomnom

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
	}

	// Create a test query with sample data
	testQuery := contentprocessors.Query{
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
						Name:    "image.jpg",
						Path:    "../../demo/small/Image 484972979.jpg",
						Context: "",
					},
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
	for i, file := range result.Folders[0].FileList {
		if file.NewName == "" {
			t.Errorf("Expected NewName to be set for file %s", file.Name)
		}
		fmt.Printf("File %d - Old Name: %s, New Name: %s\n", i+1, file.Name, file.NewName)
	}
}
