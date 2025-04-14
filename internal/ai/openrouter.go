package nomnom

import (
	"fmt"
	"os"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQueryWithOpenRouter sends a query to the OpenRouter API to generate new file names
func SendQueryWithOpenRouter(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the client with OpenRouter base URL
	baseURL := "https://openrouter.ai/api/v1/"
	client := deepseek.NewClient(os.Getenv("OPENROUTER_API_KEY"), baseURL)

	opts := QueryOpts{
		Model: config.AI.Model,
		Case:  config.Case,
	}

	if err := SendQueryToLLM(client, query, opts); err != nil {
		fmt.Printf("❌ Failed to process query with OpenRouter: %v", err)
		return contentprocessors.Query{}, err
	}
	return query, nil
}
