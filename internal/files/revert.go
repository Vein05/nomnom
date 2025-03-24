package nomnom

import (
	"fmt"
	"os"
	"path/filepath"

	utils "nomnom/internal/utils"

	log "log"

	"github.com/manifoldco/promptui"
)

// RevertOptions contains the configuration for the revert operation
type RevertOptions struct {
	ChangeLogPath string
	EnableLogging bool
	AutoApprove   bool
}

// ProcessRevert handles the revert operation for files that were previously renamed
func ProcessRevert(opts RevertOptions) error {
	fmt.Printf("[1/3] Loading changes file...")
	changeLog, err := utils.LoadLog(opts.ChangeLogPath)
	if err != nil {
		return err
	}

	// Use the directory of the first entry as the base directory for logs
	var baseDir string
	if len(changeLog.Entries) > 0 {
		baseDir = filepath.Dir(changeLog.Entries[0].OriginalPath)
	} else {
		baseDir = "."
	}

	fmt.Printf("[2/3] Setting up revert logger...")
	// Create revert directory
	revertDir := filepath.Join(baseDir, "nomnom", "reverted", changeLog.SessionID)
	if err := os.MkdirAll(revertDir, 0755); err != nil {
		return err
	}

	logger, err := utils.NewLogger(opts.EnableLogging, baseDir)
	if err != nil {
		return err
	}
	defer logger.Close()

	fmt.Printf("[3/3] Reverting changes...")
	for _, entry := range changeLog.Entries {
		if entry.Success {
			// Calculate the new path in the revert directory
			relPath, err := filepath.Rel(baseDir, entry.OriginalPath)
			if err != nil {
				log.Printf("❌ Error calculating relative path for: %s, error: %v", entry.OriginalPath, err)
				logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
				continue
			}
			revertPath := filepath.Join(revertDir, relPath)

			// Prompt user for approval if auto-approve is not enabled
			if !opts.AutoApprove {
				prompt := promptui.Select{
					Label: "Approve revert for " + filepath.Base(entry.NewPath) + " to " + filepath.Base(revertPath),
					Items: []string{"yes", "no", "approve all"},
				}
				_, result, err := prompt.Run()
				if err != nil {
					log.Printf("❌ Error running prompt: %v", err)
					continue
				}
				if result == "no" {
					fmt.Printf("Skipping revert for: %s", filepath.Base(entry.NewPath))
					continue
				}
				if result == "approve all" {
					opts.AutoApprove = true
					fmt.Printf("Auto approving all reverts")
				}
			}

			// Create necessary directories
			if err := os.MkdirAll(filepath.Dir(revertPath), 0755); err != nil {
				log.Printf("❌ Error creating directory for: %s, error: %v", revertPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Copy file to revert location
			input, err := os.ReadFile(entry.NewPath)
			if err != nil {
				log.Printf("❌ Error reading file: %s, error: %v", entry.NewPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			if err := os.WriteFile(revertPath, input, 0644); err != nil {
				log.Printf("❌ Error writing file: %s, error: %v", revertPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Log successful revert operation
			logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, true, nil)
			fmt.Printf("Reverted: %s to %s, status: ✅ DONE", filepath.Base(entry.NewPath), filepath.Base(revertPath))
		}
	}

	fmt.Printf("Revert operation completed. Files have been placed in: %s", revertDir)
	return nil
}
