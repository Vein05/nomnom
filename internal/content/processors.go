package content

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	utils "nomnom/internal/utils"

	"slices"
)

const defaultPrompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."

type QueryParams struct {
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	MoveFiles   bool
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
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	MoveFiles   bool
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
	query       *Query
	output      string
	createdDirs map[string]struct{}
}

type FileTypeCategory struct {
	Name       string
	Extensions []string
}

type Prompts struct {
	Name     string
	Path     string
	TestPath string
}

var defaultCategories = []FileTypeCategory{
	{Name: "Images", Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}},
	{Name: "Documents", Extensions: []string{".pdf", ".doc", ".docx", ".txt", ".md", ".rtf"}},
	{Name: "Audios", Extensions: []string{".mp3", ".wav", ".flac", ".m4a", ".aac"}},
	{Name: "Videos", Extensions: []string{".mp4", ".mov", ".avi", ".mkv", ".wmv"}},
	{Name: "Others", Extensions: []string{}},
}

var NomNomPrompts = []Prompts{
	{Name: "research", Path: "data/prompts/research.txt", TestPath: "../../data/prompts/research.txt"},
	{Name: "images", Path: "data/prompts/images.txt", TestPath: "../../data/prompts/images.txt"},
}

func NewQuery(params QueryParams) *Query {
	reporter := params.Reporter
	if reporter == nil {
		reporter = utils.NopReporter{}
	}

	dir := params.Dir
	if dir == "" {
		dir = params.Scan.RootDir
	}

	return &Query{
		Prompt:      params.Prompt,
		Dir:         dir,
		ConfigPath:  params.ConfigPath,
		AutoApprove: params.AutoApprove,
		MoveFiles:   params.MoveFiles,
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

	results := make([]ProcessResult, 0, len(p.query.Plan))
	for _, entry := range p.query.Plan {
		result, err := p.processEntry(entry, approvals)
		if err != nil {
			reporter.Errorf("Failed to process %s: %v", entry.File.OriginalName, err)
		}
		results = append(results, result)
		if p.query.Analytics != nil && !p.query.DryRun {
			p.query.Analytics.RecordRenameResult(result.Success)
		}
	}

	reporter.Infof("Completed file processing phase")
	return results, nil
}

func (p *SafeProcessor) processEntry(entry RenamePlanEntry, approvals map[string]utils.ApprovalDecision) (ProcessResult, error) {
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
					return result, approveErr
				}
			}
			if decision == utils.ApprovalNo {
				result.Success = false
				result.Error = fmt.Errorf("rename not approved")
				return result, result.Error
			}
			if decision == utils.ApprovalAll {
				p.query.AutoApprove = true
			}
		}

		if err := p.ensureDir(filepath.Dir(targetPath)); err != nil {
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
	if p.query.MoveFiles {
		return moveOrCopyFile(src, dst)
	}

	return copyFile(src, dst)
}

func moveOrCopyFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return fmt.Errorf("failed to move file: %w", err)
	}

	if err := copyFile(src, dst); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source after cross-device copy: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

func (p *SafeProcessor) reporter() utils.Reporter {
	if p.query != nil && p.query.Reporter != nil {
		return p.query.Reporter
	}
	return utils.NopReporter{}
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

	switch strings.ToLower(trimmedPrompt) {
	case "research":
		return readPromptFile(NomNomPrompts[0], defaultPrompt)
	case "images":
		return readPromptFile(NomNomPrompts[1], defaultPrompt)
	default:
		return trimmedPrompt, nil
	}
}

func readPromptFile(promptFile Prompts, fallback string) (string, error) {
	content, err := os.ReadFile(promptFile.Path)
	if err == nil {
		return string(content), nil
	}

	content, testErr := os.ReadFile(promptFile.TestPath)
	if testErr == nil {
		return string(content), nil
	}

	return fallback, fmt.Errorf("failed to read prompt %q from both paths: %w", promptFile.Name, err)
}
