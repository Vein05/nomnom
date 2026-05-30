package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	appsvc "nomnom/internal/app"
	content "nomnom/internal/content"
	utils "nomnom/internal/utils"

	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type jobState struct {
	status          JobStatus
	plan            []RenameEntry
	planned         []content.RenamePlanEntry
	scannedFiles    []content.ScannedFile
	dir             string
	outputDir       string
	configPath      string
	previewTokens   int
	cancel          context.CancelFunc
	cancelRequested bool
}

type desktopApprover struct{}

func (desktopApprover) Approve(string, string, string) (utils.ApprovalDecision, error) {
	return utils.ApprovalYes, nil
}

type jobReporter struct {
	app   *App
	jobID string
}

func (r jobReporter) Infof(format string, args ...any) {
	r.app.updateJobMessage(r.jobID, fmt.Sprintf(format, args...))
}

func (r jobReporter) Warnf(format string, args ...any) {
	r.app.updateJobMessage(r.jobID, fmt.Sprintf(format, args...))
}

func (r jobReporter) Errorf(format string, args ...any) {
	r.app.updateJobMessage(r.jobID, fmt.Sprintf(format, args...))
}

func (r jobReporter) ReportProgress(done, total int, currentFile string) {
	r.app.updateJobProgress(r.jobID, done, total, currentFile)
}

type App struct {
	ctx             context.Context
	mu              sync.RWMutex
	jobs            map[string]*jobState
	configPath      string
	lastHistoryErr  string
}

func NewApp() *App {
	return &App{
		jobs:       make(map[string]*jobState),
		configPath: defaultConfigPath(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFolder(defaultDirectory string) (string, error) {
	options := runtime.OpenDialogOptions{
		Title:                      "Select source directory",
		ShowHiddenFiles:            true,
		CanCreateDirectories:       false,
		TreatPackagesAsDirectories: true,
	}

	if defaultDirectory != "" {
		if absPath, err := filepath.Abs(defaultDirectory); err == nil {
			if info, statErr := os.Stat(absPath); statErr == nil && info.IsDir() {
				options.DefaultDirectory = absPath
			}
		}
	}

	selectedPath, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return selectedPath, nil
}

func (a *App) SelectConfigFile(defaultPath string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app context not ready")
	}

	options := runtime.OpenDialogOptions{
		Title:           "Select NomNom config",
		ShowHiddenFiles: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "Config Files (*.json)", Pattern: "*.json"},
		},
	}

	if defaultDirectory, defaultFilename := seedFileDialog(defaultPath); defaultDirectory != "" {
		options.DefaultDirectory = defaultDirectory
		options.DefaultFilename = defaultFilename
	}

	selectedPath, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return selectedPath, nil
}

func (a *App) CreateConfigFile(defaultPath string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app context not ready")
	}

	options := runtime.SaveDialogOptions{
		Title:                "Create NomNom config",
		DefaultFilename:      "config.json",
		ShowHiddenFiles:      true,
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "Config Files (*.json)", Pattern: "*.json"},
		},
	}

	if defaultDirectory, defaultFilename := seedFileDialog(defaultPath); defaultDirectory != "" {
		options.DefaultDirectory = defaultDirectory
		options.DefaultFilename = defaultFilename
	}

	selectedPath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return selectedPath, nil
}

func (a *App) GetConfigPath() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.configPath
}

func (a *App) SetConfigPath(configPath string) (DesktopConfig, error) {
	resolvedPath, err := normalizeConfigPath(configPath)
	if err != nil {
		return DesktopConfig{}, err
	}

	config, err := loadDesktopConfig(resolvedPath)
	if err != nil {
		return DesktopConfig{}, err
	}

	a.mu.Lock()
	a.configPath = resolvedPath
	a.mu.Unlock()

	return config, nil
}

func (a *App) GetConfig() (DesktopConfig, error) {
	return loadDesktopConfig(a.GetConfigPath())
}

