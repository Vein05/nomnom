package nomnom

import (
	"fmt"
	"strconv"

	utils "nomnom/internal/utils"
	"os"
	"path/filepath"

	log "log"

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

	if prompt == "" {
		if config.AI.Prompt != "" {
			prompt = config.AI.Prompt
		} else {
			prompt = "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."
		}
	}

	folders, err := ProcessDirectory(dir, config)
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
	fmt.Printf("[2/6] Starting safe mode processing \n")
	results := make([]ProcessResult, 0)

	if p.query.DryRun {
		fmt.Printf("[2/6] Dry run: Would create output directory\n")
	} else {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(p.output, 0755); err != nil {
			log.Printf("[2/6] ❌ Failed to create output directory: %v", err)
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
		fmt.Printf("[2/6] Created output directory: %s\n", p.output)
	}

	// First phase: Copy all files with original structure
	if !p.query.DryRun {
		fmt.Printf("[2/6] Starting file copy phase\n")
		if err := p.copyOriginalStructure(); err != nil {
			log.Printf("[2/6] ❌ Failed to copy original structure: %v", err)
			return nil, fmt.Errorf("failed to copy original structure: %w", err)
		}
		fmt.Printf("[2/6] Completed file copy phase\n")
	}

	// Second phase: Process and rename files
	fmt.Printf("[2/6] Starting file processing phase\n")
	for _, folder := range p.query.Folders {
		// Get the output folder path
		outFolder := filepath.Join(p.output, folder.Name)
		fmt.Printf("[2/6] Processing folder: %s, files: %d\n", folder.Name, len(folder.FileList))

		// Process each file
		counter := 0
		for _, file := range folder.FileList {
			fmt.Printf("[2/6] Processing file: %s, size: %s\n", file.Name, file.FormattedSize)
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
					// when auto approve is false, ask the user to approve the rename
					if counter == 0 {
						fmt.Printf("[2/6] Auto approve is: %t\n", p.query.AutoApprove)
					}
					if !p.query.AutoApprove {
						prompt := promptui.Select{
							Label: "Approve rename for " + file.Name + " to " + file.NewName,
							Items: []string{"yes", "no", "approve all"},
						}
						_, result, err := prompt.Run()
						if err != nil {
							log.Printf("[2/6] ❌ Error running prompt: %v", err)
						}
						if result == "no" {
							fmt.Printf("[2/6] Skipping rename for: %s to %s\n", file.Name, file.NewName)
							continue
						}
						if result == "approve all" {
							p.query.AutoApprove = true
							fmt.Printf("[2/6] Auto approving all renames")
						}
					}
					// before renaming, check if the file exists, if it exsits add for n to the end of the new name
					if _, err := os.Stat(result.NewPath); err == nil {
						result.NewPath = filepath.Join(outFolder, file.NewName+"_"+strconv.Itoa(counter))
					}
					if err := os.Rename(currentPath, result.NewPath); err != nil {
						log.Printf("[2/6] ❌ Failed to rename file: %s to %s, error: %v", file.Name, file.NewName, err)
						result.Success = false
						result.Error = fmt.Errorf("failed to rename file: %w", err)
					} else {
						fmt.Printf("[2/6] Successfully renamed file: %s to %s\n", file.Name, file.NewName)
					}
				} else {
					fmt.Printf("[2/6] Dry run: Would rename %s to %s\n", file.Name, file.NewName)
				}

				// Log the operation if logging is enabled
				if p.query.Logger != nil && !p.query.DryRun {
					// Convert paths to absolute
					absOrigPath, err := filepath.Abs(file.Path)
					if err != nil {
						log.Printf("[2/6] ❌ Could not get absolute path for: %s, error: %v", file.Path, err)
						absOrigPath = file.Path
					}
					absNewPath, err := filepath.Abs(result.NewPath)
					if err != nil {
						log.Printf("[2/6] ❌ Could not get absolute path for: %s, error: %v", result.NewPath, err)
						absNewPath = result.NewPath
					}
					p.query.Logger.LogOperation(absOrigPath, absNewPath, result.Success, result.Error)
				}
			} else {
				result.NewPath = currentPath
				fmt.Printf("[2/6] No new name generated for: %s\n", file.Name)
			}

			results = append(results, result)
			counter++
		}
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			log.Printf("[2/6] ❌ Failed to close logger: %v", err)
		} else {
			fmt.Printf("[2/6] Successfully closed logger")
		}
	}

	fmt.Printf("[2/6] Completed file processing phase\n")
	return results, nil
}

// copyOriginalStructure copies the entire directory structure and files to the output directory
func (p *SafeProcessor) copyOriginalStructure() error {
	for _, folder := range p.query.Folders {
		// Create corresponding folder in output directory
		outFolder := filepath.Join(p.output, folder.Name)
		if err := os.MkdirAll(outFolder, 0755); err != nil {
			log.Printf("[2/6] ❌ Failed to create output folder: %s, error: %v", outFolder, err)
			return fmt.Errorf("failed to create output folder %s: %w", outFolder, err)
		}
		fmt.Printf("[2/6] Created output folder: %s\n", outFolder)

		// Copy each file with original name
		for _, file := range folder.FileList {
			dstPath := filepath.Join(outFolder, file.Name)
			if err := copyFile(file.Path, dstPath); err != nil {
				log.Printf("[2/6] ❌ Failed to copy file: %s to %s, error: %v", file.Path, dstPath, err)
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
			fmt.Printf("[2/6] Copied file: %s to %s\n", file.Path, dstPath)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	fmt.Printf("[2/6] Copying file from: %s to %s\n", src, dst)
	input, err := os.ReadFile(src)
	if err != nil {
		log.Printf("[2/6] ❌ Failed to read source file: %s, error: %v", src, err)
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, input, 0644); err != nil {
		log.Printf("[2/6] ❌ Failed to write destination file: %s, error: %v", dst, err)
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	fmt.Printf("[2/6] Successfully copied file from: %s to %s\n", src, dst)
	return nil
}
