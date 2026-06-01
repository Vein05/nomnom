package content

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	prompts "nomnom/data/prompts"
	utils "nomnom/internal/utils"

	"slices"
)

const defaultPrompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."

func DefaultPrompt() string {
	return defaultPrompt
}

type QueryParams struct {
	Context     context.Context
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	HotRename   bool
	DryRun      bool
	Log         bool
	Logger      *utils.Logger
	Organize    bool
	Reporter    utils.Reporter
	Approver    utils.Approver
	Analytics   *utils.AnalyticsStore
	Scan        ScanResult
}

type Query struct {
	Context     context.Context
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	HotRename   bool
	DryRun      bool
	Log         bool
	Logger      *utils.Logger
	Organize    bool
	Reporter    utils.Reporter
	Approver    utils.Approver
	Analytics   *utils.AnalyticsStore
	Scan        ScanResult
	Plan        []RenamePlanEntry
}

type RenamePlanEntry struct {
	File          ScannedFile
	SuggestedName string
}

type ProcessResult struct {
	OriginalPath     string
	NewPath          string
	FullOriginalPath string
	FullNewPath      string
	Success          bool
	Error            error
}

type SafeProcessor struct {
	query         *Query
	output        string
	createdDirs   map[string]struct{}
	createdDirsMu sync.Mutex
}

type FileTypeCategory struct {
	Name       string
	Extensions []string
}

var defaultCategories = []FileTypeCategory{
	{Name: "Images", Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}},
	{Name: "Documents", Extensions: []string{".pdf", ".doc", ".docx", ".txt", ".md", ".rtf"}},
	{Name: "Audios", Extensions: []string{".mp3", ".wav", ".flac", ".m4a", ".aac"}},
	{Name: "Videos", Extensions: []string{".mp4", ".mov", ".avi", ".mkv", ".wmv"}},
	{Name: "Others", Extensions: []string{}},
}

// embeddedPrompts maps prompt names to their compile-time embedded content.
var embeddedPrompts = map[string]string{
	"research": prompts.ResearchPrompt,
	"images":   prompts.ImagesPrompt,
}

func NewQuery(params QueryParams) *Query {
	reporter := params.Reporter
	if reporter == nil {
		reporter = utils.NopReporter{}
	}
	ctx := params.Context
	if ctx == nil {
		ctx = context.Background()
	}

	dir := params.Dir
	if dir == "" {
		dir = params.Scan.RootDir
	}

	return &Query{
		Context:     ctx,
		Prompt:      params.Prompt,
		Dir:         dir,
		ConfigPath:  params.ConfigPath,
		AutoApprove: params.AutoApprove,
		HotRename:   params.HotRename,
		DryRun:      params.DryRun,
		Log:         params.Log,
		Logger:      params.Logger,
		Organize:    params.Organize,
		Reporter:    reporter,
		Approver:    params.Approver,
		Analytics:   params.Analytics,
		Scan:        params.Scan,
		Plan:        make([]RenamePlanEntry, 0, len(params.Scan.Files)),
	}
}

func NewSafeProcessor(query *Query, output string) *SafeProcessor {
	return &SafeProcessor{query: query, output: output, createdDirs: make(map[string]struct{})}
}

