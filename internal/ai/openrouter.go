package nomnom

import (
	"os"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	log "github.com/charmbracelet/log"
	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQueryWithOpenRouter sends a query to the OpenRouter API to generate new file names
func SendQueryWithOpenRouter(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the client with OpenRouter base URL
	baseURL := "https://openrouter.ai/api/v1/"
	client := deepseek.NewClient(os.Getenv("OPENROUTER_API_KEY"), baseURL)

	log.Info("Using OpenRouter: ", "model", config.AI.Model)

	opts := QueryOpts{
		Model: config.AI.Model,
		Case:  config.Case,
	}

	if err := SendQueryToLLM(client, query, opts); err != nil {
		log.Error("Failed to process query with OpenRouter: ", "error", err)
		return contentprocessors.Query{}, err
	}

	log.Info("Successfully processed query with OpenRouter")
	return query, nil
}
