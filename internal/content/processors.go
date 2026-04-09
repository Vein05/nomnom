package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	utils "nomnom/internal/utils"

	"slices"
)

const defaultPrompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."

// Query represents the query parameters for content processing with the following fields:
type Query struct {
	// Prompt holds the user-provided text prompt for content processing
	Prompt string
	// Dir specifies the root directory path to process
	Dir string
	// ConfigPath is the path to the configuration file
	ConfigPath string
	// AutoApprove when true skips confirmation prompts for file operations
	AutoApprove bool
	// DryRun when true simulates operations without making actual changes
	DryRun bool
	// Log enables logging of operations when true
	Log bool
	// Folders contains the hierarchical structure of directories and files to process
	Folders []FolderType
	// Logger provides logging functionality for operations
	Logger *utils.Logger
	// Organize when true enables structured organization of files by category
	Organize bool
	// Reporter handles user-facing progress output
	Reporter utils.Reporter
	// Approver handles interactive approvals
	Approver utils.Approver
}

// ProcessResult captures the outcome of file processing operations including:
type ProcessResult struct {
	// OriginalPath stores the initial relative path of the processed file
	OriginalPath string
	// NewPath contains the new relative path after processing
	NewPath string
	// FullOriginalPath stores the initial absolute path of the processed file
	FullOriginalPath string
	// FullNewPath contains the new absolute path after processing
	FullNewPath string
	// Success indicates whether the processing operation succeeded
	Success bool
	// Error holds any error that occurred during processing
	Error error
}

// SafeProcessor implements safe file processing operations with validation
type SafeProcessor struct {
	// query holds the processing parameters and configuration
	query *Query
	// output specifies the destination directory for processed files
	output string
}

// FileTypeCategory defines a group of file extensions belonging to a category
type FileTypeCategory struct {
	// Name is the category identifier (e.g., "Images", "Documents")
	Name string
	// Extensions lists the file extensions that belong to this category
	Extensions []string
}

var defaultCategories = []FileTypeCategory{
	{
		Name:       "Images",
		Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
	},
	{
		Name:       "Documents",
		Extensions: []string{".pdf", ".doc", ".docx", ".txt", ".md", ".rtf"},
	},
	{
		Name:       "Audios",
		Extensions: []string{".mp3", ".wav", ".flac", ".m4a", ".aac"},
	},
	{
		Name:       "Videos",
		Extensions: []string{".mp4", ".mov", ".avi", ".mkv", ".wmv"},
	},
	{
		Name:       "Others",
		Extensions: []string{}, // Catch-all for unmatched types
	},
}

// FolderType represents a directory structure containing files and subfolders
type Prompts struct {
	Name     string // Name of the prompt
	Path     string // Path to the prompt file in the production environment
	TestPath string // Path to the prompt file in the test environment
}

var NomNomPrompts []Prompts = []Prompts{
	{
		Name:     "research",
		Path:     "data/prompts/research.txt",
		TestPath: "../../data/prompts/research.txt",
	},
	{
		Name:     "images",
		Path:     "data/prompts/images.txt",
		TestPath: "../../data/prompts/images.txt",
	},
}

// NewQuery creates a new Query object with the given parameters.
func NewQuery(prompt string, dir string, configPath string, config utils.Config, autoApprove bool, dryRun bool, log bool, organize bool, reporter utils.Reporter, approver utils.Approver) (*Query, error) {
	prompt, err := resolvePrompt(prompt, config)
	if err != nil {
		return nil, fmt.Errorf("error resolving prompt: %w", err)
	}
	if reporter == nil {
		reporter = utils.NopReporter{}
	}
	folders, err := ProcessDirectory(dir, config, reporter)
	if err != nil {
		return nil, fmt.Errorf("error processing directory: %w", err)
	}

	var logger *utils.Logger
	if !dryRun {
		logger, err = utils.NewLogger(log, dir)
		if err != nil {
			return nil, fmt.Errorf("error creating logger: %w", err)
		}
	}

	return &Query{
		Dir:         dir,
		ConfigPath:  configPath,
		AutoApprove: autoApprove,
		DryRun:      dryRun,
		Log:         log,
		Prompt:      prompt,
		Folders:     folders.Folders,
		Logger:      logger,
		Organize:    organize,
		Reporter:    reporter,
		Approver:    approver,
	}, nil
}