func (p *SafeProcessor) Process() ([]ProcessResult, error) {
	reporter := p.reporter()
	if err := p.contextErr(); err != nil {
		return nil, err
	}
	reporter.Infof("Starting safe mode processing")

	if len(p.query.Plan) == 0 {
		return []ProcessResult{}, nil
	}

	if p.query.DryRun {
		reporter.Infof("Dry run: would create output directory")
	} else {
		if err := p.ensureDir(p.output); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
		reporter.Infof("Created output directory: %s", p.output)
	}

	approvals := make(map[string]utils.ApprovalDecision)
	if !p.query.DryRun && !p.query.AutoApprove {
		var err error
		approvals, err = p.collectApprovals()
		if err != nil {
			return nil, err
		}
	}

	results := make([]ProcessResult, len(p.query.Plan))
	workers := runtime.NumCPU()
	if workers > len(p.query.Plan) {
		workers = len(p.query.Plan)
	}

	jobs := make(chan int, len(p.query.Plan))
	var wg sync.WaitGroup
	var completed sync.Mutex
	done := 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				if err := p.contextErr(); err != nil {
					return
				}
				entry := p.query.Plan[idx]
				result, err := p.processEntry(entry, approvals)
				if err != nil {
					reporter.Errorf("Failed to process %s: %v", entry.File.OriginalName, err)
				}
				results[idx] = result
				if p.query.Analytics != nil && !p.query.DryRun {
					p.query.Analytics.RecordRenameResult(result.Success)
				}
				completed.Lock()
				done++
				currentDone := done
				completed.Unlock()
				if progress, ok := reporter.(utils.ProgressReporter); ok {
					progress.ReportProgress(currentDone, len(p.query.Plan), entry.File.RelativePath)
				}
			}
		}()
	}

	for i := range p.query.Plan {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if err := p.contextErr(); err != nil {
		return results, err
	}

	reporter.Infof("Completed file processing phase")
	return results, nil
}

func (p *SafeProcessor) processEntry(entry RenamePlanEntry, approvals map[string]utils.ApprovalDecision) (ProcessResult, error) {
	if err := p.contextErr(); err != nil {
		return ProcessResult{
			OriginalPath: entry.File.SourcePath,
			Success:      false,
			Error:        err,
		}, err
	}

	sourcePath, err := filepath.Abs(entry.File.SourcePath)
	if err != nil {
		return ProcessResult{OriginalPath: entry.File.SourcePath, Success: false, Error: err}, err
	}

	if entry.SuggestedName == "" {
		err := fmt.Errorf("no suggested name generated")
		return ProcessResult{
			OriginalPath:     entry.File.SourcePath,
			NewPath:          entry.File.SourcePath,
			FullOriginalPath: sourcePath,
			FullNewPath:      sourcePath,
			Success:          false,
			Error:            err,
		}, err
	}

	targetPath := p.destinationPath(entry)
	if _, err := os.Stat(targetPath); err == nil {
		targetPath = utils.GenerateUniqueFilename(targetPath)
	}

	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return ProcessResult{OriginalPath: entry.File.SourcePath, Success: false, Error: err}, err
	}

	result := ProcessResult{
		OriginalPath:     entry.File.SourcePath,
		NewPath:          targetPath,
		FullOriginalPath: sourcePath,
		FullNewPath:      targetAbs,
		Success:          true,
	}

	if !p.query.DryRun {
		if !p.query.AutoApprove {
			decision, ok := approvals[entry.File.SourcePath]
			if !ok {
				var approveErr error
				decision, approveErr = p.promptForRenameApproval(entry.File.OriginalName, filepath.Base(targetPath))
				if approveErr != nil {
					result.Success = false
					result.Error = approveErr
					return result, approveErr
				}
			}
			if decision == utils.ApprovalNo {
				result.Success = false
				result.Error = fmt.Errorf("rename not approved")
				return result, result.Error
			}
		}

		if err := p.ensureDir(filepath.Dir(targetPath)); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		if err := p.contextErr(); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		if err := p.writeFile(entry.File.SourcePath, targetPath); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		if p.query.Logger != nil {
			p.query.Logger.LogOperation(sourcePath, targetAbs, result.Success, result.Error)
		}
	}

	return result, nil
}

func (p *SafeProcessor) collectApprovals() (map[string]utils.ApprovalDecision, error) {
	approvals := make(map[string]utils.ApprovalDecision, len(p.query.Plan))

	for _, entry := range p.query.Plan {
		if err := p.contextErr(); err != nil {
			return nil, err
		}
		if p.query.AutoApprove {
			break
		}
		if entry.SuggestedName == "" {
			continue
		}

		decision, err := p.promptForRenameApproval(entry.File.OriginalName, entry.SuggestedName)
		if err != nil {
			return nil, err
		}
		approvals[entry.File.SourcePath] = decision

		if decision == utils.ApprovalAll {
			p.query.AutoApprove = true
		}
	}

	return approvals, nil
}

