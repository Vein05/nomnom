package ai

import (
	"fmt"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQueryWithOpenRouter sends a query to the OpenRouter API to generate new file names
func SendQueryWithOpenRouter(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the client with OpenRouter base URL
	baseURL := "https://openrouter.ai/api/v1/"

	var key string
	if config.AI.APIKey != "" {
		key = config.AI.APIKey
	} else {
		return contentprocessors.Query{}, fmt.Errorf("no API key provided for OpenRouter")
	}

	client := deepseek.NewClient(key, baseURL)

	model := config.AI.Model
	if model == "" {
		return contentprocessors.Query{}, fmt.Errorf("no model provided for OpenRouter")
	}
	opts := QueryOpts{
		Model: model,
		Case:  config.Case,
	}

	// check if config.ai.max_tokens is set
	if config.AI.MaxTokens != 0 {
		opts.MaxTokens = config.AI.MaxTokens
	}

	// check if config.ai.temperature is set
	if config.AI.Temperature != 0 {
		opts.Temperature = config.AI.Temperature
	}

	reporterFor(query).Infof("You're using OpenRouter with model: %s", model)

	if err := SendQueryToLLM(client, config, query, opts); err != nil {
		reporterFor(query).Errorf("Failed to process query with OpenRouter: %v", err)
		return contentprocessors.Query{}, err
	}
	return query, nil
}
