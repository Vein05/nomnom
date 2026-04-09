package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	content "nomnom/internal/content"
	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

type QueryOpts struct {
	Provider    string
	Model       string
	Case        string
	MaxTokens   int
	Temperature float64
}

func HandleAI(config utils.Config, query content.Query) (content.Query, error) {
	reporter := reporterFor(query)
	if config.AI == (utils.AIConfig{}) {
		return content.Query{}, fmt.Errorf("AI configuration is empty")
	}

	provider := config.AI.Provider
	if provider == "" {
		provider = "deepseek"
		reporter.Infof("No AI provider set, defaulting to deepseek")
	}
	if provider != "deepseek" && provider != "openrouter" && provider != "ollama" {
		return content.Query{}, fmt.Errorf("invalid AI provider: %s", provider)
	}

	if provider != "ollama" && config.AI.APIKey == "" {
		switch provider {
		case "deepseek":
			config.AI.APIKey = os.Getenv("DEEPSEEK_API_KEY")
		case "openrouter":
			config.AI.APIKey = os.Getenv("OPENROUTER_API_KEY")
		}
	}

	if provider != "ollama" && config.AI.APIKey == "" {
		return content.Query{}, fmt.Errorf("no API key found for provider %s", provider)
	}

	if config.AI.APIKey == "dummy-key" {
		return query, nil
	}

	switch provider {
	case "deepseek":
		return SendQueryWithDeepSeek(config, query)
	case "ollama":
		return SendQueryWithOllama(config, query)
	case "openrouter":
		return SendQueryWithOpenRouter(config, query)
	default:
		return content.Query{}, fmt.Errorf("invalid AI provider: %s", provider)
	}
}

func SendQueryToLLM(client *deepseek.Client, config utils.Config, query content.Query, opts QueryOpts) error {
	if client == nil {
		return fmt.Errorf("nil client")
	}
	if len(query.Scan.Files) == 0 {
		return fmt.Errorf("no files to process")
	}

	workers, retries, timeout, err := aiRuntime(config)
	if err != nil {
		return err
	}

	client.Timeout = timeout
	reporter := reporterFor(query)
	reporter.Infof("AI processing configuration - Workers: %d, Timeout: %s, Retries: %d", workers, timeout, retries)

	plan := buildRenamePlan(query.Scan.Files, workers, retries, reporter, func(file content.ScannedFile, retryHint string) (string, error) {
		if config.AI.Vision.Enabled && hasVisionSource(file) {
			return requestVisionName(client, query.Prompt, file, retryHint, opts, query.Analytics)
		}
		return requestTextName(client, query.Prompt, file, retryHint, opts, query.Analytics)
	})

	query.Plan = plan
	return nil
}

func aiRuntime(config utils.Config) (workers int, retries int, timeout time.Duration, err error) {
	workers = config.Performance.AI.Workers
	if workers == 0 {
		workers = 1
	}

	retries = config.Performance.AI.Retries
	if retries == 0 {
		retries = 1
	}

	timeoutRaw := config.Performance.AI.Timeout
	if timeoutRaw == "" {
		timeoutRaw = "30s"
	}

	timeout, err = time.ParseDuration(timeoutRaw)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse timeout: %w", err)
	}

	return workers, retries, timeout, nil
}

func buildRenamePlan(files []content.ScannedFile, workers, retries int, reporter utils.Reporter, nameFunc func(content.ScannedFile, string) (string, error)) []content.RenamePlanEntry {
	results := make([]content.RenamePlanEntry, len(files))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for index, file := range files {
		index := index
		file := file
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[index] = content.RenamePlanEntry{
				File:          file,
				SuggestedName: nameWithRetry(file, retries, reporter, nameFunc),
			}
		}()
	}

	wg.Wait()
	return results
}

