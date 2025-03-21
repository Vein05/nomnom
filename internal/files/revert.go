package nomnom

import (
	"os"
	"path/filepath"

	utils "nomnom/internal/utils"

	log "github.com/charmbracelet/log"
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
	log.Info("[1/3] Loading changes file...")
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

	log.Info("[2/3] Setting up revert logger...")
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

	log.Info("[3/3] Reverting changes...")
	for _, entry := range changeLog.Entries {
		if entry.Success {
			// Calculate the new path in the revert directory
			relPath, err := filepath.Rel(baseDir, entry.OriginalPath)
			if err != nil {
				log.Error("Error calculating relative path for: ", "path", entry.OriginalPath, "error", err)
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
					log.Error("Error running prompt: ", "error", err)
					continue
				}
				if result == "no" {
					log.Info("Skipping revert for: ", "file", filepath.Base(entry.NewPath))
					continue
				}
				if result == "approve all" {
					opts.AutoApprove = true
					log.Info("Auto approving all reverts")
				}
			}

			// Create necessary directories
			if err := os.MkdirAll(filepath.Dir(revertPath), 0755); err != nil {
				log.Error("Error creating directory for: ", "path", revertPath, "error", err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Copy file to revert location
			input, err := os.ReadFile(entry.NewPath)
			if err != nil {
				log.Error("Error reading file: ", "path", entry.NewPath, "error", err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			if err := os.WriteFile(revertPath, input, 0644); err != nil {
				log.Error("Error writing file: ", "path", revertPath, "error", err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Log successful revert operation
			logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, true, nil)
			log.Info("Reverted:",
				"from", filepath.Base(entry.NewPath),
				"to", filepath.Base(revertPath),
				"status", "✅ DONE")
		}
	}

	log.Info("Revert operation completed. Files have been placed in: ", "path", revertDir)
	return nil
}
