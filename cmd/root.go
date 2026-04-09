package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	ai "nomnom/internal/ai"
	content "nomnom/internal/content"
	files "nomnom/internal/files"
	utils "nomnom/internal/utils"

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

		// Load configuration
		config, err := utils.LoadConfig(cmdArgs.configPath, "")
		if err != nil {
			color.Red("Error loading config: %v\n", err)
			os.Exit(1)
		}
		presenter.Divider()
		// Create a new query
		query, err := content.NewQuery(
			cmdArgs.prompt,
			cmdArgs.dir,
			cmdArgs.configPath,
			config,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.log,
			cmdArgs.organize,
			presenter,
			presenter,
		)
		if err != nil {
			color.Red("Error creating query: %v\n", err)
			os.Exit(1)
		}
		presenter.Divider()

		// Set up output directory
		outputDir := config.Output
		if outputDir == "" {
			outputDir = filepath.Join(cmdArgs.dir, "nomnom", "renamed")
		}

		output_text := fmt.Sprintf("Output directory set up at: %s", outputDir)
		if cmdArgs.dryRun {
			output_text = fmt.Sprintf("Output directory would be set up at: %s", outputDir)
		}
		presenter.Titlef(output_text)

		presenter.Divider()

		presenter.Titlef("Processing files with AI to generate new names")

		// Process files with AI to get new names
		aiResult, err := ai.HandleAI(config, *query)
		if err != nil {
			color.Red("Error processing files with AI: %v\n", err)
			os.Exit(1)
		}

		presenter.Divider()

		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		presenter.Titlef("Processing file renames")

		presenter.Divider()

		processor := content.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
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
