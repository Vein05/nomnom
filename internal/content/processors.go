package nomnom

import (
	"fmt"
	"log"
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
	log.Printf("[INFO] Starting safe mode processing")
	results := make([]ProcessResult, 0)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(p.output, 0755); err != nil {
		log.Printf("[ERROR] Failed to create output directory: %v", err)
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	log.Printf("[INFO] Created output directory: %s", p.output)

	// First phase: Copy all files with original structure
	if !p.query.DryRun {
		log.Printf("[INFO] Starting file copy phase")
		if err := p.copyOriginalStructure(); err != nil {
			log.Printf("[ERROR] Failed to copy original structure: %v", err)
			return nil, fmt.Errorf("failed to copy original structure: %w", err)
		}
		log.Printf("[INFO] Completed file copy phase")
	}

	// Second phase: Process and rename files
	log.Printf("[INFO] Starting file processing phase")
	for _, folder := range p.query.Folders {
		// Get the output folder path
		outFolder := filepath.Join(p.output, folder.Name)
		log.Printf("[INFO] Processing folder: %s", folder.Name)

		// Process each file
		for _, file := range folder.FileList {
			log.Printf("[INFO] Processing file: %s", file.Name)
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
						log.Printf("[ERROR] Failed to rename file %s to %s: %v", file.Name, file.NewName, err)
						result.Success = false
						result.Error = fmt.Errorf("failed to rename file: %w", err)
					} else {
						log.Printf("[INFO] Successfully renamed file %s to %s", file.Name, file.NewName)
					}
				} else {
					log.Printf("[INFO] Dry run: Would rename %s to %s", file.Name, file.NewName)
				}

				// Log the operation if logging is enabled
				if p.query.Logger != nil {
					// Convert paths to absolute
					absOrigPath, err := filepath.Abs(file.Path)
					if err != nil {
						log.Printf("[WARN] Could not get absolute path for %s: %v", file.Path, err)
						absOrigPath = file.Path
					}
					absNewPath, err := filepath.Abs(result.NewPath)
					if err != nil {
						log.Printf("[WARN] Could not get absolute path for %s: %v", result.NewPath, err)
						absNewPath = result.NewPath
					}
					p.query.Logger.LogOperation(absOrigPath, absNewPath, result.Success, result.Error)
				}
			} else {
				result.NewPath = currentPath
				log.Printf("[INFO] No new name generated for %s", file.Name)
			}

			results = append(results, result)
		}
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			log.Printf("[WARN] Failed to close logger: %v", err)
		} else {
			log.Printf("[INFO] Successfully closed logger")
		}
	}

	log.Printf("[INFO] Completed file processing phase")
	return results, nil
}

// copyOriginalStructure copies the entire directory structure and files to the output directory
func (p *SafeProcessor) copyOriginalStructure() error {
	for _, folder := range p.query.Folders {
		// Create corresponding folder in output directory
		outFolder := filepath.Join(p.output, folder.Name)
		if err := os.MkdirAll(outFolder, 0755); err != nil {
			log.Printf("[ERROR] Failed to create output folder %s: %v", outFolder, err)
			return fmt.Errorf("failed to create output folder %s: %w", outFolder, err)
		}
		log.Printf("[INFO] Created output folder: %s", outFolder)

		// Copy each file with original name
		for _, file := range folder.FileList {
			dstPath := filepath.Join(outFolder, file.Name)
			if err := copyFile(file.Path, dstPath); err != nil {
				log.Printf("[ERROR] Failed to copy file %s: %v", file.Path, err)
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
			log.Printf("[INFO] Copied file: %s to %s", file.Path, dstPath)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	log.Printf("[INFO] Copying file from %s to %s", src, dst)
	input, err := os.ReadFile(src)
	if err != nil {
		log.Printf("[ERROR] Failed to read source file %s: %v", src, err)
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, input, 0644); err != nil {
		log.Printf("[ERROR] Failed to write destination file %s: %v", dst, err)
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	log.Printf("[INFO] Successfully copied file")
	return nil
}