func (a *App) SaveConfig(config DesktopConfig) (bool, error) {
	coreConfig := coreConfigFromDesktop(config)
	if _, err := utils.SaveConfig(a.GetConfigPath(), coreConfig); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) GetAIStatus() (AIStatus, error) {
	config, err := a.GetConfig()
	if err != nil {
		return AIStatus{}, err
	}

	status := AIStatus{
		Provider: config.AI.Provider,
		Model:    config.AI.Model,
	}

	if strings.TrimSpace(config.AI.Provider) != "ollama" {
		status.Ollama.Message = "Ollama is not selected"
		return status, nil
	}

	ollamaStatus := probeOllamaStatus(strings.TrimSpace(config.AI.Model))
	status.Ollama = ollamaStatus
	if status.Ollama.Message == "" {
		status.Ollama.Message = "Ollama is selected"
	}

	return status, nil
}

func (a *App) CheckOpenRouterAPIKey() (OpenRouterKeyStatus, error) {
	config, err := a.GetConfig()
	if err != nil {
		return OpenRouterKeyStatus{}, err
	}

	key, source := resolveOpenRouterAPIKey(config.AI.APIKey)
	return OpenRouterKeyStatus{
		Available: key != "",
		Source:    source,
	}, nil
}

func (a *App) TestOpenRouterAPIKey(apiKey string, model string) (OpenRouterTestResult, error) {
	config, err := a.GetConfig()
	if err != nil {
		return OpenRouterTestResult{}, err
	}

	resolvedKey, source := resolveOpenRouterAPIKey(apiKey)
	if strings.TrimSpace(model) == "" {
		model = config.AI.Model
	}
	if strings.TrimSpace(model) == "" {
		model = defaultDesktopConfig().AI.Model
	}
	if resolvedKey == "" {
		return OpenRouterTestResult{
			Ok:      false,
			Source:  source,
			Message: "no OpenRouter API key found in config or OPENROUTER_API_KEY",
		}, nil
	}

	payload, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": "hello world",
		}},
		"max_tokens":  1,
		"temperature": 0,
	})
	if err != nil {
		return OpenRouterTestResult{}, fmt.Errorf("marshal openrouter test request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return OpenRouterTestResult{}, fmt.Errorf("create openrouter test request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+resolvedKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nomnom.local")
	req.Header.Set("X-Title", "NomNom Desktop")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return OpenRouterTestResult{}, fmt.Errorf("openrouter test request failed: %w", err)
	}
	defer resp.Body.Close()

	result := OpenRouterTestResult{
		Ok:         resp.StatusCode == http.StatusOK,
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Source:     source,
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if readErr != nil {
		result.Message = resp.Status
		return result, nil
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = resp.Status
	}
	result.Message = message
	result.Response = message
	if result.Ok {
		result.Message = "OpenRouter returned 200 OK"
	}
	return result, nil
}

func (a *App) ScanDirectory(path string) (string, error) {
	configPath := a.GetConfigPath()
	config, err := loadDesktopConfig(configPath)
	if err != nil {
		config = defaultDesktopConfig()
	}

	coreConfig := coreConfigFromDesktop(config)
	scanResult, err := content.ScanDirectory(path, coreConfig, utils.NopReporter{})
	if err != nil {
		return "", err
	}

	entries := mapScannedFiles(scanResult.Files)
	outputDir := coreConfig.Output
	if outputDir == "" {
		outputDir = filepath.Join(scanResult.RootDir, "nomnom", "renamed")
	}

	jobID := newJobID()
	status := JobStatus{
		JobID:   jobID,
		State:   "files-ready",
		Done:    0,
		Total:   len(entries),
		Message: fmt.Sprintf("%d files found", len(entries)),
		Summary:   JobSummary{Planned: len(entries)},
			OutputDir: outputDir,
	}

	a.mu.Lock()
	a.jobs[jobID] = &jobState{
		status:       status,
		plan:         entries,
		scannedFiles: scanResult.Files,
		dir:          scanResult.RootDir,
		outputDir:    outputDir,
		configPath:   configPath,
	}
	a.mu.Unlock()

	return jobID, nil
}

func (a *App) GetPlan(jobID string) ([]RenameEntry, error) {
	a.mu.RLock()
	job, ok := a.jobs[jobID]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	plan := make([]RenameEntry, len(job.plan))
	copy(plan, job.plan)
	return plan, nil
}

func (a *App) GenerateNames(jobID string) error {
	a.mu.RLock()
	job, ok := a.jobs[jobID]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if len(job.scannedFiles) == 0 {
		return fmt.Errorf("no files to process")
	}

	service := appsvc.NewService()
	run, err := service.PrepareRun(appsvc.RunOptions{
		Dir:         job.dir,
		ConfigPath:  job.configPath,
		AutoApprove: true,
		DryRun:      true,
		Log:         false,
		Organize:    true,
	}, utils.NopReporter{}, desktopApprover{})
	if err != nil {
		return fmt.Errorf("prepare run: %w", err)
	}

	if err := service.GeneratePlan(run); err != nil {
		if closeErr := run.Close(); closeErr != nil {
			a.logErrorf("close after generate plan failure: %v", closeErr)
		}
		return fmt.Errorf("generate names: %w", err)
	}

	entries := mapRenamePlan(run.Query.Plan)
	clonedPlan := cloneRenamePlan(run.Query.Plan)
	_ = run.Close()

	previewTokens := 0
	if analyticsSession, err := latestAnalyticsSession(job.dir); err == nil {
		for _, model := range analyticsSession.Models {
			previewTokens += model.TotalTokens
		}
	}

	a.mu.Lock()
	if j, ok := a.jobs[jobID]; ok {
		j.plan = entries
		j.planned = clonedPlan
		j.previewTokens = previewTokens
		j.status.State = "preview-ready"
		j.status.Total = len(entries)
		j.status.Message = fmt.Sprintf("names generated for %d files", len(entries))
		j.status.Summary = JobSummary{Planned: len(entries)}
	}
	a.mu.Unlock()

	return nil
}

func (a *App) RunJob(jobID string, opts RunJobOptions) (string, error) {
	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if !ok {
		a.mu.Unlock()
		return "", fmt.Errorf("job not found: %s", jobID)
	}
	if job.status.State == "running" {
		a.mu.Unlock()
		return "", fmt.Errorf("job already running: %s", jobID)
	}
	job.status.State = "generating"
	job.status.Message = "generating names"
	needsGenerate := len(job.planned) == 0 && len(job.scannedFiles) > 0
	a.mu.Unlock()

	if needsGenerate {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					a.logErrorf("job %s panicked during generation: %v", jobID, r)
					a.finishJobFailure(jobID, fmt.Errorf("internal panic: %v", r), nil, DesktopConfig{}, opts, "")
				}
			}()
			if err := a.GenerateNames(jobID); err != nil {
				a.finishJobFailure(jobID, err, nil, DesktopConfig{}, opts, "")
				return
			}
			a.mu.Lock()
			job, ok := a.jobs[jobID]
			if !ok {
				a.mu.Unlock()
				return
			}
			if job.cancelRequested {
				job.status.State = "canceled"
				job.status.Message = "job canceled"
				a.mu.Unlock()
				a.finishJobCanceled(jobID, nil, DesktopConfig{}, opts, "")
				return
			}
			a.mu.Unlock()
			a.startJobExecution(jobID, opts)
		}()
		return jobID, nil
	}

	a.mu.Lock()
	if job, ok = a.jobs[jobID]; !ok {
		a.mu.Unlock()
		return "", fmt.Errorf("job lost: %s", jobID)
	}
	parentCtx := a.ctx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	runCtx, cancel := context.WithCancel(parentCtx)
	job.cancel = cancel
	job.status.State = "running"
	job.status.Done = 0
	job.status.Total = len(job.plan)
	job.status.CurrentFile = ""
	job.status.Message = "applying rename plan"
	a.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.logErrorf("job %s panicked: %v", jobID, r)
				a.finishJobFailure(jobID, fmt.Errorf("internal panic: %v", r), nil, DesktopConfig{}, opts, "")
			}
		}()
		a.executeJob(jobID, opts, runCtx, cancel)
	}()
	return jobID, nil
}