// NewSafeProcessor creates a new SafeProcessor instance
func NewSafeProcessor(query *Query, output string) *SafeProcessor {
	return &SafeProcessor{
		query:  query,
		output: output,
	}
}

// Process handles the safe mode processing workflow
func (p *SafeProcessor) Process() ([]ProcessResult, error) {
	reporter := p.reporter()
	reporter.Infof("Starting safe mode processing")
	results := make([]ProcessResult, 0)

	if p.query.DryRun {
		reporter.Infof("Dry run: would create output directory")
	} else {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(p.output, 0755); err != nil {
			reporter.Errorf("Failed to create output directory: %v", err)
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
		reporter.Infof("Created output directory: %s", p.output)
	}

	// First phase: Copy all files with original structure
	if !p.query.DryRun {
		if !p.query.Organize {
			reporter.Infof("Starting file copy phase in original structure")
			if err := p.copyOriginalStructure(); err != nil {
				reporter.Errorf("Failed to copy original structure: %v", err)
				return nil, fmt.Errorf("failed to copy original structure: %w", err)
			}
		} else {
			reporter.Infof("Starting file copy phase in organized structure")
			if err := p.copyOrganizedStructure(); err != nil {
				reporter.Errorf("Failed to copy organized structure: %v", err)
				return nil, fmt.Errorf("failed to copy organized structure: %w", err)
			}
		}
		reporter.Infof("Completed file copy phase")
	}

	// Second phase: Process and rename files
	reporter.Infof("Starting file processing phase")

	var processFolder func(folder FolderType, relativePath string) error
	processFolder = func(folder FolderType, relativePath string) error {
		reporter.Infof("Processing folder: %s, files: %d", folder.Name, len(folder.FileList))
		counter := 0
		for _, file := range folder.FileList {
			currentPath := file.Path
			newPath := filepath.Join(filepath.Dir(currentPath), file.NewName)

			fullPath, err := filepath.Abs(file.UNCHANGEDPATH)
			if err != nil {
				reporter.Errorf("Error getting absolute path: %v", err)
			}

			result := ProcessResult{
				OriginalPath:     file.UNCHANGEDPATH,
				FullOriginalPath: fullPath,
				Success:          true,
			}

			// Skip if no new name was generated
			if file.NewName == "" {
				result.NewPath = currentPath
				result.FullNewPath, _ = filepath.Abs(currentPath)
				reporter.Warnf("No new name generated for: %s", file.Name)
				results = append(results, result)
				continue
			}

			fullNewPath, err := filepath.Abs(newPath)
			if err != nil {
				reporter.Errorf("Error getting absolute path: %v", err)
			}
			result.NewPath = newPath
			result.FullNewPath = fullNewPath

			if _, err := os.Stat(newPath); err == nil {
				newPath = utils.GenerateUniqueFilename(newPath)
				result.NewPath = newPath
				result.FullNewPath, _ = filepath.Abs(newPath)
				reporter.Warnf("Duplicate file detected, renaming to: %s", newPath)
			}

			// In non-dry-run mode, rename the copied file
			if !p.query.DryRun {
				if !p.query.AutoApprove && counter == 0 {
					reporter.Warnf("Auto approve is disabled")
				}

				if !p.query.AutoApprove {
					response, err := p.promptForRenameApproval(file.Name, file.NewName)
					if err != nil {
						reporter.Errorf("Error running prompt: %v", err)
					}
					if response == utils.ApprovalNo {
						reporter.Warnf("Skipping rename for: %s", file.Name)
						continue
					}
					if response == utils.ApprovalAll {
						p.query.AutoApprove = true
						reporter.Infof("Auto approving all renames")
					}
				}

				if err := os.Rename(currentPath, newPath); err != nil {
					reporter.Errorf("Failed to rename file: %v", err)
					result.Success = false
					result.Error = fmt.Errorf("failed to rename file: %w", err)
				} else {
					reporter.Infof("Successfully renamed file")
				}
			}

			// Log the operation if logging is enabled
			if p.query.Logger != nil && !p.query.DryRun {
				absOrigPath, _ := filepath.Abs(file.UNCHANGEDPATH)
				absNewPath, _ := filepath.Abs(newPath)
				p.query.Logger.LogOperation(absOrigPath, absNewPath, result.Success, result.Error)
			}

			results = append(results, result)
			counter++
		}
		// Process subfolders recursively
		for _, subFolder := range folder.SubFolders {
			newRelPath := filepath.Join(relativePath, subFolder.Name)
			if err := processFolder(subFolder, newRelPath); err != nil {
				return err
			}
		}
		return nil
	}
	// Start processing from root folders
	for _, folder := range p.query.Folders {
		if err := processFolder(folder, folder.Name); err != nil {
			reporter.Errorf("Failed to process folder: %s, error: %v", folder.Name, err)
			return nil, fmt.Errorf("failed to process folder %s: %w", folder.Name, err)
		}
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			reporter.Errorf("Failed to close logger: %v", err)
		} else {
			reporter.Infof("Successfully closed logger")
		}
	}

	reporter.Infof("Completed file processing phase")
	return results, nil
}

