// Package ai provides a collection of AI models and functions for use in the NomNom project.
package nomnom

import (
	"context"
	"fmt"
	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"
	"strings"

	utils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// QueryOpts contains options for the query
type QueryOpts struct {
	Model string
	Case  string
}

const prompt string = `You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension.`

// HandleAI is a function that handles the AI model selection and query execution and returns the result.
func HandleAI(config configutils.Config, query contentprocessors.Query) (contentprocessors.Query, error) {
	// Select the AI model based on the config
	var aiModel string

	// we first check if the provider is set, if not we default to deepseek
	// we currently check if we are serving deepseek, ollama or openrouter
	if config.AI.Provider != "" {
		if config.AI.Provider == "deepseek" {
			aiModel = "deepseek"
		} else if config.AI.Provider == "ollama" {
			aiModel = "ollama"
		} else if config.AI.Provider == "openrouter" {
			aiModel = "openrouter"
		}
	} else {
		aiModel = "deepseek"
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

	// Iterate through the folders
	for i := range query.Folders {
		folder := &query.Folders[i]
		fmt.Println("Working on Folder Name:", folder.Name)

		// Iterate through the files in the folder
		for j := range folder.FileList {
			file := &folder.FileList[j]
			fmt.Println("Working on File Name:", file.Name)

			// Create a chat completion request
			request := &deepseek.ChatCompletionRequest{
				Model: opts.Model,
				Messages: []deepseek.ChatCompletionMessage{
					{Role: deepseek.ChatMessageRoleSystem, Content: prompt},
					{Role: deepseek.ChatMessageRoleUser, Content: file.Context},
				},
			}

			// Send the request and handle the response
			ctx := context.Background()
			response, err := client.CreateChatCompletion(ctx, request)
			if err != nil {
				return fmt.Errorf("error creating chat completion: %v", err)
			}
			// add a check to see if the response is empty
			if response.Choices[0].Message.Content == "" {
				return fmt.Errorf("empty response from AI")
			}

			// convert the response to the given case in the config
			newName := utils.ConvertCase(response.Choices[0].Message.Content, "snake", opts.Case)

			// remove new lines and spaces from the new name
			newName = strings.ReplaceAll(newName, "\n", "")
			newName = strings.ReplaceAll(newName, " ", "")
			file.NewName = newName
		}
	}
	return nil
}
