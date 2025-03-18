package nomnom

import (
	"log"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQueryWithOllama sends a query to the Ollama API to generate new file names
func SendQueryWithOllama(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	// Set up the client with Ollama base URL
	baseURL := "http://localhost:11434/api/"
	client := deepseek.NewClient("", baseURL) // No API key needed for local Ollama

	log.Printf("[INFO] Using Ollama model: %s", config.AI.Model)

	opts := QueryOpts{
		Model: config.AI.Model,
		Case:  config.Case,
	}

	if err := SendQueryToLLM(client, query, opts); err != nil {
		log.Printf("[ERROR] Failed to process query with Ollama: %v", err)
		return contentprocessors.Query{}, err
	}

	log.Printf("[INFO] Successfully processed query with Ollama")
	return query, nil
}
