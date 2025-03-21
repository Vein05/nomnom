package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
	files "nomnom/internal/files"
	utils "nomnom/internal/utils"

	log "github.com/charmbracelet/log"

	"github.com/spf13/cobra"
)

type args struct {
	dir         string
	configPath  string
	autoApprove bool
	dryRun      bool
	log         bool
	revert      string
}

var cmdArgs = &args{}

var rootCmd = &cobra.Command{
	Use:   "nomnom",
	Short: "A CLI tool to rename files using AI",
	Long:  `NomNom is a command-line tool that renames files in a folder based on their content using AI models.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Check if revert flag is set
		if cmdArgs.revert != "" {
			opts := files.RevertOptions{
				ChangeLogPath: cmdArgs.revert,
				EnableLogging: cmdArgs.log,
				AutoApprove:   cmdArgs.autoApprove,
			}

			if err := files.ProcessRevert(opts); err != nil {
				log.Error("Error processing revert: ", "error", err)
				os.Exit(1)
			}
			return
		}

		log.Info("[1/6] Loading configuration...")
		// Load configuration
		config := utils.LoadConfig(cmdArgs.configPath)

		log.Info("[2/6] Creating new query...")
		// Create a new query
		query, err := contentprocessors.NewQuery(
			"What is a descriptive name for this file based on its content? Respond with just the filename.",
			cmdArgs.dir,
			cmdArgs.configPath,
			config,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.log,
		)
		if err != nil {
			log.Error("Error creating query: ", "error", err)
			os.Exit(1)
		}

		// Set up output directory
		log.Info("[3/6] Setting up output directory...")
		outputDir := config.Output
		if outputDir == "" {
			outputDir = filepath.Join(cmdArgs.dir, "nomnom", "renamed")
		}

		log.Info("[4/6] Processing files with AI to generate new names...")
		// Process files with AI to get new names
		aiResult, err := aideepseek.HandleAI(config, *query)
		if err != nil {
			log.Error("Error processing with AI: ", "error", err)
			os.Exit(1)
		}
		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		log.Info("[5/6] Processing file renames...")
		processor := contentprocessors.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
		if err != nil {
			log.Error("Error processing files: ", "error", err)
			os.Exit(1)
		}

		// Print processing results
		log.Info("[6/6] Generating summary...")
		fmt.Println("\n📊 Processing Results")
		fmt.Println("══════════════════════")

		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
				if cmdArgs.dryRun {
					log.Info("Would rename:",
						"from", filepath.Base(result.OriginalPath),
						"to", filepath.Base(result.NewPath),
						"status", "🔍 DRY RUN")
				} else {
					log.Info("Renamed:",
						"from", filepath.Base(result.OriginalPath),
						"to", filepath.Base(result.NewPath),
						"status", "✅ DONE")
				}
			} else {
				log.Error("Failed to process:",
					"file", filepath.Base(result.OriginalPath),
					"error", result.Error,
					"status", "❌ ERROR")
			}
		}

		fmt.Printf("\n📈 Summary Stats")
		fmt.Printf("\n═══════════════\n")
		if cmdArgs.dryRun {
			log.Info("Results (Dry Run):",
				"would rename", successCount,
				"failed", len(results)-successCount,
				"total", len(results))
			log.Info("To apply these changes, run:",
				"command", "nomnom -d \""+cmdArgs.dir+"\" --dry-run=false")
		} else {
			log.Info("Results:",
				"renamed", successCount,
				"failed", len(results)-successCount,
				"total", len(results))
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cmdArgs.dir, "dir", "d", "", "Source directory containing files to rename (required when not using revert)")
	rootCmd.Flags().StringVarP(&cmdArgs.configPath, "config", "c", "", "Path to the JSON configuration file. If not provided, the default ~/.config/nomnom/config.json will be used.")
	rootCmd.Flags().BoolVarP(&cmdArgs.autoApprove, "auto-approve", "y", false, "Automatically approve changes without user confirmation")
	rootCmd.Flags().BoolVarP(&cmdArgs.dryRun, "dry-run", "n", true, "Preview changes without actually renaming files")
	rootCmd.Flags().BoolVarP(&cmdArgs.log, "log", "l", true, "Enable logging to file")
	rootCmd.Flags().StringVarP(&cmdArgs.revert, "revert_path", "r", "", "Path to the changes file to revert operations from")

	// Add a PreRunE to validate flags
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmdArgs.revert == "" && cmdArgs.dir == "" {
			return fmt.Errorf("required flag \"dir\" not set when not using revert")
		}
		return nil
	}

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Short:  "Display help",
		Hidden: true,
	})
}