func (p *SafeProcessor) destinationPath(entry RenamePlanEntry) string {
	if p.query.HotRename {
		return filepath.Join(filepath.Dir(entry.File.SourcePath), entry.SuggestedName)
	}

	relativeDir := filepath.Dir(entry.File.RelativePath)
	if relativeDir == "." {
		relativeDir = ""
	}

	if p.query.Organize {
		return filepath.Join(p.output, entry.File.Category, relativeDir, entry.SuggestedName)
	}
	return filepath.Join(p.output, relativeDir, entry.SuggestedName)
}

func (p *SafeProcessor) promptForRenameApproval(oldName, newName string) (utils.ApprovalDecision, error) {
	if p.query.Approver == nil {
		return utils.ApprovalNo, fmt.Errorf("no approver configured")
	}
	return p.query.Approver.Approve("rename", oldName, newName)
}

func (p *SafeProcessor) ensureDir(dir string) error {
	if err := p.contextErr(); err != nil {
		return err
	}

	p.createdDirsMu.Lock()
	defer p.createdDirsMu.Unlock()

	if _, ok := p.createdDirs[dir]; ok {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	p.createdDirs[dir] = struct{}{}
	return nil
}

func (p *SafeProcessor) writeFile(src, dst string) error {
	if p.query.HotRename {
		return moveOrCopyFile(p.context(), src, dst)
	}

	return copyFile(p.context(), src, dst)
}

func moveOrCopyFile(ctx context.Context, src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return fmt.Errorf("failed to move file: %w", err)
	}

	if err := copyFile(ctx, src, dst); err != nil {
		return err
	}
	if err := contextErr(ctx); err != nil {
		_ = os.Remove(dst)
		return err
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source after cross-device copy: %w", err)
	}

	return nil
}

func copyFile(ctx context.Context, src, dst string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	closeOutput := true
	defer func() {
		if closeOutput {
			_ = output.Close()
		}
	}()

	buffer := make([]byte, 1024*1024)
	for {
		if err := contextErr(ctx); err != nil {
			_ = output.Close()
			closeOutput = false
			_ = os.Remove(dst)
			return err
		}

		n, readErr := input.Read(buffer)
		if n > 0 {
			if _, err := output.Write(buffer[:n]); err != nil {
				_ = output.Close()
				closeOutput = false
				_ = os.Remove(dst)
				return fmt.Errorf("failed to copy file contents: %w", err)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			_ = output.Close()
			closeOutput = false
			_ = os.Remove(dst)
			return fmt.Errorf("failed to copy file contents: %w", readErr)
		}
	}

	if err := output.Close(); err != nil {
		closeOutput = false
		_ = os.Remove(dst)
		return fmt.Errorf("failed to close destination file: %w", err)
	}
	closeOutput = false

	if err := contextErr(ctx); err != nil {
		_ = os.Remove(dst)
		return err
	}

	return nil
}

func (p *SafeProcessor) reporter() utils.Reporter {
	if p.query != nil && p.query.Reporter != nil {
		return p.query.Reporter
	}
	return utils.NopReporter{}
}

func (p *SafeProcessor) context() context.Context {
	if p.query != nil && p.query.Context != nil {
		return p.query.Context
	}
	return context.Background()
}

func (p *SafeProcessor) contextErr() error {
	return contextErr(p.context())
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func categoryForFile(fileName string) string {
	ext := filepath.Ext(fileName)
	for _, category := range defaultCategories {
		if slices.Contains(category.Extensions, ext) {
			return category.Name
		}
	}
	return "Others"
}

func ResolvePrompt(prompt string, config utils.Config) (string, error) {
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		if strings.TrimSpace(config.AI.Prompt) != "" {
			return config.AI.Prompt, nil
		}
		return defaultPrompt, nil
	}

	key := strings.ToLower(trimmedPrompt)
	if content, ok := embeddedPrompts[key]; ok {
		return content, nil
	}

	return trimmedPrompt, nil
}
