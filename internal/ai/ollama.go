package nomnom

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	contentprocessors "nomnom/internal/content"
	fileutils "nomnom/internal/files"
	configutils "nomnom/internal/utils"
	utils "nomnom/internal/utils"

	"github.com/cohesion-org/deepseek-go"
	"github.com/fatih/color"
	api "github.com/ollama/ollama/api"
)

func SendQueryWithOllama(config configutils.Config, query contentprocessors.Query) (q contentprocessors.Query, err error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to create client: %v", err))
		return contentprocessors.Query{}, err
	}

	model := config.AI.Model
	if model == "" {
		fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("No model provided", err))
		return contentprocessors.Query{}, err
	}

	fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.WhiteString("You're using Ollama with model: %s", model))

	performanceOpts := config.Performance.AI
	workers := performanceOpts.Workers
	timeout := performanceOpts.Timeout
	retries := performanceOpts.Retries

	fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("AI processing configuration - Workers: %d, Timeout: %s, Retries: %d",
		workers, timeout, retries))

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
				defer func() { <-sem }()                            // Release worker slot
				doAIOllama(j, file, query, client, results, config) // Process file
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
					doAIOllama(index, file, query, client, retryResults, config)
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
			return contentprocessors.Query{}, err
		}
	}

	wg.Wait()
	return query, nil

}

func removeThink(s string) string {
	// Remove everything between <Think> and </Think> tags
	startTag := "<think>"
	endTag := "</think>"
	result := s

	for {
		startIdx := strings.Index(result, startTag)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(result[startIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + len(endTag)

		result = result[:startIdx] + result[endIdx:]
	}

	// Remove spaces and return
	s = strings.TrimSpace(result)
	s = strings.ReplaceAll(s, "  ", " ")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func createMessage(file contentprocessors.File, vision bool, prompt string, context string) []api.Message {
	if vision && fileutils.IsImageFile(file.Path) {
		imageData, err := deepseek.ImageToBase64(file.Path) // Depseek-go is used to convert image to base64

		if err != nil {
			fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to convert image to base64: %s", err))
			return nil
		}
		base64Str := strings.Split(imageData, ",")[1] // Note to maintainer: Ollama doesn't accpet Base64 header it only accepts the base64 string in []bytes
		bytes, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Error decoding: %v", err))
			return nil
		}
		return []api.Message{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:   "user",
				Images: []api.ImageData{bytes},
				// We don't necessarily need to give content here as ollama will use the image data
				// plus models get confused if you provide both of them.
				// Another note: the checkAndAddExtension function handels the extension even if the AI screws it up.
			},
		}
	}
	return []api.Message{
		{
			Role:    "system",
			Content: prompt,
		},
		{
			Role:    "user",
			Content: context,
		},
	}
}

func doAIOllama(j int, file *contentprocessors.File, query contentprocessors.Query, client *api.Client, results chan Result, config configutils.Config) {
	prompt := config.AI.Prompt
	if prompt == "" {
		prompt = query.Prompt
		if prompt == "" {
			fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("No prompt provided, using default prompt"))
			prompt = "What is the name of this document? Only respond with the name and the extension of the file in snake case. Do not respond with anything else!"
		}
	}

	messages := createMessage(*file, fileutils.IsImageFile(file.Path), prompt, file.Context)
	if len(messages) == 0 {
		fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to create message for file %s", file.Name))
		results <- Result{j, "", fmt.Errorf("failed to create message for file %s", file.Name)}
		return
	}

	var newName string
	response := func(response api.ChatResponse) error {
		newName = removeThink(response.Message.Content)
		newName = fileutils.CheckAndAddExtension(newName, file.Name)
		return nil
	}

	stream := false
	err := client.Chat(context.Background(), &api.ChatRequest{
		Model:    config.AI.Model,
		Messages: messages,
		Stream:   &stream,
	}, response)

	if err != nil {
		fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to process file %s: %v", file.Name, err))
		results <- Result{j, "", fmt.Errorf("error creating chat completion: %v", err)}
		return
	}

	if newName == "" {
		results <- Result{j, "", fmt.Errorf("empty response from AI")}
		return
	}

	refinedName := fileutils.RefinedName(newName)

	// check if the response is valid
	isValid, reason := fileutils.IsAValidFileName(refinedName)

	if !isValid {
		file.Context = "This is a retry for this file because it failed file validation last time for the reason: " + reason + "\n" + "Please check the file context and try again." + file.Context
		results <- Result{j, "NOMNOMFAILED", fmt.Errorf("invalid response from AI: %s", reason)}
		return
	}

	newName = utils.ConvertCase(refinedName, "snake", config.Case)

	// Remove new lines and spaces from the new name
	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")
	newName = fileutils.CheckAndAddExtension(newName, file.Name)

	results <- Result{j, newName, nil}

}
