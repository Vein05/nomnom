package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	content "nomnom/internal/content"
	utils "nomnom/internal/utils"
)

const defaultPromptText = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."

var ollamaTagsURL = "http://localhost:11434/api/tags"

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func defaultConfigPath() string {
	path, err := utils.DefaultConfigPath()
	if err != nil {
		return "nomnom-config.json"
	}
	return path
}

func defaultDesktopConfig() DesktopConfig {
	return desktopConfigFromCore(utils.DefaultConfig())
}

func desktopConfigFromCore(config utils.Config) DesktopConfig {
	desktop := DesktopConfig{
		Output: config.Output,
		Case:   config.Case,
		AI: AIConfig{
			Provider:    config.AI.Provider,
			Model:       config.AI.Model,
			APIKey:      config.AI.APIKey,
			MaxTokens:   config.AI.MaxTokens,
			Temperature: config.AI.Temperature,
			Vision: VisionConfig{
				Enabled:      config.AI.Vision.Enabled,
				MaxImageSize: config.AI.Vision.MaxImageSize,
			},
			Prompt: config.AI.Prompt,
		},
		FileHandling: FileHandlingConfig{
			MaxSize:      config.FileHandling.MaxSize,
			AutoApprove:  config.FileHandling.AutoApprove,
			HotRename:    config.FileHandling.HotRename,
			SkipDotFiles: config.FileHandling.SkipDotFiles,
		},
		ContentExtraction: ContentExtractionConfig{
			ExtractText:      config.ContentExtraction.ExtractText,
			ExtractMetadata:  config.ContentExtraction.ExtractMetadata,
			MaxContentLength: config.ContentExtraction.MaxContentLength,
			SkipLargeFiles:   config.ContentExtraction.SkipLargeFiles,
			ReadContext:      config.ContentExtraction.ReadContext,
		},
		Performance: PerformanceConfig{
			AI: PerformancePipelineConfig{
				Workers: config.Performance.AI.Workers,
				Timeout: config.Performance.AI.Timeout,
				Retries: config.Performance.AI.Retries,
			},
			File: PerformancePipelineConfig{
				Workers: config.Performance.File.Workers,
				Timeout: config.Performance.File.Timeout,
				Retries: config.Performance.File.Retries,
			},
		},
		Logging: LoggingConfig{
			Enabled: config.Logging.Enabled,
			LogPath: config.Logging.LogPath,
		},
	}

	if strings.TrimSpace(desktop.AI.Prompt) == "" {
		desktop.AI.Prompt = defaultPromptText
	}

	return desktop
}

func coreConfigFromDesktop(config DesktopConfig) utils.Config {
	core := utils.DefaultConfig()
	core.Output = config.Output
	core.Case = config.Case
	core.AI.Provider = config.AI.Provider
	core.AI.Model = config.AI.Model
	core.AI.APIKey = config.AI.APIKey
	core.AI.MaxTokens = config.AI.MaxTokens
	core.AI.Temperature = config.AI.Temperature
	core.AI.Vision.Enabled = config.AI.Vision.Enabled
	core.AI.Vision.MaxImageSize = config.AI.Vision.MaxImageSize
	core.AI.Prompt = strings.TrimSpace(config.AI.Prompt)
	core.FileHandling.MaxSize = config.FileHandling.MaxSize
	core.FileHandling.AutoApprove = config.FileHandling.AutoApprove
	core.FileHandling.HotRename = config.FileHandling.HotRename
	core.FileHandling.SkipDotFiles = config.FileHandling.SkipDotFiles
	core.ContentExtraction.ExtractText = config.ContentExtraction.ExtractText
	core.ContentExtraction.ExtractMetadata = config.ContentExtraction.ExtractMetadata
	core.ContentExtraction.MaxContentLength = config.ContentExtraction.MaxContentLength
	core.ContentExtraction.SkipLargeFiles = config.ContentExtraction.SkipLargeFiles
	core.ContentExtraction.ReadContext = config.ContentExtraction.ReadContext
	core.Performance.AI.Workers = config.Performance.AI.Workers
	core.Performance.AI.Timeout = config.Performance.AI.Timeout
	core.Performance.AI.Retries = config.Performance.AI.Retries
	core.Performance.File.Workers = config.Performance.File.Workers
	core.Performance.File.Timeout = config.Performance.File.Timeout
	core.Performance.File.Retries = config.Performance.File.Retries
	core.Logging.Enabled = config.Logging.Enabled
	core.Logging.LogPath = config.Logging.LogPath

	if core.AI.Prompt == "" {
		core.AI.Prompt = defaultPromptText
	}

	return core
}

