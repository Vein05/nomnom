package cmd

import (
	"fmt"
	"os"

	app "nomnom/internal/app"
	files "nomnom/internal/files"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type args struct {
	dir         string
	configPath  string
	autoApprove bool
	dryRun      bool
	log         bool
	revert      string
	organize    bool
	prompt      string
}

var cmdArgs = &args{}

var rootCmd = &cobra.Command{
	Use:   "nomnom",
	Short: "A Go CLI tool to bulk rename and organize files using AI.",
	Long:  `NomNom is a command-line tool that renames files in a folder based on their content using AI models.`,
	Example: `nomnom setup
nomnom analytics -d ~/Documents/files
nomnom -d ~/Documents/files
nomnom -d ~/Documents/files -n=false
nomnom -d ~/Documents/files -p research
nomnom -r .nomnom/logs/changes_123.json`,
	Run: func(cmd *cobra.Command, _ []string) {
		presenter := newCLIPresenter()
		presenter.Banner()
		presenter.Divider()

		// Check if revert flag is set
		if cmdArgs.revert != "" {
			opts := files.RevertOptions{
				ChangeLogPath: cmdArgs.revert,
				EnableLogging: cmdArgs.log,
				AutoApprove:   cmdArgs.autoApprove,
				Reporter:      presenter,
				Approver:      presenter,
			}

			if err := files.ProcessRevert(opts); err != nil {
				color.Red("Error processing revert: %v\n", err)
				os.Exit(1)
			}
			return
		}

		service := app.NewService()
		presenter.Divider()
		run, err := service.PrepareRun(app.RunOptions{
			Dir:         cmdArgs.dir,
			ConfigPath:  cmdArgs.configPath,
			Prompt:      cmdArgs.prompt,
			AutoApprove: cmdArgs.autoApprove,
			DryRun:      cmdArgs.dryRun,
			Log:         cmdArgs.log,
			Organize:    cmdArgs.organize,
		}, presenter, presenter)
		if err != nil {
			color.Red("Error preparing run: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := run.Close(); err != nil {
				color.Red("Error closing run resources: %v\n", err)
			}
		}()
		presenter.Divider()

		outputText := fmt.Sprintf("Output directory set up at: %s", run.OutputDir)
		if cmdArgs.dryRun {
			outputText = fmt.Sprintf("Output directory would be set up at: %s", run.OutputDir)
		}
		presenter.Titlef(outputText)

		presenter.Divider()

		presenter.Titlef("Processing files with AI to generate new names")

		if err := service.GeneratePlan(run); err != nil {
			color.Red("Error processing files with AI: %v\n", err)
			os.Exit(1)
		}

		presenter.Divider()

		presenter.Titlef("Processing file renames")

		presenter.Divider()

		results, err := service.ApplyPlan(run)
		if err != nil {
			color.Red("Error processing files: %v\n", err)
			os.Exit(1)
		}
		presenter.Divider()
		presenter.Titlef("Processing files with AI to generate new names")
		presenter.Divider()

		successCount := presenter.PrintResults(results, cmdArgs.dryRun)

		presenter.Divider()

		if cmdArgs.dryRun {
			color.Green("\n%s %d files would be renamed successfully.\n", ("✅"), successCount)
			color.Yellow("\nTo apply these changes, run: nomnom -d \"%s\" --dry-run=false\n", cmdArgs.dir)
		} else {
			presenter.PrintSummary(results)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	Init()
}
