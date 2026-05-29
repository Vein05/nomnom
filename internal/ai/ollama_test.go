package ai

import (
	"net/url"
	"path/filepath"
	"testing"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	api "github.com/ollama/ollama/api"
	"github.com/stretchr/testify/assert"
)

func TestSendQueryWithOllamaMockServer(t *testing.T) {
	server, getRecords := newOllamaMockServer("renamed_preview")
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	originalFactory := newOllamaClient
	newOllamaClient = func() (*api.Client, error) {
		return api.NewClient(serverURL, server.Client()), nil
	}
	t.Cleanup(func() {
		newOllamaClient = originalFactory
	})

	imagePath, err := filepath.Abs(filepath.Join("..", "..", "data", "summary.png"))
	if err != nil {
		t.Fatalf("resolve image path: %v", err)
	}

	config := configutils.Config{
		AI: configutils.AIConfig{
			Model:  "mock-ollama-model",
			Prompt: "Rename files from their content.",
			Vision: configutils.VisionConfig{Enabled: true},
		},
		Performance: configutils.PerformanceConfig{
			AI: configutils.PerformanceAIConfig{
				Workers: 1,
				Timeout: "2s",
				Retries: 1,
			},
		},
	}

	query := contentprocessors.Query{
		Prompt: config.AI.Prompt,
		Scan: contentprocessors.ScanResult{
			RootDir: "/tmp/mock",
			Files: []contentprocessors.ScannedFile{
				{SourcePath: "/tmp/mock/a.txt", RelativePath: "a.txt", OriginalName: "a.txt", Context: "alpha"},
				{SourcePath: imagePath, RelativePath: "summary.png", OriginalName: "summary.png", Context: "image"},
			},
		},
	}

	result, err := SendQueryWithOllama(config, query)
	if err != nil {
		t.Fatalf("SendQueryWithOllama() error = %v", err)
	}

	records := getRecords()
	assert.Len(t, records, 2)
	assert.Equal(t, "/api/chat", records[0].Path)
	assert.Equal(t, "/api/chat", records[1].Path)

	firstMessages, ok := records[0].Request["messages"].([]any)
	if !ok {
		t.Fatalf("expected first request to include messages")
	}
	assert.Len(t, firstMessages, 2)

	secondMessages, ok := records[1].Request["messages"].([]any)
	if !ok {
		t.Fatalf("expected second request to include messages")
	}
	assert.Len(t, secondMessages, 2)
	secondUserMessage, ok := secondMessages[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second user message to be a map")
	}
	images, ok := secondUserMessage["images"].([]any)
	if !ok {
		t.Fatalf("expected vision request to include images")
	}
	assert.NotEmpty(t, images)

	assert.Equal(t, "renamed_preview.txt", result.Plan[0].SuggestedName)
	assert.Equal(t, "renamed_preview.png", result.Plan[1].SuggestedName)
}
