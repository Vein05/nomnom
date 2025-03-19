package nomnom

import (
	"fmt"
	"strconv"

	utils "nomnom/internal/utils"
	"os"
	"path/filepath"

	log "github.com/charmbracelet/log"
	"github.com/manifoldco/promptui"
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
func NewQuery(prompt string, dir string, configPath string, config utils.Config, autoApprove bool, dryRun bool, log bool) (*Query, error) {
	if prompt == "" || config.AI.Prompt == "" {
		prompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."
	}

	folders, err := ProcessDirectory(dir, config)
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
	log.Info("Starting safe mode processing")
	results := make([]ProcessResult, 0)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(p.output, 0755); err != nil {
		log.Error("Failed to create output directory: %v", err)
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	log.Info("Created output directory:", "directory", p.output)

	// First phase: Copy all files with original structure
	if !p.query.DryRun {
		log.Info("Starting file copy phase")
		if err := p.copyOriginalStructure(); err != nil {
			log.Error("Failed to copy original structure: %v", err)
			return nil, fmt.Errorf("failed to copy original structure: %w", err)
		}
		log.Info("Completed file copy phase")
	}

	// Second phase: Process and rename files
	log.Info("Starting file processing phase")
	for _, folder := range p.query.Folders {
		// Get the output folder path
		outFolder := filepath.Join(p.output, folder.Name)
		log.Info("Processing folder: ", "folder", folder.Name, "files", len(folder.FileList))

		// Process each file
		counter := 0
		for _, file := range folder.FileList {
			log.Info("Processing file: ", "file", file.Name, "size", file.FormattedSize)
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
					// create a promptui prompt for the user to approve the rename
					// when auto approve is false, ask the user to approve the rename
					if counter == 0 {
						log.Info("Auto approve is: ", "autoApprove", p.query.AutoApprove)
					}
					if !p.query.AutoApprove {
						prompt := promptui.Select{
							Label: "Approve rename for " + file.Name + " to " + file.NewName,
							Items: []string{"yes", "no", "approve all"},
						}
						_, result, err := prompt.Run()
						if err != nil {
							log.Error("Error running prompt: ", "error", err)
						}
						if result == "no" {
							log.Info("Skipping rename for: ", "file", file.Name, "to", file.NewName)
							continue
						}
						if result == "approve all" {
							p.query.AutoApprove = true
							log.Info("Auto approving all renames")
						}
					}
					// before renaming, check if the file exists, if it exsits add for n to the end of the new name
					if _, err := os.Stat(result.NewPath); err == nil {
						result.NewPath = filepath.Join(outFolder, file.NewName+"_"+strconv.Itoa(counter))
					}
					if err := os.Rename(currentPath, result.NewPath); err != nil {
						log.Error("Failed to rename file: ", "file", file.Name, "to", file.NewName, "error", err)
						result.Success = false
						result.Error = fmt.Errorf("failed to rename file: %w", err)
					} else {
						log.Info("Successfully renamed file: ", "file", file.Name, "to", file.NewName)
					}
				} else {
					log.Info("Dry run: Would rename ", "file", file.Name, "to", file.NewName)
				}

				// Log the operation if logging is enabled
				if p.query.Logger != nil {
					// Convert paths to absolute
					absOrigPath, err := filepath.Abs(file.Path)
					if err != nil {
						log.Error("Could not get absolute path for: ", "path", file.Path, "error", err)
						absOrigPath = file.Path
					}
					absNewPath, err := filepath.Abs(result.NewPath)
					if err != nil {
						log.Error("Could not get absolute path for: ", "path", result.NewPath, "error", err)
						absNewPath = result.NewPath
					}
					p.query.Logger.LogOperation(absOrigPath, absNewPath, result.Success, result.Error)
				}
			} else {
				result.NewPath = currentPath
				log.Info("No new name generated for: ", "file", file.Name)
			}

			results = append(results, result)
			counter++
		}
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			log.Error("Failed to close logger: ", "error", err)
		} else {
			log.Info("Successfully closed logger")
		}
	}

	log.Info("Completed file processing phase")
	return results, nil
}

// copyOriginalStructure copies the entire directory structure and files to the output directory
func (p *SafeProcessor) copyOriginalStructure() error {
	for _, folder := range p.query.Folders {
		// Create corresponding folder in output directory
		outFolder := filepath.Join(p.output, folder.Name)
		if err := os.MkdirAll(outFolder, 0755); err != nil {
			log.Error("Failed to create output folder: ", "folder", outFolder, "error", err)
			return fmt.Errorf("failed to create output folder %s: %w", outFolder, err)
		}
		log.Info("Created output folder: ", "folder", outFolder)

		// Copy each file with original name
		for _, file := range folder.FileList {
			dstPath := filepath.Join(outFolder, file.Name)
			if err := copyFile(file.Path, dstPath); err != nil {
				log.Error("Failed to copy file: ", "src", file.Path, "dst", dstPath, "error", err)
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
			log.Info("Copied file: ", "src", file.Path, "dst", dstPath)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	log.Info("Copying file from: ", "src", src, "dst", dst)
	input, err := os.ReadFile(src)
	if err != nil {
		log.Error("Failed to read source file: ", "src", src, "err", err)
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, input, 0644); err != nil {
		log.Error("Failed to write destination file: ", "dst", dst, "err", err)
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	log.Info("Successfully copied file from: ", "src", src, "dst", dst)
	return nil
}