func nameWithRetry(file content.ScannedFile, retries int, reporter utils.Reporter, nameFunc func(content.ScannedFile, string) (string, error)) string {
	retryHint := ""
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		name, err := nameFunc(file, retryHint)
		if err == nil {
			return name
		}

		lastErr = err
		retryHint = retryReason(err)
		if attempt < retries {
			reporter.Warnf("Retry attempt %d/%d for %s", attempt+1, retries, file.OriginalName)
		}
	}

	reporter.Errorf("Failed to process file: %s. Error: %v", file.OriginalName, lastErr)
	return ""
}

func requestTextName(client *deepseek.Client, prompt string, file content.ScannedFile, retryHint string, opts QueryOpts, analytics *utils.AnalyticsStore) (string, error) {
	request := &deepseek.ChatCompletionRequest{
		Model: opts.Model,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleSystem, Content: prompt},
			{Role: deepseek.ChatMessageRoleUser, Content: promptContext(file, retryHint)},
		},
	}

	response, err := client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}
	if response.Choices == nil || len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in AI response")
	}
	recordAnalyticsUsage(analytics, opts.Provider, response.Model, response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens, false)
	return normalizeSuggestedName(response.Choices[0].Message.Content, file, opts.Case)
}

func requestVisionName(client *deepseek.Client, prompt string, file content.ScannedFile, retryHint string, opts QueryOpts, analytics *utils.AnalyticsStore) (string, error) {
	base64Image, err := deepseek.ImageToBase64(visionSourcePath(file))
	if err != nil {
		return "", fmt.Errorf("error opening image file: %w", err)
	}

	request := &deepseek.ChatCompletionRequestWithImage{
		Model: opts.Model,
		Messages: []deepseek.ChatCompletionMessageWithImage{
			{Role: deepseek.ChatMessageRoleSystem, Content: prompt},
			deepseek.NewImageMessage("user", promptContext(file, retryHint), base64Image),
		},
	}

	response, err := client.CreateChatCompletionWithImage(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}
	if response.Choices == nil || len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in AI response")
	}
	recordAnalyticsUsage(analytics, opts.Provider, response.Model, response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens, true)
	return normalizeSuggestedName(response.Choices[0].Message.Content, file, opts.Case)
}

func normalizeSuggestedName(raw string, file content.ScannedFile, caseStyle string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("empty response from AI")
	}

	refinedName := fileutils.RefinedName(raw)
	newName := utils.ConvertCase(refinedName, "snake", caseStyle)
	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")
	newName = fileutils.CheckAndAddExtension(newName, file.OriginalName)

	if isValid, reason := fileutils.IsAValidFileName(newName); !isValid {
		return "", fmt.Errorf("invalid response from AI: %s", reason)
	}

	return newName, nil
}

func promptContext(file content.ScannedFile, retryHint string) string {
	if retryHint == "" {
		return file.Context
	}

	return "Previous filename suggestion failed validation for this reason: " + retryHint + "\nPlease return only a valid filename with the original extension.\n\n" + file.Context
}

func retryReason(err error) string {
	message := err.Error()
	const prefix = "invalid response from AI: "
	if strings.HasPrefix(message, prefix) {
		return strings.TrimPrefix(message, prefix)
	}
	return ""
}

func reporterFor(query content.Query) utils.Reporter {
	if query.Reporter != nil {
		return query.Reporter
	}
	return utils.NopReporter{}
}

func hasVisionSource(file content.ScannedFile) bool {
	return file.VisualPath != "" || fileutils.IsImageFile(file.SourcePath)
}

func visionSourcePath(file content.ScannedFile) string {
	if file.VisualPath != "" {
		return file.VisualPath
	}
	return file.SourcePath
}

func recordAnalyticsUsage(analytics *utils.AnalyticsStore, provider, model string, promptTokens, completionTokens, totalTokens int, vision bool) {
	if analytics == nil {
		return
	}

	analytics.RecordAIUsage(utils.AnalyticsUsage{
		Provider:         provider,
		Model:            model,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		Vision:           vision,
	})
}
