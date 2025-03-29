package nomnom

import (
	"fmt"

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
	Organize    bool
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

type FileTypeCategory struct {
	Name       string
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

// NewQuery creates a new Query object with the given parameters.
func NewQuery(prompt string, dir string, configPath string, config utils.Config, autoApprove bool, dryRun bool, log bool, organize bool) (*Query, error) {

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
		Organize:    organize,
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
		if !p.query.Organize {
			fmt.Printf("[2/6] Starting file copy phase in original structure.\n")
			if err := p.copyOriginalStructure(); err != nil {
				log.Printf("[2/6] ❌ Failed to copy original structure: %v", err)
				return nil, fmt.Errorf("failed to copy original structure: %w", err)
			}
		} else {
			fmt.Printf("[2/6] Starting file copy phase in Organized structure\n")
			if err := p.copyOrganizedStructure(); err != nil {
				log.Printf("[2/6] ❌ Failed to copy organized structure: %v", err)
				return nil, fmt.Errorf("failed to copy organized structure: %w", err)
			}
		}
		fmt.Printf("[2/6] Completed file copy phase\n")
	}

	// Second phase: Process and rename files
	fmt.Printf("[2/6] Starting file processing phase\n")
	for _, folder := range p.query.Folders {
		fmt.Printf("[2/6] Processing folder: %s, files: %d\n", folder.Name, len(folder.FileList))

		counter := 0
		for _, file := range folder.FileList {
			// Define all path structures at the start of the loop
			category := getCategoryForFile(file.Name)
			baseOutputPath := filepath.Join(p.output, folder.Name)

			var (
				currentPath string // Path of the copied file before rename
				newPath     string // Destination path after rename
			)

			if p.query.Organize {
				currentPath = filepath.Join(p.output, category, file.Name)
				newPath = filepath.Join(p.output, category, file.NewName)
			} else {
				currentPath = filepath.Join(baseOutputPath, file.Name)
				newPath = filepath.Join(baseOutputPath, file.NewName)
			}

			// Handle duplicate filenames
			if _, err := os.Stat(newPath); err == nil {
				newPath = utils.GenerateUniqueFilename(newPath, counter)
				fmt.Printf("[2/6] Duplicate file detected, renaming to: %s\n", newPath)

			}

			fmt.Printf("[2/6] Processing file: %s\n", file.Name)
			fmt.Printf("  - Current path: %s\n", currentPath)
			fmt.Printf("  - New path: %s\n", newPath)

			result := ProcessResult{
				OriginalPath: file.Path,
				NewPath:      newPath,
				Success:      true,
			}

			// Skip if no new name was generated
			if file.NewName == "" {
				fmt.Printf("[2/6] No new name generated for: %s\n", file.Name)
				results = append(results, result)
				continue
			}

			// In non-dry-run mode, rename the copied file
			if !p.query.DryRun {
				if !p.query.AutoApprove && counter == 0 {
					fmt.Printf("[2/6] Auto approve is disabled\n")
				}

				if !p.query.AutoApprove {
					response, err := p.promptForRenameApproval(file.Name, file.NewName)
					if err != nil {
						log.Printf("[2/6] ❌ Error running prompt: %v", err)
					}
					if response == "no" {
						fmt.Printf("[2/6] Skipping rename for: %s\n", file.Name)
						continue
					}
					if response == "approve all" {
						p.query.AutoApprove = true
						fmt.Printf("[2/6] Auto approving all renames\n")
					}
				}

				if err := os.Rename(currentPath, newPath); err != nil {
					log.Printf("[2/6] ❌ Failed to rename file: %v", err)
					result.Success = false
					result.Error = fmt.Errorf("failed to rename file: %w", err)
				} else {
					fmt.Printf("[2/6] Successfully renamed file\n")
				}
			} else {
				fmt.Printf("[2/6] Dry run: Would rename %s to %s\n", file.Name, file.NewName)
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
	}

	// Close the logger if it exists
	if p.query.Logger != nil {
		if err := p.query.Logger.Close(); err != nil {
			log.Printf("[2/6] ❌ Failed to close logger: %v", err)
		} else {
			fmt.Printf("[2/6] Successfully closed logger\n")
		}
	}

	fmt.Printf("[2/6] Completed file processing phase\n")
	return results, nil
}

// promptForRenameApproval handles the user prompt for rename approval
func (p *SafeProcessor) promptForRenameApproval(oldName, newName string) (string, error) {
	prompt := promptui.Select{
		Label: fmt.Sprintf("Approve rename for %s to %s", oldName, newName),
		Items: []string{"yes", "no", "approve all"},
	}
	_, result, err := prompt.Run()
	return result, err
}

func getCategoryForFile(fileName string) string {
	ext := filepath.Ext(fileName)
	for _, category := range defaultCategories {
		for _, categoryExt := range category.Extensions {
			if categoryExt == ext {
				return category.Name
			}
		}
	}
	return "Others"
}

func (p *SafeProcessor) copyOrganizedStructure() error {
	// Create category folders
	for _, category := range defaultCategories {
		categoryPath := filepath.Join(p.output, category.Name)
		if err := os.MkdirAll(categoryPath, 0755); err != nil {
			return fmt.Errorf("failed to create category folder %s: %w", categoryPath, err)
		}
	}

	// Copy files into appropriate category folders
	for _, folder := range p.query.Folders {
		for i, file := range folder.FileList {
			category := getCategoryForFile(file.Name)
			dstPath := filepath.Join(p.output, category, file.Name)
			// query should be updated to reflect the new path
			folder.FileList[i].Path = dstPath
			if err := copyFile(file.Path, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", file.Path, err)
			}
		}
	}
	return nil
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
