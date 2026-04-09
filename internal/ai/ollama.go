package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	content "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	"github.com/cohesion-org/deepseek-go"
	api "github.com/ollama/ollama/api"
)

func SendQueryWithOllama(config configutils.Config, query content.Query) (content.Query, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return content.Query{}, fmt.Errorf("failed to create client: %w", err)
	}
	if config.AI.Model == "" {
		return content.Query{}, fmt.Errorf("no model provided")
	}
	if len(query.Scan.Files) == 0 {
		return content.Query{}, fmt.Errorf("no files to process")
	}

	workers := config.Performance.AI.Workers
	if workers == 0 {
		workers = 1
	}
	retries := config.Performance.AI.Retries
	if retries == 0 {
		retries = 1
	}

	reporterFor(query).Infof("You're using Ollama with model: %s", config.AI.Model)
	reporterFor(query).Infof("AI processing configuration - Workers: %d, Retries: %d", workers, retries)

	query.Plan = buildRenamePlan(query.Scan.Files, workers, retries, reporterFor(query), func(file content.ScannedFile, retryHint string) (string, error) {
		return requestOllamaName(client, config, query, file, retryHint)
	})

	return query, nil
}

func removeThink(s string) string {
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

	result = strings.TrimSpace(result)
	result = strings.ReplaceAll(result, "  ", " ")
	result = strings.ReplaceAll(result, "\n", "")
	return result
}

func requestOllamaName(client *api.Client, config configutils.Config, query content.Query, file content.ScannedFile, retryHint string) (string, error) {
	prompt := config.AI.Prompt
	if prompt == "" {
		prompt = query.Prompt
	}
	if prompt == "" {
		prompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."
	}

	messages, err := createOllamaMessages(file, config.AI.Vision.Enabled && hasVisionSource(file), prompt, promptContext(file, retryHint))
	if err != nil {
		return "", err
	}

	var newName string
	var lastResponse api.ChatResponse
	stream := false
	err = client.Chat(context.Background(), &api.ChatRequest{
		Model:    config.AI.Model,
		Messages: messages,
		Stream:   &stream,
	}, func(response api.ChatResponse) error {
		lastResponse = response
		newName = removeThink(response.Message.Content)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	modelName := lastResponse.Model
	if modelName == "" {
		modelName = config.AI.Model
	}

	recordAnalyticsUsage(
		query.Analytics,
		"ollama",
		modelName,
		lastResponse.PromptEvalCount,
		lastResponse.EvalCount,
		lastResponse.PromptEvalCount+lastResponse.EvalCount,
		config.AI.Vision.Enabled && hasVisionSource(file),
	)

	return normalizeSuggestedName(newName, file, config.Case)
}

func createOllamaMessages(file content.ScannedFile, vision bool, prompt string, context string) ([]api.Message, error) {
	if !vision {
		return []api.Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: context},
		}, nil
	}

	imageData, err := deepseek.ImageToBase64(visionSourcePath(file))
	if err != nil {
		return nil, fmt.Errorf("failed to convert image to base64: %w", err)
	}

	base64Str := strings.Split(imageData, ",")[1]
	bytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	return []api.Message{
		{Role: "system", Content: prompt},
		{Role: "user", Images: []api.ImageData{bytes}, Content: context},
	}, nil
}