func (a *App) startJobExecution(jobID string, opts RunJobOptions) {
	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if !ok {
		a.mu.Unlock()
		return
	}
	parentCtx := a.ctx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	runCtx, cancel := context.WithCancel(parentCtx)
	job.cancel = cancel
	job.status.State = "running"
	job.status.Done = 0
	job.status.Total = len(job.plan)
	job.status.CurrentFile = ""
	job.status.Message = "applying rename plan"
	a.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.logErrorf("job %s panicked: %v", jobID, r)
				a.finishJobFailure(jobID, fmt.Errorf("internal panic: %v", r), nil, DesktopConfig{}, opts, "")
			}
		}()
		a.executeJob(jobID, opts, runCtx, cancel)
	}()
}

func (a *App) CancelJob(jobID string) bool {
	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if !ok {
		a.mu.Unlock()
		return false
	}

	if job.status.State == "generating" {
		job.cancelRequested = true
		job.status.Message = "cancel requested"
		a.mu.Unlock()
		return true
	}

	if job.status.State != "running" || job.cancel == nil {
		a.mu.Unlock()
		return false
	}
	cancel := job.cancel
	job.status.Message = "cancel requested"
	a.mu.Unlock()

	cancel()
	return true
}

func (a *App) GetJobStatus(jobID string) (JobStatus, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	job, ok := a.jobs[jobID]
	if !ok {
		return JobStatus{}, fmt.Errorf("job not found: %s", jobID)
	}
	return job.status, nil
}

