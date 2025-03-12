package nomnom

import (
	"fmt"
	"log"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// SendQueryWithOllama sends a query to the Ollama API to generate new file names
func SendQueryWithOllama(config configutils.Config, query contentprocessors.Query) {
	// Set up the client with Ollama base URL
	baseURL := "http://localhost:11434/api/"
	client := deepseek.NewClient("", baseURL) // No API key needed for local Ollama

	fmt.Println("Using Ollama model:", config.AI.OllamaModel)

	if err := SendQueryToLLM(client, query, QueryOpts{Model: config.AI.OllamaModel}); err != nil {
		log.Fatalf("error: %v", err)
	}
}