func loadDesktopConfig(path string) (DesktopConfig, error) {
	resolvedPath, err := normalizeConfigPath(path)
	if err != nil {
		return DesktopConfig{}, err
	}

	config := defaultDesktopConfig()
	coreConfig, err := utils.LoadConfig(resolvedPath)
	if err == nil {
		return desktopConfigFromCore(coreConfig), nil
	}
	if os.IsNotExist(err) || strings.Contains(err.Error(), "config file not found") {
		return config, nil
	}
	return DesktopConfig{}, err
}

func resolveOpenRouterAPIKey(explicit string) (string, string) {
	trimmed := strings.TrimSpace(explicit)
	if trimmed != "" {
		return trimmed, "config"
	}

	if envKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")); envKey != "" {
		return envKey, "env"
	}

	return "", "missing"
}

func probeOllamaStatus(selectedModel string) OllamaStatus {
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get(ollamaTagsURL)
	if err != nil {
		return OllamaStatus{Running: false, ModelAvailable: false, Message: "Ollama is not running"}
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return OllamaStatus{Running: false, ModelAvailable: false, Message: fmt.Sprintf("Ollama returned %s", resp.Status)}
	}

	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return OllamaStatus{Running: true, ModelAvailable: false, Message: "Ollama is running"}
	}

	for _, model := range payload.Models {
		if ollamaModelMatches(selectedModel, model.Name) {
			return OllamaStatus{Running: true, ModelAvailable: true, Message: "Ollama is running and the model is available"}
		}
	}

	return OllamaStatus{Running: true, ModelAvailable: false, Message: "Ollama is running but the model is not available"}
}

func ollamaModelMatches(selected, available string) bool {
	selected = strings.TrimSpace(selected)
	available = strings.TrimSpace(available)
	if selected == "" || available == "" {
		return false
	}
	if strings.EqualFold(selected, available) {
		return true
	}
	if strings.HasPrefix(available, selected+":") {
		return true
	}
	if strings.HasPrefix(selected, available+":") {
		return true
	}
	return false
}

func seedFileDialog(path string) (string, string) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", ""
	}

	resolvedPath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", ""
	}

	if info, statErr := os.Stat(resolvedPath); statErr == nil && info.IsDir() {
		return resolvedPath, "config.json"
	}

	return filepath.Dir(resolvedPath), filepath.Base(resolvedPath)
}

func normalizeConfigPath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return defaultConfigPath(), nil
	}

	resolvedPath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	if info, statErr := os.Stat(resolvedPath); statErr == nil && info.IsDir() {
		return "", fmt.Errorf("config path must be a file")
	}

	return resolvedPath, nil
}

func newJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func slugify(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = slugPattern.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	return value
}

func mapScannedFiles(files []content.ScannedFile) []RenameEntry {
	entries := make([]RenameEntry, 0, len(files))
	for index, file := range files {
		fileType := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.OriginalName)), ".")
		if fileType == "" {
			fileType = strings.TrimPrefix(strings.ToLower(file.Extension), ".")
		}
		entries = append(entries, RenameEntry{
			Index:     index + 1,
			Original:  file.RelativePath,
			NewName:   file.OriginalName,
			Type:      fileType,
			Status:    "Pending",
			SizeBytes: file.Size,
		})
	}
	return entries
}

func mapRenamePlan(plan []content.RenamePlanEntry) []RenameEntry {
	entries := make([]RenameEntry, 0, len(plan))
	for index, entry := range plan {
		fileType := strings.TrimPrefix(strings.ToLower(filepath.Ext(entry.File.OriginalName)), ".")
		if fileType == "" {
			fileType = strings.TrimPrefix(strings.ToLower(entry.File.Extension), ".")
		}
		entries = append(entries, RenameEntry{
			Index:     index + 1,
			Original:  entry.File.RelativePath,
			NewName:   entry.SuggestedName,
			Type:      fileType,
			Status:    "Pending",
			SizeBytes: entry.File.Size,
		})
	}
	return entries
}

func cloneRenamePlan(plan []content.RenamePlanEntry) []content.RenamePlanEntry {
	cloned := make([]content.RenamePlanEntry, len(plan))
	copy(cloned, plan)
	return cloned
}

