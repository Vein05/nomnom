package cmd

import (
	"fmt"
	"path/filepath"
	"slices"

	app "nomnom/internal/app"
	"nomnom/internal/utils"

	"github.com/spf13/cobra"
)

var analyticsDir string

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Show local NomNom analytics for a directory",
	Example: `nomnom analytics -d ~/Downloads
nomnom analytics --dir /path/to/files`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := newCLIPresenter()
		presenter.Banner()
		presenter.Divider()

		rootDir, err := filepath.Abs(analyticsDir)
		if err != nil {
			return fmt.Errorf("resolve analytics directory: %w", err)
		}

		service := app.NewService()
		summary, sessions, err := service.LoadAnalytics(rootDir)
		if err != nil {
			return err
		}

		presenter.Titlef("Analytics Summary")
		presenter.Infof("Directory: %s", rootDir)
		presenter.Infof("Sessions: %d", summary.Sessions)
		presenter.Infof("Files scanned: %d", summary.FilesScanned)
		presenter.Infof("Planned renames: %d", summary.PlannedRenames)
		presenter.Infof("Attempted renames: %d", summary.AttemptedRenames)
		presenter.Infof("Successful renames: %d", summary.SuccessfulRenames)
		presenter.Infof("Failed renames: %d", summary.FailedRenames)
		if !summary.UpdatedAt.IsZero() {
			presenter.Infof("Last updated: %s", summary.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
		}

		presenter.Divider()
		presenter.Titlef("Model Usage")
		models := sortedModelUsage(summary.Models)
		if len(models) == 0 {
			presenter.Warnf("No model usage has been recorded yet.")
		} else {
			for _, model := range models {
				presenter.Infof(
					"%s/%s: requests=%d vision=%d tokens=%d (prompt=%d completion=%d)",
					model.Provider,
					model.Model,
					model.Requests,
					model.VisionRequests,
					model.TotalTokens,
					model.PromptTokens,
					model.CompletionTokens,
				)
			}
		}

		presenter.Divider()
		presenter.Titlef("Recent Sessions")
		if len(sessions) == 0 {
			presenter.Warnf("No analytics sessions found under %s", filepath.Join(rootDir, ".nomnom", "analytics", "sessions"))
			return nil
		}

		limit := min(5, len(sessions))
		for _, session := range sessions[:limit] {
			presenter.Infof(
				"%s | scanned=%d planned=%d renamed=%d failed=%d dry_run=%t",
				session.StartedAt.Local().Format("2006-01-02 15:04:05"),
				session.FilesScanned,
				session.PlannedRenames,
				session.SuccessfulRenames,
				session.FailedRenames,
				session.DryRun,
			)
		}

		return nil
	},
}

func init() {
	analyticsCmd.Flags().StringVarP(&analyticsDir, "dir", "d", "", "Directory containing .nomnom analytics")
	analyticsCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(analyticsCmd)
}

func sortedModelUsage(models map[string]utils.ModelAnalytics) []utils.ModelAnalytics {
	ordered := make([]utils.ModelAnalytics, 0, len(models))
	for _, model := range models {
		ordered = append(ordered, model)
	}

	slices.SortFunc(ordered, func(a, b utils.ModelAnalytics) int {
		if a.TotalTokens == b.TotalTokens {
			if a.Provider == b.Provider {
				switch {
				case a.Model < b.Model:
					return -1
				case a.Model > b.Model:
					return 1
				default:
					return 0
				}
			}
			switch {
			case a.Provider < b.Provider:
				return -1
			case a.Provider > b.Provider:
				return 1
			default:
				return 0
			}
		}
		if a.TotalTokens > b.TotalTokens {
			return -1
		}
		return 1
	})

	return ordered
}
