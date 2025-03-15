package nomnom

import (
	"os"
	"testing"

	contentprocessors "nomnom/internal/content"
	utils "nomnom/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestHandleAIProviderSelection(t *testing.T) {
	// Save original environment variables
	originalDeepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	originalOllamaKey := os.Getenv("OLLAMA_API_KEY")
	originalOpenRouterKey := os.Getenv("OPENROUTER_API_KEY")

	// Clean up function to restore environment variables
	defer func() {
		os.Setenv("DEEPSEEK_API_KEY", originalDeepseekKey)
		os.Setenv("OLLAMA_API_KEY", originalOllamaKey)
		os.Setenv("OPENROUTER_API_KEY", originalOpenRouterKey)
	}()

	// Test cases
	tests := []struct {
		name          string
		config        utils.Config
		envSetup      map[string]string
		expectedModel string
		expectedError bool
	}{
		{
			name: "Config with Deepseek API key",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "deepseek",
			expectedError: false,
		},
		{
			name: "No config API key but Deepseek env variable set",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
				},
			},
			envSetup: map[string]string{
				"DEEPSEEK_API_KEY": "dummy-key",
			},
			expectedModel: "deepseek",
			expectedError: false,
		},
		{
			name: "Ollama provider without API key",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "ollama",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "ollama",
			expectedError: false,
		},
		{
			name: "No config API key but OpenRouter env variable set",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "openrouter",
				},
			},
			envSetup: map[string]string{
				"OPENROUTER_API_KEY": "dummy-key",
			},
			expectedModel: "openrouter",
			expectedError: false,
		},
		{
			name: "Multiple env variables set - should use provider from config",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "ollama",
					APIKey:   "dummy-key",
				},
			},
			envSetup: map[string]string{
				"DEEPSEEK_API_KEY":   "dummy-key",
				"OPENROUTER_API_KEY": "dummy-key",
			},
			expectedModel: "ollama",
			expectedError: false,
		},
		{
			name: "No API keys available for Deepseek",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "No API keys available for OpenRouter",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "openrouter",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "Invalid provider",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "invalid-provider",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "No provider set - should default to deepseek",
			config: utils.Config{
				AI: utils.AIConfig{
					APIKey: "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "deepseek",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			os.Unsetenv("DEEPSEEK_API_KEY")
			os.Unsetenv("OLLAMA_API_KEY")
			os.Unsetenv("OPENROUTER_API_KEY")

			// Set up environment variables for the test
			for key, value := range tt.envSetup {
				os.Setenv(key, value)
			}

			// Create a minimal query just to pass to HandleAI
			query := contentprocessors.Query{
				Folders: []contentprocessors.FolderType{
					{
						Name: "test",
						FileList: []contentprocessors.File{
							{
								Name:    "test.txt",
								Context: "test",
							},
						},
					},
				},
			}

			// Call HandleAI but only check for error conditions
			_, err := HandleAI(tt.config, query)

			if tt.expectedError {
				assert.Error(t, err, "Expected an error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
			}
		})
	}
}