func (a *App) GetHistory() ([]Session, error) {
	records, err := loadHistoryRecords(a.GetConfigPath())
	if err != nil {
		return nil, err
	}

	rows := make([]Session, 0, len(records))
	for _, record := range records {
		rows = append(rows, historySession(record))
	}
	return rows, nil
}

func (a *App) GetAnalytics() (AnalyticsSummary, error) {
	records, err := loadHistoryRecords(a.GetConfigPath())
	if err != nil {
		return AnalyticsSummary{}, err
	}
	summary := analyticsSummary(records)
	a.mu.RLock()
	summary.HistoryError = a.lastHistoryErr
	a.mu.RUnlock()
	return summary, nil
}

// OpenFile opens the file at the given path with the OS default handler.
func (a *App) OpenFile(path string) error {
	path = filepath.Clean(path)
	switch goruntime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func (a *App) executeJob(jobID string, opts RunJobOptions, runCtx context.Context, cancel context.CancelFunc) {
	a.mu.RLock()
	job, ok := a.jobs[jobID]
	if !ok {
		a.mu.RUnlock()
		return
	}
	planned := cloneRenamePlan(job.planned)
	dir := job.dir
	configPath := job.configPath
	previewTokens := job.previewTokens
	a.mu.RUnlock()

	reporter := jobReporter{app: a, jobID: jobID}
	service := appsvc.NewService()
	hotRename := opts.HotRename
	defer cancel()
	defer a.clearJobCancel(jobID)

	run, err := service.PrepareRun(appsvc.RunOptions{
		Context:     runCtx,
		Dir:         dir,
		ConfigPath:  configPath,
		AutoApprove: opts.AutoApprove,
		HotRename:   &hotRename,
		DryRun:      opts.DryRun,
		Log:         opts.LogSession,
		Organize:    opts.Organize,
	}, reporter, desktopApprover{})
	if err != nil {
		a.finishJobFailure(jobID, fmt.Errorf("prepare run: %w", err), nil, DesktopConfig{}, opts, "")
		return
	}

	run.Query.Plan = planned
	config := desktopConfigFromCore(run.Config)
	results, runErr := service.ApplyPlan(run)

	logPath := ""
	if run.Query.Logger != nil {
		logPath = run.Query.Logger.GetLogFile()
	}
	closeErr := run.Close()
	if closeErr != nil && runErr == nil {
		runErr = closeErr
	}

	if runErr != nil {
		if errors.Is(runErr, context.Canceled) {
			a.finishJobCanceled(jobID, results, config, opts, logPath)
			return
		}
		a.finishJobFailure(jobID, runErr, results, config, opts, logPath)
		return
	}

	a.finishJobSuccess(jobID, results, config, opts, logPath, dir, previewTokens)
}

func (a *App) finishJobSuccess(jobID string, results []content.ProcessResult, config DesktopConfig, opts RunJobOptions, logPath string, baseDir string, previewTokens int) {
	summary := planSummary(len(results), results)
	analyticsSession, analyticsErr := latestAnalyticsSession(baseDir)
	record := desktopHistoryRecord{
		Date:      time.Now().UTC(),
		Directory: baseDir,
		Provider:  config.AI.Provider,
		Model:     config.AI.Model,
		DryRun:    opts.DryRun,
		Status:    "Complete",
		Planned:   summary.Planned,
		Renamed:   summary.Renamed,
		Skipped:   summary.Skipped,
		Errors:    summary.Errors,
		LogPath:   logPath,
	}
	if analyticsErr == nil {
		for _, model := range analyticsSession.Models {
			record.Tokens += model.TotalTokens
		}
	}
	if record.Tokens == 0 {
		record.Tokens = previewTokens
	}

	if err := appendHistoryRecord(a.GetConfigPath(), record); err != nil {
		a.mu.Lock()
		a.lastHistoryErr = err.Error()
		a.mu.Unlock()
		a.logErrorf("append desktop history: %v", err)
	}

	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if ok {
		job.cancel = nil
		job.plan = updatePlanFromResults(job.plan, results)
		job.status.State = "complete"
		job.status.Done = summary.Renamed + summary.Errors
		job.status.Total = len(job.plan)
		job.status.CurrentFile = ""
		job.status.Message = "job complete"
		job.status.Summary = summary
	}
	status := JobStatus{}
	if ok {
		status = job.status
	}
	a.mu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "job:complete", map[string]any{
			"jobID":   jobID,
			"summary": status.Summary,
		})
	}
}