func planSummary(planned int, results []content.ProcessResult) JobSummary {
	summary := JobSummary{Planned: planned}
	for _, result := range results {
		if result.OriginalPath == "" && result.NewPath == "" && result.Error == nil {
			continue
		}
		if result.Success {
			summary.Renamed++
			continue
		}
		summary.Errors++
	}
	summary.Skipped = max(planned-summary.Renamed-summary.Errors, 0)
	return summary
}

type desktopHistoryRecord struct {
	Date      time.Time `json:"date"`
	Directory string    `json:"directory"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	DryRun    bool      `json:"dry_run"`
	Status    string    `json:"status"`
	Planned   int       `json:"planned"`
	Renamed   int       `json:"renamed"`
	Skipped   int       `json:"skipped"`
	Errors    int       `json:"errors"`
	Tokens    int       `json:"tokens"`
	LogPath   string    `json:"log_path,omitempty"`
}

func historyStorePath(configPath string) string {
	dir := filepath.Dir(configPath)
	return filepath.Join(dir, "desktop-history.json")
}

func loadHistoryRecords(configPath string) ([]desktopHistoryRecord, error) {
	path := historyStorePath(configPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []desktopHistoryRecord{}, nil
		}
		return nil, fmt.Errorf("read desktop history: %w", err)
	}

	var records []desktopHistoryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse desktop history: %w", err)
	}

	slices.SortFunc(records, func(a, b desktopHistoryRecord) int {
		if a.Date.Equal(b.Date) {
			return 0
		}
		if a.Date.After(b.Date) {
			return -1
		}
		return 1
	})

	return records, nil
}

func appendHistoryRecord(configPath string, record desktopHistoryRecord) error {
	records, err := loadHistoryRecords(configPath)
	if err != nil {
		return err
	}

	records = append([]desktopHistoryRecord{record}, records...)

	path := historyStorePath(configPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create desktop history directory: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal desktop history: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write desktop history: %w", err)
	}

	return nil
}

func historySession(record desktopHistoryRecord) Session {
	filesText := fmt.Sprintf("%d renamed", record.Renamed)
	if record.Skipped > 0 {
		filesText = fmt.Sprintf("%s · %d skipped", filesText, record.Skipped)
	}
	if record.Errors > 0 {
		filesText = fmt.Sprintf("%s · %d errors", filesText, record.Errors)
	}

	mode := "Applied"
	if record.DryRun {
		mode = "Dry Run"
	}

	model := record.Model
	if record.Provider != "" {
		model = fmt.Sprintf("%s / %s", record.Provider, record.Model)
	}

	return Session{
		Date:      record.Date.Format("Jan 2, 2006 · 15:04"),
		Directory: record.Directory,
		Files:     filesText,
		Model:     model,
		Mode:      mode,
		Status:    record.Status,
	}
}

func analyticsSummary(records []desktopHistoryRecord) AnalyticsSummary {
	summary := AnalyticsSummary{
		Sessions:   len(records),
		RecentRuns: min(5, len(records)),
	}

	models := make(map[string]struct{})
	for _, record := range records {
		summary.Renamed += record.Renamed
		summary.Tokens += record.Tokens
		if record.Provider != "" || record.Model != "" {
			models[record.Provider+":"+record.Model] = struct{}{}
		}
	}

	if len(records) > 0 {
		summary.AvgPerRun = float64(summary.Renamed) / float64(len(records))
	}
	summary.UniqueModels = len(models)

	recent := records
	if len(recent) > 7 {
		recent = recent[len(recent)-7:]
	}
	summary.RecentSessions = make([]AnalyticsSessionPoint, len(recent))
	for i, r := range recent {
		summary.RecentSessions[i] = AnalyticsSessionPoint{
			Date:    r.Date.Format("Jan 2"),
			Renamed: r.Renamed,
		}
	}

	return summary
}

func latestAnalyticsSession(baseDir string) (utils.SessionAnalytics, error) {
	paths, err := utils.ListAnalyticsSessions(baseDir)
	if err != nil {
		return utils.SessionAnalytics{}, err
	}
	if len(paths) == 0 {
		return utils.SessionAnalytics{}, os.ErrNotExist
	}

	latest := paths[len(paths)-1]
	return utils.LoadAnalyticsSession(latest)
}
