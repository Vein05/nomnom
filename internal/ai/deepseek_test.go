package ai

import (
	"testing"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/stretchr/testify/assert"
)

func TestSendQueryWithDeepSeekMockServer(t *testing.T) {
	server, getRecords := newDeepSeekMockServer("renamed_document")
	defer server.Close()

	originalFactory := newDeepSeekClient
	newDeepSeekClient = func(apiKey string) (*deepseek.Client, error) {
		return deepseek.NewClientWithOptions(apiKey, deepseek.WithBaseURL(server.URL+"/"))
	}
	t.Cleanup(func() {
		newDeepSeekClient = originalFactory
	})

	config := configutils.Config{
		AI: configutils.AIConfig{
			Provider: "deepseek",
			APIKey:   "test-api-key",
			Model:    "mock-deepseek-model",
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

	result, err := SendQueryWithDeepSeek(config, query)
	if err != nil {
		t.Fatalf("SendQueryWithDeepSeek() error = %v", err)
	}

	records := getRecords()
	assert.Len(t, records, 2)
	for _, record := range records {
		assert.Equal(t, "/chat/completions", record.Path)
		assert.Equal(t, "mock-deepseek-model", record.Request.Model)
		assert.Len(t, record.Request.Messages, 2)
	}

	assert.Equal(t, "renamed_document.txt", result.Plan[0].SuggestedName)
	assert.Equal(t, "renamed_document.txt", result.Plan[1].SuggestedName)
}
