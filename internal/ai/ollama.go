package nomnom

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	contentprocessors "nomnom/internal/content"
	fileutils "nomnom/internal/files"
	configutils "nomnom/internal/utils"

	"github.com/cohesion-org/deepseek-go"
	"github.com/fatih/color"
	api "github.com/ollama/ollama/api"
)

func SendQueryWithOllama(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to create client: %v", err))
		return contentprocessors.Query{}, err
	}

	model := config.AI.Model
	if model == "" {
		return contentprocessors.Query{}, fmt.Errorf("Ollama model not specified. Quiting the program!")
	}

	prompt := config.AI.Prompt
	if prompt == "" {
		prompt = query.Prompt
		if prompt == "" {
			prompt = "What is the name of this document? Only respond with the name and the extension of the file in snake case. Do not respond with anything else!"
		}
	}

	fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.WhiteString("You're using Ollama with model: %s", model))

	// performanceOpts := config.Performance.AI
	// workers := performanceOpts.Workers
	// timeout := performanceOpts.Timeout
	// retries := performanceOpts.Retries

	// fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.GreenString("AI processing configuration - Workers: %d, Timeout: %s, Retries: %d",
	// 	workers, timeout, retries))

	fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.YellowString("Parallel processing doesn't work with Ollama yet."))

	// Process each folder's files
	// This doesn't go through every sub folder and such.
	// Todo: Add a recursive function to process directories for this
	for i := range query.Folders {
		folder := &query.Folders[i]
		fmt.Printf("%s %s\n", color.WhiteString("▶ "), color.WhiteString("Processing folder: %s", folder.Name))

		// Process each file in the folder
		for j := range folder.FileList {
			file := &folder.FileList[j]

			messages := createMessage(*file, fileutils.IsImageFile(file.Path), prompt, file.Context)
			if len(messages) == 0 {
				fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to create message for file %s", file.Name))
				continue
			}

			var newName string
			resp := func(response api.ChatResponse) error {
				newName = removeThink(response.Message.Content)
				newName = fileutils.CheckAndAddExtension(newName, file.Name)
				return nil
			}

			stream := false
			err := client.Chat(context.Background(), &api.ChatRequest{
				Model:    config.AI.Model,
				Messages: messages,
				Stream:   &stream,
			}, resp)

			if err != nil {
				fmt.Printf("%s %s\n", color.RedString("❌"), color.RedString("Failed to process file %s: %v", file.Name, err))
				continue
			}

			file.NewName = newName
		}
	}

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
