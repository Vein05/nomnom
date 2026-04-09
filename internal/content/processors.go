package content

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	utils "nomnom/internal/utils"

	"slices"
)

const defaultPrompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."

type QueryParams struct {
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
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
	query  *Query
	output string
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
	return &SafeProcessor{query: query, output: output}
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
		if err := os.MkdirAll(p.output, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
		reporter.Infof("Created output directory: %s", p.output)
	}

	results := make([]ProcessResult, 0, len(p.query.Plan))
	for _, entry := range p.query.Plan {
		result, err := p.processEntry(entry)
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

func (p *SafeProcessor) processEntry(entry RenamePlanEntry) (ProcessResult, error) {
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
			decision, approveErr := p.promptForRenameApproval(entry.File.OriginalName, filepath.Base(targetPath))
			if approveErr != nil {
				return result, approveErr
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

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		if err := copyFile(entry.File.SourcePath, targetPath); err != nil {
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
