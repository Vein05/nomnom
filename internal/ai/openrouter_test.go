package ai

import (
	"testing"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/stretchr/testify/assert"
)

func TestSendQueryWithOpenRouterMockServer(t *testing.T) {
	server, getRecords := newDeepSeekMockServer("renamed_document")
	defer server.Close()

	originalFactory := newOpenRouterClient
	newOpenRouterClient = func(apiKey string) (*deepseek.Client, error) {
		return deepseek.NewClientWithOptions(apiKey, deepseek.WithBaseURL(server.URL+"/"))
	}
	t.Cleanup(func() {
		newOpenRouterClient = originalFactory
	})

	config := configutils.Config{
		AI: configutils.AIConfig{
			Provider: "openrouter",
			APIKey:   "test-api-key",
			Model:    "mock-openrouter-model",
			Prompt:   "Rename files from their content.",
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
				{SourcePath: "/tmp/mock/b.txt", RelativePath: "b.txt", OriginalName: "b.txt", Context: "beta"},
			},
		},
	}

	result, err := SendQueryWithOpenRouter(config, query)
	if err != nil {
		t.Fatalf("SendQueryWithOpenRouter() error = %v", err)
	}

	records := getRecords()
	assert.Len(t, records, 2)
	for _, record := range records {
		assert.Equal(t, "/chat/completions", record.Path)
		assert.Equal(t, "mock-openrouter-model", record.Request.Model)
		assert.Len(t, record.Request.Messages, 2)
	}

	assert.Equal(t, "renamed_document.txt", result.Plan[0].SuggestedName)
	assert.Equal(t, "renamed_document.txt", result.Plan[1].SuggestedName)
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
