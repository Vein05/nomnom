// Package ai provides a collection of AI models and functions for use in the NomNom project.
package nomnom

import (
	"context"
	"fmt"
	"log"
	"time"

	contentprocessors "nomnom/internal/content"
	"os"
	"strings"

	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// QueryOpts contains options for the query
type QueryOpts struct {
	Model       string
	Case        string
	MaxTokens   int
	Temperature float64
}

type result struct {
	index int
	name  string
	err   error
}

// HandleAI is a function that handles the AI model selection and query execution and returns the result.
func HandleAI(config utils.Config, query contentprocessors.Query) (contentprocessors.Query, error) {
	// Select the AI model based on the config
	var aiModel string

	// we first check if the provider is set, if not we default to deepseek
	// we currently check if we are serving deepseek, ollama or openrouter
	if config.AI.Provider != "" {
		if config.AI.Provider == "deepseek" {
			log.Printf("🤖 Using deepseek as AI provider")
			aiModel = "deepseek"
		} else if config.AI.Provider == "ollama" {
			log.Printf("🤖 Using ollama as AI provider")
			aiModel = "ollama"
		} else if config.AI.Provider == "openrouter" {
			log.Printf("🤖 Using openrouter as AI provider")
			aiModel = "openrouter"
		} else {
			log.Printf("❌ Invalid AI provider: %s", config.AI.Provider)
			return contentprocessors.Query{}, fmt.Errorf("invalid AI provider: %s", config.AI.Provider)
		}
	} else {
		aiModel = "deepseek"
		log.Printf("ℹ️ No AI provider set, defaulting to deepseek")
	}

	// now we check if have an api key for the provider, if not let the user know and default to env variable
	// we skip ollama as it does not require an api key
	if aiModel != "ollama" {
		if config.AI.APIKey == "" {
			log.Printf("ℹ️ No API key set for AI provider, checking environment variables")

			// we check if the api key is set in the environment variables
			if os.Getenv("DEEPSEEK_API_KEY") != "" && (aiModel == "deepseek" || aiModel == "") {
				log.Printf("✅ Found deepseek API key in environment variable")
				config.AI.APIKey = os.Getenv("DEEPSEEK_API_KEY")
				aiModel = "deepseek"
			} else if os.Getenv("OPENROUTER_API_KEY") != "" && (aiModel == "openrouter" || aiModel == "") {
				log.Printf("✅ Found openrouter API key in environment variable")
				config.AI.APIKey = os.Getenv("OPENROUTER_API_KEY")
				aiModel = "openrouter"
			} else {
				log.Printf("❌ No API key found for %s provider", aiModel)
				return contentprocessors.Query{}, fmt.Errorf("no API key found for provider %s", aiModel)
			}
		}
	}

	// for testing purposes, if the key is "dummy-key", just return a dummy query
	if config.AI.APIKey == "dummy-key" {
		return contentprocessors.Query{}, nil
	}

	// now we switch on the ai model and call the appropriate function
	switch aiModel {
	case "deepseek":
		query, err := SendQueryWithDeepSeek(config, query)
		if err != nil {
			return contentprocessors.Query{}, err
		}
		return query, nil
	case "ollama":
		query, err := SendQueryWithOllama(config, query)
		if err != nil {
			return contentprocessors.Query{}, err
		}
		return query, nil
	case "openrouter":
		query, err := SendQueryWithOpenRouter(config, query)
		if err != nil {
			return contentprocessors.Query{}, err
		}
		return query, nil
	}

	return contentprocessors.Query{}, fmt.Errorf("invalid AI model: %s", aiModel)
}

// SendQueryToLLM sends a query to an LLM API to generate new file names
func SendQueryToLLM(client *deepseek.Client, query contentprocessors.Query, opts QueryOpts) error {
	// load config
	config := utils.LoadConfig(query.ConfigPath)
	performanceOpts := config.Performance.AI
	workers := performanceOpts.Workers
	timeout := performanceOpts.Timeout
	retries := performanceOpts.Retries

	log.Printf("⚙️ AI processing configuration - Workers: %d, Timeout: %s, Retries: %d",
		workers, timeout, retries)

	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		log.Printf("❌ Failed to parse timeout: %v", err)
		return err
	}

	client.Timeout = parsedTimeout

	// Create a semaphore channel to limit concurrent workers
	sem := make(chan struct{}, workers)

	// Iterate through the folders
	for i := range query.Folders {
		folder := &query.Folders[i]
		log.Printf("📁 Processing folder: %s", folder.Name)

		// we create a channel to collect results and errors
		results := make(chan result, len(folder.FileList))

		// Process files concurrently with worker limit
		for j := range folder.FileList {
			// Acquire a worker slot
			sem <- struct{}{}
			go func(j int, file *contentprocessors.File) {
				defer func() { <-sem }() // Release the worker slot when done
				doAI(j, file, opts, query, client, results)
			}(j, &folder.FileList[j])
		}

		// Collect results
		for range folder.FileList {
			res := <-results
			if res.err != nil {
				log.Printf("❌ Failed to process file: %v", res.err)
				continue
			}
			folder.FileList[res.index].NewName = res.name
		}

		// if we have failed files, we retry them with a new worker
		for retries > 0 {
			failedFiles := 0
			for i, file := range folder.FileList {
				if file.NewName == "" {
					failedFiles++
					log.Printf("🔄 Retrying failed file: %s", file.Name)
					sem <- struct{}{} // Acquire a worker slot for retry
					go func(i int, file *contentprocessors.File) {
						defer func() { <-sem }() // Release the worker slot when done
						doAI(i, file, opts, query, client, results)
					}(i, &file)
				}
			}
			if failedFiles == 0 {
				break
			}
			retries--
		}
	}
	// Wait for any remaining workers to finish
	for range workers {
		sem <- struct{}{}
	}
	return nil
}

func doAI(j int, file *contentprocessors.File, opts QueryOpts, query contentprocessors.Query, client *deepseek.Client, results chan result) {
	// Create a chat completion request

	if query.Prompt == "" {
		results <- result{j, "", fmt.Errorf("no prompt provided")}
		return
	}

	request := &deepseek.ChatCompletionRequest{
		Model: opts.Model,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleSystem, Content: query.Prompt},
			{Role: deepseek.ChatMessageRoleUser, Content: file.Context},
		},
	}

	// Send the request and handle the response
	ctx := context.Background()
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		results <- result{j, "", fmt.Errorf("error creating chat completion: %v", err)}
		return
	}

	if response.Choices[0].Message.Content == "" {
		results <- result{j, "", fmt.Errorf("empty response from AI")}
		return
	}

	// Convert the response to the given case in the config
	refinedName := fileutils.RefinedName(response.Choices[0].Message.Content)
	newName := utils.ConvertCase(refinedName, "snake", opts.Case)

	// Remove new lines and spaces from the new name
	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")

	results <- result{j, newName, nil}
}