func (a *App) finishJobFailure(jobID string, runErr error, results []content.ProcessResult, config DesktopConfig, opts RunJobOptions, logPath string) {
	if config.AI.Provider == "" {
		if loadedConfig, err := a.GetConfig(); err == nil {
			config = loadedConfig
		}
	}

	summary := planSummary(len(results), results)
	record := desktopHistoryRecord{
		Date:      time.Now().UTC(),
		Directory: "",
		Provider:  config.AI.Provider,
		Model:     config.AI.Model,
		DryRun:    opts.DryRun,
		Status:    "Failed",
		Planned:   summary.Planned,
		Renamed:   summary.Renamed,
		Skipped:   summary.Skipped,
		Errors:    summary.Errors + 1,
		LogPath:   logPath,
	}

	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if ok {
		job.cancel = nil
		record.Directory = job.dir
		job.plan = updatePlanFromResults(job.plan, results)
		job.status.State = "failed"
		job.status.Done = summary.Renamed + summary.Errors
		job.status.Total = len(job.plan)
		job.status.CurrentFile = ""
		job.status.Message = runErr.Error()
		job.status.Summary = JobSummary{
			Planned: summary.Planned,
			Renamed: summary.Renamed,
			Skipped: summary.Skipped,
			Errors:  summary.Errors + 1,
		}
	}
	a.mu.Unlock()

	if record.Directory != "" {
		if err := appendHistoryRecord(a.GetConfigPath(), record); err != nil {
			a.mu.Lock()
			a.lastHistoryErr = err.Error()
			a.mu.Unlock()
			a.logErrorf("append failed desktop history: %v", err)
		}
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "job:error", map[string]any{
			"jobID":   jobID,
			"message": runErr.Error(),
		})
	}
}

