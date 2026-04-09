package files

import (
	"fmt"
	"os"
	"path/filepath"

	utils "nomnom/internal/utils"
)

// RevertOptions contains the configuration for the revert operation
type RevertOptions struct {
	ChangeLogPath string
	EnableLogging bool
	AutoApprove   bool
	Reporter      utils.Reporter
	Approver      utils.Approver
}

// ProcessRevert handles the revert operation for files that were previously renamed
func ProcessRevert(opts RevertOptions) error {
	reporter := opts.Reporter
	if reporter == nil {
		reporter = utils.NopReporter{}
	}

	reporter.Infof("[1/3] Loading changes file...")
	changeLog, err := utils.LoadLog(opts.ChangeLogPath)
	if err != nil {
		return err
	}

	// Use the directory of the first entry as the base directory for logs
	var baseDir string
	if len(changeLog.Entries) > 0 {
		baseDir = changeLog.Entries[0].BaseDir
		if baseDir == "" {
			baseDir = filepath.Dir(changeLog.Entries[0].OriginalPath)
		}
	} else {
		baseDir = "."
	}

	reporter.Infof("[2/3] Setting up revert logger...")
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

	reporter.Infof("[3/3] Reverting changes...")
	for _, entry := range changeLog.Entries {
		if entry.Success {
			// Calculate the new path in the revert directory
			relPath, err := filepath.Rel(baseDir, entry.OriginalPath)
			if err != nil {
				reporter.Errorf("Error calculating relative path for: %s, error: %v", entry.OriginalPath, err)
				logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
				continue
			}
			revertPath := filepath.Join(revertDir, relPath)

			// Prompt user for approval if auto-approve is not enabled
			if !opts.AutoApprove {
				if opts.Approver == nil {
					return fmt.Errorf("no approver configured")
				}
				result, err := opts.Approver.Approve("revert", filepath.Base(entry.NewPath), filepath.Base(revertPath))
				if err != nil {
					reporter.Errorf("Error running prompt: %v", err)
					continue
				}
				if result == utils.ApprovalNo {
					reporter.Warnf("Skipping revert for: %s", filepath.Base(entry.NewPath))
					continue
				}
				if result == utils.ApprovalAll {
					opts.AutoApprove = true
					reporter.Infof("Auto approving all reverts")
				}
			}

			// Create necessary directories
			if err := os.MkdirAll(filepath.Dir(revertPath), 0755); err != nil {
				reporter.Errorf("Error creating directory for: %s, error: %v", revertPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Copy file to revert location
			input, err := os.ReadFile(entry.NewPath)
			if err != nil {
				reporter.Errorf("Error reading file: %s, error: %v", entry.NewPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			if err := os.WriteFile(revertPath, input, 0644); err != nil {
				reporter.Errorf("Error writing file: %s, error: %v", revertPath, err)
				logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, false, err)
				continue
			}

			// Log successful revert operation
			logger.LogOperationWithType(entry.NewPath, revertPath, utils.OperationRevert, true, nil)
			reporter.Infof("Reverted: %s to %s, status: done", filepath.Base(entry.NewPath), filepath.Base(revertPath))
		}
	}

	reporter.Infof("Revert operation completed. Files have been placed in: %s", revertDir)
	return nil
}
