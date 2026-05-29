package ai

import (
	deepseek "github.com/cohesion-org/deepseek-go"
	api "github.com/ollama/ollama/api"
)

var newDeepSeekClient = func(apiKey string) (*deepseek.Client, error) {
	return deepseek.NewClientWithOptions(apiKey)
}

var newOpenRouterClient = func(apiKey string) (*deepseek.Client, error) {
	return deepseek.NewClientWithOptions(apiKey, deepseek.WithBaseURL("https://openrouter.ai/api/v1/"))
}

var newOllamaClient = api.ClientFromEnvironment
