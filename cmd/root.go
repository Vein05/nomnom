package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
	utils "nomnom/internal/utils"

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
			fmt.Println("\n[1/3] Loading changes file...")
			changeLog, err := utils.LoadLog(cmdArgs.revert)
			if err != nil {
				fmt.Printf("Error loading changes file: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("[2/3] Setting up revert logger...")
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
				fmt.Printf("Error creating logger: %v\n", err)
				os.Exit(1)
			}
			defer logger.Close()

			fmt.Println("[3/3] Reverting changes...")
			for _, entry := range changeLog.Entries {
				if entry.Success {
					// Create necessary directories
					if err := os.MkdirAll(filepath.Dir(entry.OriginalPath), 0755); err != nil {
						fmt.Printf("Error creating directory for %s: %v\n", entry.OriginalPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					// Copy file back to original location
					input, err := os.ReadFile(entry.NewPath)
					if err != nil {
						fmt.Printf("Error reading file %s: %v\n", entry.NewPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					if err := os.WriteFile(entry.OriginalPath, input, 0644); err != nil {
						fmt.Printf("Error writing file %s: %v\n", entry.OriginalPath, err)
						logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, false, err)
						continue
					}

					// Log successful revert operation
					logger.LogOperationWithType(entry.NewPath, entry.OriginalPath, utils.OperationRevert, true, nil)
					fmt.Printf("Reverted: %s -> %s\n", filepath.Base(entry.NewPath), filepath.Base(entry.OriginalPath))
				}
			}
			fmt.Println("\nRevert operation completed.")
			return
		}

		fmt.Println("\n[1/6] Loading configuration...")
		// Load configuration
		config := utils.LoadConfig(cmdArgs.configPath)

		fmt.Println("[2/6] Creating new query...")
		// Create a new query
		query, err := contentprocessors.NewQuery(
			"What is a descriptive name for this file based on its content? Respond with just the filename.",
			cmdArgs.dir,
			cmdArgs.configPath,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.log,
		)
		if err != nil {
			fmt.Printf("Error creating query: %v\n", err)
			os.Exit(1)
		}

		// Set up output directory
		fmt.Println("[3/6] Setting up output directory...")
		outputDir := config.Output
		if outputDir == "" {
			outputDir = filepath.Join(cmdArgs.dir, "renamed")
		}

		fmt.Println("[4/6] Processing files with AI to generate new names...")
		// Process files with AI to get new names
		aiResult, err := aideepseek.HandleAI(config, *query)
		if err != nil {
			fmt.Printf("Error processing with AI: %v\n", err)
			os.Exit(1)
		}
		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		fmt.Println("[5/6] Processing file renames...")
		processor := contentprocessors.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
		if err != nil {
			fmt.Printf("Error processing files: %v\n", err)
			os.Exit(1)
		}

		// Print processing results
		fmt.Println("\n[6/6] Generating summary...")
		fmt.Println("\nProcessing Results:")
		fmt.Println("===================")

		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
				if cmdArgs.dryRun {
					fmt.Printf("Would rename: %s -> %s\n", filepath.Base(result.OriginalPath), filepath.Base(result.NewPath))
				} else {
					fmt.Printf("Renamed: %s -> %s\n", filepath.Base(result.OriginalPath), filepath.Base(result.NewPath))
				}
			} else {
				fmt.Printf("Error processing %s: %v\n", filepath.Base(result.OriginalPath), result.Error)
			}
		}

		fmt.Printf("\nSummary: Successfully processed %d/%d files\n", successCount, len(results))
		if cmdArgs.dryRun {
			fmt.Println("This was a dry run. No files were actually modified.")
			fmt.Println("To perform actual changes, run with --dry-run=false")
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
	rootCmd.Flags().StringVarP(&cmdArgs.configPath, "config", "c", "config.json", "Path to the JSON configuration file")
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
