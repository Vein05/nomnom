package nomnom

import (
	"log"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQuery sends a query to the deepseek API to generate new file names
func SendQueryWithDeepSeek(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the Deepseek client

	// check if config.ai.apikey is set

	client := deepseek.NewClient(config.AI.APIKey)

	opts := QueryOpts{
		Model: deepseek.DeepSeekChat,
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

	if err := SendQueryToLLM(client, query, opts); err != nil {
		log.Fatalf("error: %v", err)
		return contentprocessors.Query{}, err
	}

	return query, nil
}
