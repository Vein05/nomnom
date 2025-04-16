// Package ai provides a collection of AI models and functions for use in the NomNom project.
package nomnom

import (
	"context"
	"fmt"
	"sync"
	"time"

	contentprocessors "nomnom/internal/content"
	"os"
	"strings"

	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/fatih/color"
)

// QueryOpts contains options for the query
type QueryOpts struct {
	Model       string
	Case        string
	MaxTokens   int
	Temperature float64
}

type Result struct {
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
			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("Using deepseek as AI provider"))
			aiModel = "deepseek"
		} else if config.AI.Provider == "ollama" {
			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("Using ollama as AI provider"))
			aiModel = "ollama"
		} else if config.AI.Provider == "openrouter" {
			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("Using openrouter as AI provider"))
			aiModel = "openrouter"
		} else {
			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("Invalid AI provider: %s", config.AI.Provider))
			return contentprocessors.Query{}, fmt.Errorf("invalid AI provider: %s", config.AI.Provider)
		}
	} else {
		aiModel = "deepseek"
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("No AI provider set, defaulting to deepseek"))
	}

	// now we check if have an api key for the provider, if not let the user know and default to env variable
	// we skip ollama as it does not require an api key
	if aiModel != "ollama" {
		if config.AI.APIKey == "" {
			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.YellowString("No API key set for AI provider, checking environment variables"))
			// we check if the api key is set in the environment variables
			if os.Getenv("DEEPSEEK_API_KEY") != "" && (aiModel == "deepseek" || aiModel == "") {
				fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("Found deepseek API key in environment variable"))
				config.AI.APIKey = os.Getenv("DEEPSEEK_API_KEY")
				aiModel = "deepseek"
			} else if os.Getenv("OPENROUTER_API_KEY") != "" && (aiModel == "openrouter" || aiModel == "") {
				fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("Found openrouter API key in environment variable"))
				config.AI.APIKey = os.Getenv("OPENROUTER_API_KEY")
				aiModel = "openrouter"
			} else {
				fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("No API key found for %s provider", aiModel))
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
	config := utils.LoadConfig(query.ConfigPath, color.WhiteString("▶  "))
	performanceOpts := config.Performance.AI
	workers := performanceOpts.Workers
	timeout := performanceOpts.Timeout
	retries := performanceOpts.Retries

	fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("AI processing configuration - Workers: %d, Timeout: %s, Retries: %d",
		workers, timeout, retries))

	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("Failed to parse timeout: %v", err))
		return err
	}

	client.Timeout = parsedTimeout

	if config.AI.Vision.Enabled {
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.YellowString("You're using the vision mode, please make sure the model you're using is multimodal"))
	}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a recursive function to process folders
	var processFolder func(folder *contentprocessors.FolderType) error
	processFolder = func(folder *contentprocessors.FolderType) error {
		// Create channels for the current folder's processing
		sem := make(chan struct{}, workers)
		results := make(chan Result, len(folder.FileList))

		// Process files in current folder
		for j := range folder.FileList {
			sem <- struct{}{} // Acquire worker slot
			wg.Add(1)
			go func(j int, file *contentprocessors.File) {
				defer wg.Done()
				defer func() { <-sem }()                                          // Release worker slot
				if config.AI.Vision.Enabled && fileutils.IsImageFile(file.Path) { // Note: Trust the user to set the right config
					doVisionAI(j, file, opts, query, client, results) // Note: We don't have any checks for failures here because our failure is handled in our retry logic with defaults to doAI()
				} else {
					doAI(j, file, opts, query, client, results)
				}
			}(j, &folder.FileList[j])
		}

		// Collect results for current folder's files
		for range folder.FileList {
			res := <-results
			if res.err != nil {
				fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("Failed to process file: %s. Error: %v",
					folder.FileList[res.index].Name, res.err))
				folder.FileList[res.index].NewName = res.name
				continue
			}
			folder.FileList[res.index].NewName = res.name
		}

		// Handle retries for failed files in current folder
		for retryAttempt := range retries {
			failedIndices := []int{}

			// Identify failed files
			for i, file := range folder.FileList {
				if file.NewName == "NOMNOMFAILED" {
					failedIndices = append(failedIndices, i)
				}
			}

			if len(failedIndices) == 0 {
				break
			}

			fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.YellowString("Retry attempt %d/%d for %d files",
				retryAttempt+1, retries, len(failedIndices)))

			retryResults := make(chan Result, len(failedIndices))

			// Process failed files
			for _, i := range failedIndices {
				sem <- struct{}{}
				wg.Add(1)
				go func(index int, file *contentprocessors.File) {
					defer wg.Done()
					defer func() { <-sem }()
					doAI(index, file, opts, query, client, retryResults)
				}(i, &folder.FileList[i])
			}

			// Collect retry results
			for range failedIndices {
				res := <-retryResults
				mu.Lock()
				folder.FileList[res.index].NewName = res.name
				mu.Unlock()
			}
		}

		// Process subfolders recursively
		for i := range folder.SubFolders {
			if err := processFolder(&folder.SubFolders[i]); err != nil {
				return err
			}
		}

		return nil
	}

	// Process all root folders
	for i := range query.Folders {
		if err := processFolder(&query.Folders[i]); err != nil {
			return err
		}
	}

	wg.Wait()
	return nil
}

