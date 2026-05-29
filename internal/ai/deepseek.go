package ai

import (
	"fmt"

	content "nomnom/internal/content"
	configutils "nomnom/internal/utils"
)

func SendQueryWithDeepSeek(config configutils.Config, query content.Query) (content.Query, error) {
	if config.AI.APIKey == "" {
		return content.Query{}, fmt.Errorf("no API key provided for DeepSeek")
	}

	client, err := newDeepSeekClient(config.AI.APIKey)
	if err != nil {
		return content.Query{}, err
	}
	model := config.AI.Model
	if model == "" {
		model = "deepseek-chat"
	}

	opts := QueryOpts{
		Provider:    "deepseek",
		Model:       model,
		Case:        config.Case,
		MaxTokens:   config.AI.MaxTokens,
		Temperature: config.AI.Temperature,
	}

	reporterFor(query).Infof("You're using DeepSeek with model: %s", model)
	if err := SendQueryToLLM(client, config, &query, opts); err != nil {
		return content.Query{}, err
	}

	return query, nil
}
