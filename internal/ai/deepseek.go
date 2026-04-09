package ai

import (
	"fmt"

	content "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

func SendQueryWithDeepSeek(config configutils.Config, query content.Query) (content.Query, error) {
	if config.AI.APIKey == "" {
		return content.Query{}, fmt.Errorf("no API key provided for DeepSeek")
	}

	client := deepseek.NewClient(config.AI.APIKey)
	model := config.AI.Model
	if model == "" {
		model = deepseek.DeepSeekChat
	}

	opts := QueryOpts{
		Provider:    "deepseek",
		Model:       model,
		Case:        config.Case,
		MaxTokens:   config.AI.MaxTokens,
		Temperature: config.AI.Temperature,
	}

	reporterFor(query).Infof("You're using DeepSeek with model: %s", model)
	if err := SendQueryToLLM(client, config, query, opts); err != nil {
		return content.Query{}, err
	}

	return query, nil
}