func (a *App) finishJobCanceled(jobID string, results []content.ProcessResult, config DesktopConfig, opts RunJobOptions, logPath string) {
	if config.AI.Provider == "" {
		if loadedConfig, err := a.GetConfig(); err == nil {
			config = loadedConfig
		}
	}

	summary := planSummary(len(results), results)
	summary.Skipped = maxInt(summary.Planned-summary.Renamed-summary.Errors, 0)
	record := desktopHistoryRecord{
		Date:      time.Now().UTC(),
		Directory: "",
		Provider:  config.AI.Provider,
		Model:     config.AI.Model,
		DryRun:    opts.DryRun,
		Status:    "Canceled",
		Planned:   summary.Planned,
		Renamed:   summary.Renamed,
		Skipped:   summary.Skipped,
		Errors:    summary.Errors,
		LogPath:   logPath,
	}

	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if ok {
		job.cancel = nil
		record.Directory = job.dir
		job.plan = updatePlanFromResults(job.plan, results)
		job.status.State = "canceled"
		job.status.Done = summary.Renamed + summary.Errors
		job.status.Total = len(job.plan)
		job.status.CurrentFile = ""
		job.status.Message = "job canceled"
		job.status.Summary = summary
	}
	a.mu.Unlock()

	if record.Directory != "" {
		if err := appendHistoryRecord(a.GetConfigPath(), record); err != nil {
			a.mu.Lock()
			a.lastHistoryErr = err.Error()
			a.mu.Unlock()
			a.logErrorf("append canceled desktop history: %v", err)
		}
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "job:canceled", map[string]any{
			"jobID":   jobID,
			"summary": summary,
		})
	}
}

func updatePlanFromResults(plan []RenameEntry, results []content.ProcessResult) []RenameEntry {
	updated := make([]RenameEntry, len(plan))
	copy(updated, plan)

	for index := range updated {
		if index >= len(results) {
			continue
		}
		result := results[index]
		if result.NewPath != "" {
			updated[index].NewName = filepath.Base(result.NewPath)
		}
		if result.Success {
			updated[index].Status = "Done"
			updated[index].Reason = ""
			continue
		}
		if result.Error != nil {
			updated[index].Status = "Error"
			updated[index].Reason = result.Error.Error()
		}
	}

	return updated
}

func (a *App) updateJobMessage(jobID, message string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	job, ok := a.jobs[jobID]
	if !ok {
		return
	}
	if message != "" {
		job.status.Message = message
	}
}

func (a *App) updateJobProgress(jobID string, done, total int, currentFile string) {
	a.mu.Lock()
	job, ok := a.jobs[jobID]
	if ok {
		job.status.State = "running"
		job.status.Done = done
		job.status.Total = total
		job.status.CurrentFile = currentFile
		job.status.Message = "processing files"
	}
	a.mu.Unlock()

	if ok && a.ctx != nil {
		runtime.EventsEmit(a.ctx, "job:progress", map[string]any{
			"jobID":       jobID,
			"done":        done,
			"total":       total,
			"currentFile": currentFile,
		})
	}
}

func (a *App) logErrorf(format string, args ...any) {
	if a.ctx != nil {
		runtime.LogErrorf(a.ctx, format, args...)
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (a *App) clearJobCancel(jobID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if job, ok := a.jobs[jobID]; ok {
		job.cancel = nil
	}
}
