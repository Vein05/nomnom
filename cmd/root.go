package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
	files "nomnom/internal/files"
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
			opts := files.RevertOptions{
				ChangeLogPath: cmdArgs.revert,
				EnableLogging: cmdArgs.log,
				AutoApprove:   cmdArgs.autoApprove,
			}

			if err := files.ProcessRevert(opts); err != nil {
				fmt.Printf("Error processing revert: %v\n", err)
				os.Exit(1)
			}
			return
		}

		fmt.Printf("[1/6] Loading configuration...\n")
		// Load configuration
		config := utils.LoadConfig(cmdArgs.configPath)

		fmt.Printf("[2/6] Creating new query...\n")
		// Create a new query
		query, err := contentprocessors.NewQuery(
			"",
			cmdArgs.dir,
			cmdArgs.configPath,
			config,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.log,
		)
		if err != nil {
			fmt.Printf("Error creating query: %v\n", err)
			os.Exit(1)
		}

		// Set up output directory
		fmt.Printf("[3/6] Setting up output directory...\n")
		outputDir := config.Output
		if outputDir == "" {
			outputDir = filepath.Join(cmdArgs.dir, "nomnom", "renamed")
		}

		fmt.Printf("[4/6] Processing files with AI to generate new names...\n")
		// Process files with AI to get new names
		aiResult, err := aideepseek.HandleAI(config, *query)
		if err != nil {
			fmt.Printf("Error processing with AI: %v\n", err)
			os.Exit(1)
		}

		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		fmt.Printf("[5/6] Processing file renames...\n")
		processor := contentprocessors.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
		if err != nil {
			fmt.Printf("Error processing files: %v\n", err)
			os.Exit(1)
		}

		// Print processing results
		fmt.Printf("[6/6] Generating summary...\n")
		fmt.Printf("\n📊 Processing Results\n")
		fmt.Printf("══════════════════════\n")

		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
				if cmdArgs.dryRun {
					fmt.Printf("🔍 Would rename: %s → %s\n",
						filepath.Base(result.OriginalPath),
						filepath.Base(result.NewPath))
				} else {
					fmt.Printf("✅ Renamed: %s → %s\n",
						filepath.Base(result.OriginalPath),
						filepath.Base(result.NewPath))
				}
			} else {
				fmt.Printf("❌ Failed to process: %s (Error: %v)\n",
					filepath.Base(result.OriginalPath),
					result.Error)
			}
		}

		fmt.Printf("\n📈 Summary Stats")
		fmt.Printf("\n═══════════════\n")
		if cmdArgs.dryRun {
			fmt.Printf("Would rename: %d files\n", successCount)
			fmt.Printf("Failed: %d files\n", len(results)-successCount)
			fmt.Printf("Total: %d files\n", len(results))
			fmt.Printf("\nTo apply these changes, run: nomnom -d \"%s\" --dry-run=false\n", cmdArgs.dir)
		} else {
			fmt.Printf("Renamed: %d files\n", successCount)
			fmt.Printf("Failed: %d files\n", len(results)-successCount)
			fmt.Printf("Total: %d files\n", len(results))
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