// promptForRenameApproval handles the user prompt for rename approval
func (p *SafeProcessor) promptForRenameApproval(oldName, newName string) (utils.ApprovalDecision, error) {
	if p.query.Approver == nil {
		return utils.ApprovalNo, fmt.Errorf("no approver configured")
	}
	return p.query.Approver.Approve("rename", oldName, newName)
}

func getCategoryForFile(fileName string) string {
	ext := filepath.Ext(fileName)
	for _, category := range defaultCategories {
		if slices.Contains(category.Extensions, ext) {
			return category.Name
		}
	}
	return "Others"
}

func (p *SafeProcessor) copyOrganizedStructure() error {
	// Create category folders only at root level
	for _, category := range defaultCategories {
		categoryPath := filepath.Join(p.output, category.Name)
		if err := os.MkdirAll(categoryPath, 0755); err != nil {
			return fmt.Errorf("failed to create category folder %s: %w", categoryPath, err)
		}
	}

	// Process all folders recursively
	var processOrganizedFolder func(folder FolderType, relativePath string) error
	processOrganizedFolder = func(folder FolderType, relativePath string) error {
		for i := range folder.FileList {
			file := &folder.FileList[i]
			category := getCategoryForFile(file.Name)

			relativeFilePath := filepath.Join(relativePath, file.Name)
			dstPath := filepath.Join(p.output, category, relativeFilePath)

			// Ensure subdirectory exists within category
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return fmt.Errorf("failed to create subdirectory: %w", err)
			}

			if err := copyFile(file.Path, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}

			// update the file path in the original folder
			file.Path = dstPath
		}

		// Process subfolders
		for _, subFolder := range folder.SubFolders {
			newRelPath := filepath.Join(relativePath, subFolder.Name)
			if err := processOrganizedFolder(subFolder, newRelPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Start processing from root folders
	for _, folder := range p.query.Folders {
		if err := processOrganizedFolder(folder, ""); err != nil {
			return err
		}
	}
	return nil
}

func (p *SafeProcessor) copyOriginalStructure() error {
	// Process all folders recursively
	var processOriginalFolder func(folder FolderType, relativePath string) error
	processOriginalFolder = func(folder FolderType, relativePath string) error {
		// Create folder in output directory
		outFolder := filepath.Join(p.output, relativePath)
		if err := os.MkdirAll(outFolder, 0755); err != nil {
			p.reporter().Errorf("Failed to create output folder: %s, error: %v", outFolder, err)
			return fmt.Errorf("failed to create output folder %s: %w", outFolder, err)
		}
		p.reporter().Infof("Created output folder: %s", outFolder)

		// Copy each file with original name
		for i := range folder.FileList {
			file := &folder.FileList[i]
			dstPath := filepath.Join(outFolder, file.Name)
			if err := copyFile(file.Path, dstPath); err != nil {
				p.reporter().Errorf("Failed to copy file: %s to %s, error: %v", file.Path, dstPath, err)
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
			file.Path = dstPath
		}

		// Process subfolders recursively
		for _, subFolder := range folder.SubFolders {
			newRelPath := filepath.Join(relativePath, subFolder.Name)
			if err := processOriginalFolder(subFolder, newRelPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Start processing from root folders
	for _, folder := range p.query.Folders {
		if err := processOriginalFolder(folder, ""); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, input, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}
	return nil
}

func (p *SafeProcessor) reporter() utils.Reporter {
	if p.query != nil && p.query.Reporter != nil {
		return p.query.Reporter
	}
	return utils.NopReporter{}
}

func resolvePrompt(prompt string, config utils.Config) (string, error) {
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
