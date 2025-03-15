package nomnom

import (
	"log"
	"os"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQuery sends a query to the deepseek API to generate new file names
func SendQueryWithDeepSeek(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the Deepseek client

	// check if config.ai.apikey is set

	var apiKey string

	if config.AI.APIKey == "" {
		log.Println("No API key found in config, using environment variable DEEPSEEK_API_KEY. If you want to use a different API key, set config.ai.apikey in your config.json file.")
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
	} else {
		apiKey = config.AI.APIKey
	}

	client := deepseek.NewClient(apiKey)

	opts := QueryOpts{
		Model: deepseek.DeepSeekChat,
		Case:  config.Case,
	}

	if err := SendQueryToLLM(client, query, opts); err != nil {
		log.Fatalf("error: %v", err)
		return contentprocessors.Query{}, err
	}

	return query, nil
}
