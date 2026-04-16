package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	ai "nomnom/internal/ai"
	content "nomnom/internal/content"
	files "nomnom/internal/files"
	"nomnom/internal/utils"
)

type RunOptions struct {
	Dir         string
	ConfigPath  string
	Prompt      string
	OutputDir   string
	AutoApprove bool
	MoveFiles   *bool
	DryRun      bool
	Log         bool
	Organize    bool
}

type PreparedRun struct {
	Config    utils.Config
	Query     *content.Query
	OutputDir string

	closed bool
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) LoadConfig(configPath string) (utils.Config, error) {
	return utils.LoadConfig(configPath, "")
}

func (Service) PrepareRun(opts RunOptions, reporter utils.Reporter, approver utils.Approver) (*PreparedRun, error) {
	config, err := utils.LoadConfig(opts.ConfigPath, "")
	if err != nil {
		return nil, err
	}

	resolvedPrompt, err := content.ResolvePrompt(opts.Prompt, config)
	if err != nil {
		return nil, fmt.Errorf("resolve prompt: %w", err)
	}

	scan, err := content.ScanDirectory(opts.Dir, config, reporter)
	if err != nil {
		return nil, fmt.Errorf("scan directory: %w", err)
	}

	var logger *utils.Logger
	if !opts.DryRun {
		logger, err = utils.NewLogger(opts.Log, scan.RootDir)
		if err != nil {
			_ = scan.Cleanup()
			return nil, fmt.Errorf("create logger: %w", err)
		}
	}

	analytics := utils.NewAnalyticsStore(scan.RootDir, opts.DryRun)
	analytics.RecordScan(len(scan.Files))

	moveFiles := config.FileHandling.MoveFiles
	if opts.MoveFiles != nil {
		moveFiles = *opts.MoveFiles
	}

	query := content.NewQuery(content.QueryParams{
		Prompt:      resolvedPrompt,
		Dir:         scan.RootDir,
		ConfigPath:  opts.ConfigPath,
		AutoApprove: opts.AutoApprove,
		MoveFiles:   moveFiles,
		DryRun:      opts.DryRun,
		Log:         opts.Log,
		Logger:      logger,
		Organize:    opts.Organize,
		Reporter:    reporter,
		Approver:    approver,
		Analytics:   analytics,
		Scan:        scan,
	})

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = config.Output
	}
	if outputDir == "" {
		outputDir = filepath.Join(scan.RootDir, "nomnom", "renamed")
	}

	return &PreparedRun{
		Config:    config,
		Query:     query,
		OutputDir: outputDir,
	}, nil
}

func (Service) GeneratePlan(run *PreparedRun) error {
	if run == nil || run.Query == nil {
		return fmt.Errorf("prepared run is nil")
	}

	result, err := ai.HandleAI(run.Config, *run.Query)
	if err != nil {
		return err
	}

	run.Query.Plan = result.Plan
	if err := persistRenamePlan(run.Query.Scan.RootDir, run.Query.Prompt, run.Query.Plan); err != nil && run.Query.Reporter != nil {
		run.Query.Reporter.Warnf("Failed to persist rename plan cache: %v", err)
	}
	if run.Query.Analytics != nil {
		run.Query.Analytics.RecordRenamePlan(len(result.Plan))
	}

	return nil
}

func (Service) ApplyPlan(run *PreparedRun) ([]content.ProcessResult, error) {
	if run == nil || run.Query == nil {
		return nil, fmt.Errorf("prepared run is nil")
	}

	processor := content.NewSafeProcessor(run.Query, run.OutputDir)
	return processor.Process()
}

func (Service) LoadAnalytics(baseDir string) (utils.AnalyticsSummary, []utils.SessionAnalytics, error) {
	summary, err := utils.LoadAnalyticsSummary(baseDir)
	if err != nil {
		return utils.AnalyticsSummary{}, nil, err
	}

	sessionPaths, err := utils.ListAnalyticsSessions(baseDir)
	if err != nil {
		return utils.AnalyticsSummary{}, nil, err
	}

	sessions := make([]utils.SessionAnalytics, 0, len(sessionPaths))
	for _, path := range sessionPaths {
		session, err := utils.LoadAnalyticsSession(path)
		if err != nil {
			return utils.AnalyticsSummary{}, nil, err
		}
		sessions = append(sessions, session)
	}

	slices.SortFunc(sessions, func(a, b utils.SessionAnalytics) int {
		if a.StartedAt.Equal(b.StartedAt) {
			return 0
		}
		if a.StartedAt.After(b.StartedAt) {
			return -1
		}
		return 1
	})

	return summary, sessions, nil
}

func (run *PreparedRun) Close() error {
	if run == nil || run.closed {
		return nil
	}
	run.closed = true

	var closeErr error
	if run.Query != nil {
		closeErr = errors.Join(closeErr, run.Query.Scan.Cleanup())
		closeErr = errors.Join(closeErr, files.CleanupPreviewTempDir())
		if run.Query.Logger != nil {
			closeErr = errors.Join(closeErr, run.Query.Logger.Close())
		}
		if run.Query.Analytics != nil {
			closeErr = errors.Join(closeErr, run.Query.Analytics.Close())
		}
	}

	return closeErr
}

type cachedRenamePlan struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	Prompt      string                    `json:"prompt"`
	Entries     []content.RenamePlanEntry `json:"entries"`
}

func persistRenamePlan(rootDir, prompt string, plan []content.RenamePlanEntry) error {
	cacheDir := filepath.Join(rootDir, ".nomnom", "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("create plan cache directory: %w", err)
	}

	payload := cachedRenamePlan{
		GeneratedAt: time.Now().UTC(),
		Prompt:      prompt,
		Entries:     plan,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan cache: %w", err)
	}
	data = append(data, '\n')

	filename := fmt.Sprintf("rename_plan_%d.json", payload.GeneratedAt.Unix())
	cachePath := filepath.Join(cacheDir, filename)
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		return fmt.Errorf("write plan cache: %w", err)
	}

	return nil
}
