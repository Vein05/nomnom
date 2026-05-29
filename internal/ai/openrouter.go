package ai

import (
	"fmt"

	content "nomnom/internal/content"
	configutils "nomnom/internal/utils"
)

func SendQueryWithOpenRouter(config configutils.Config, query content.Query) (content.Query, error) {
	if config.AI.APIKey == "" {
		return content.Query{}, fmt.Errorf("no API key provided for OpenRouter")
	}
	if config.AI.Model == "" {
		return content.Query{}, fmt.Errorf("no model provided for OpenRouter")
	}

	client, err := newOpenRouterClient(config.AI.APIKey)
	if err != nil {
		return content.Query{}, err
	}
	opts := QueryOpts{
		Provider:    "openrouter",
		Model:       config.AI.Model,
		Case:        config.Case,
		MaxTokens:   config.AI.MaxTokens,
		Temperature: config.AI.Temperature,
	}

	reporterFor(query).Infof("You're using OpenRouter with model: %s", config.AI.Model)
	if err := SendQueryToLLM(client, config, &query, opts); err != nil {
		return content.Query{}, err
	}
	return query, nil
}