func doAI(j int, file *contentprocessors.File, opts QueryOpts, query contentprocessors.Query, client *deepseek.Client, results chan Result) {
	// Create a chat completion request
	if query.Prompt == "" {
		results <- Result{j, "", fmt.Errorf("no prompt provided")}
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
		results <- Result{j, "", fmt.Errorf("error creating chat completion: %v", err)}
		return
	}

	if response.Choices[0].Message.Content == "" {
		results <- Result{j, "", fmt.Errorf("empty response from AI")}
		return
	}

	refinedName := fileutils.RefinedName(response.Choices[0].Message.Content)

	// check if the response is valid
	isValid, reason := fileutils.IsAValidFileName(refinedName)

	if !isValid {
		file.Context = "This is a retry for this file because it failed file validation last time for the reason: " + reason + "\n" + "Please check the file context and try again." + file.Context
		results <- Result{j, "NOMNOMFAILED", fmt.Errorf("invalid response from AI: %s", reason)}
		return
	}

	newName := utils.ConvertCase(refinedName, "snake", opts.Case)

	// Remove new lines and spaces from the new name
	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")
	newName = fileutils.CheckAndAddExtension(newName, file.Name)

	results <- Result{j, newName, nil}
}

func doVisionAI(j int, file *contentprocessors.File, opts QueryOpts, query contentprocessors.Query, client *deepseek.Client, results chan Result) {
	if query.Prompt == "" {
		results <- Result{j, "", fmt.Errorf("no prompt provided")}
		return
	}

	base64Image, err := deepseek.ImageToBase64(file.Path)
	if err != nil {
		results <- Result{j, "", fmt.Errorf("error opening image file: %v", err)}
		return
	}
	request := &deepseek.ChatCompletionRequestWithImage{
		Model: opts.Model,
		Messages: []deepseek.ChatCompletionMessageWithImage{
			{Role: deepseek.ChatMessageRoleSystem, Content: query.Prompt},
			deepseek.NewImageMessage("user", file.Context, base64Image),
		},
	}

	// Send the request and handle the response
	ctx := context.Background()
	response, err := client.CreateChatCompletionWithImage(ctx, request)
	if err != nil {
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("Error creating chat completion for %s: will get added to retry!", file.Name))
		results <- Result{j, "", fmt.Errorf("error creating chat completion: %v", err)}
		return
	}

	if response.Choices[0].Message.Content == "" {
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.RedString("Error creating chat completion for %s: will get added to retry!", file.Name))
		results <- Result{j, "", fmt.Errorf("empty response from AI")}
		return
	}

	refinedName := fileutils.RefinedName(response.Choices[0].Message.Content)

	newName := utils.ConvertCase(refinedName, "snake", opts.Case)

	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")
	newName = fileutils.CheckAndAddExtension(newName, file.Name)

	results <- Result{j, newName, nil}
}
