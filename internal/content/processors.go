package nomnom

import (
	"fmt"
	utils "nomnom/internal/utils"
	"os"
	"path/filepath"
)

// Query represents the query parameters for content processing.
type Query struct {
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	DryRun      bool
	Log         bool
	Folders     []FolderType
	Logger      *utils.Logger
}

// ProcessResult represents the result of processing files
type ProcessResult struct {
	OriginalPath string
	NewPath      string
	Success      bool
	Error        error
}

// SafeProcessor handles file processing in safe mode
type SafeProcessor struct {
	query  *Query
	output string
}

// NewQuery creates a new Query object with the given parameters.
func NewQuery(prompt string, dir string, configPath string, autoApprove bool, dryRun bool, log bool) (*Query, error) {
	if prompt == "" {
		prompt = "What is the title of this document? Only respond with the title."
	}

	folders, err := ProcessDirectory(dir)
	if err != nil {
		return nil, fmt.Errorf("error processing directory: %w", err)
	}

	logger, err := utils.NewLogger(log, dir)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
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
	results := make([]ProcessResult, 0)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(p.output, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// First phase: Copy all files with original structure
	if !p.query.DryRun {
		if err := p.copyOriginalStructure(); err != nil {
			return nil, fmt.Errorf("failed to copy original structure: %w", err)
		}
	}

	// Second phase: Process and rename files
	for _, folder := range p.query.Folders {
		// Get the output folder path
		outFolder := filepath.Join(p.output, folder.Name)

		// Process each file
		for _, file := range folder.FileList {
			result := ProcessResult{
				OriginalPath: file.Path,
				Success:      true,
			}

			// Current copied file path (before rename)
			currentPath := filepath.Join(outFolder, file.Name)

			// Generate new path
			if file.NewName != "" {
				result.NewPath = filepath.Join(outFolder, file.NewName)

				// In non-dry-run mode, rename the copied file
				if !p.query.DryRun {
					if err := os.Rename(currentPath, result.NewPath); err != nil {
						result.Success = false
						result.Error = fmt.Errorf("failed to rename file: %w", err)
					}
				}

				// Log the operation if logging is enabled
				if p.query.Logger != nil {
					// Convert paths to absolute
					absOrigPath, err := filepath.Abs(file.Path)
					if err != nil {
						fmt.Printf("Warning: Could not get absolute path for %s: %v\n", file.Path, err)
						absOrigPath = file.Path
					}
					absNewPath, err := filepath.Abs(result.NewPath)
					if err != nil {
						fmt.Printf("Warning: Could not get absolute path for %s: %v\n", result.NewPath, err)
						absNewPath = result.NewPath
					}
					p.query.Logger.LogOperation(absOrigPath, absNewPath, result.Success, result.Error)
				}
			} else {
				result.NewPath = currentPath
			}

			results = append(results, result)
		}
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			fmt.Printf("Warning: Failed to close logger: %v\n", err)
		}
	}

	return results, nil
}

// copyOriginalStructure copies the entire directory structure and files to the output directory
func (p *SafeProcessor) copyOriginalStructure() error {
	for _, folder := range p.query.Folders {
		// Create corresponding folder in output directory
		outFolder := filepath.Join(p.output, folder.Name)
		if err := os.MkdirAll(outFolder, 0755); err != nil {
			return fmt.Errorf("failed to create output folder %s: %w", outFolder, err)
		}

		// Copy each file with original name
		for _, file := range folder.FileList {
			dstPath := filepath.Join(outFolder, file.Name)
			if err := copyFile(file.Path, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
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
