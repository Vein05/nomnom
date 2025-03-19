package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
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
			log.Info("[1/3] Loading changes file...")
			changeLog, err := utils.LoadLog(cmdArgs.revert)
			if err != nil {
				log.Error("Error loading changes file: %v", err)
				os.Exit(1)
			}

			log.Info("[2/3] Setting up revert logger...")
			// Create a new logger for the revert operation
			// Use the directory of the first entry as the base directory for logs
			var baseDir string
			if len(changeLog.Entries) > 0 {
				baseDir = filepath.Dir(changeLog.Entries[0].OriginalPath)
			} else {
				baseDir = "."
			}
			logger, err := utils.NewLogger(cmdArgs.log, baseDir)
			if err != nil {
				log.Error("Error creating logger: %v", err)
				os.Exit(1)
			}
			defer logger.Close()

			log.Info("[3/3] Reverting changes...")
			for _, entry := range changeLog.Entries {
				if entry.Success {
					// Create necessary directories
					if err := os.MkdirAll(filepath.Dir(entry.OriginalPath), 0755); err != nil {
						log.Error("Error creating directory for %s: %v", entry.OriginalPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					// Copy file back to original location
					input, err := os.ReadFile(entry.NewPath)
					if err != nil {
						log.Error("Error reading file %s: %v", entry.NewPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					if err := os.WriteFile(entry.OriginalPath, input, 0644); err != nil {
						log.Error("Error writing file %s: %v", entry.OriginalPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					// Log successful revert operation
					logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, true, nil)
					log.Info("Reverted: %s -> %s", filepath.Base(entry.NewPath), filepath.Base(entry.OriginalPath))
				}
			}
			log.Info("Revert operation completed.")
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
			log.Error("Error creating query: %v", err)
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
			log.Error("Error processing with AI: %v", err)
			os.Exit(1)
		}
		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		log.Info("[5/6] Processing file renames...")
		processor := contentprocessors.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
		if err != nil {
			log.Error("Error processing files: %v", err)
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
						"status", "--")
				} else {
					log.Info("Renamed:",
						"from", filepath.Base(result.OriginalPath),
						"to", filepath.Base(result.NewPath),
						"status", "✅ SUCCESS")
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
		log.Info("Results:",
			"successful", successCount,
			"failed", len(results)-successCount,
			"total", len(results))

		if cmdArgs.dryRun {
			log.Info("This was a dry run - no files were modified",
				"note", "Run with --dry-run=false to apply changes")
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
	rootCmd.Flags().StringVarP(&cmdArgs.revert, "revert", "r", "", "Path to the changes file to revert operations from")

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
