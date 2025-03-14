package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	"github.com/spf13/cobra"
)

type args struct {
	dir         string
	configPath  string
	autoApprove bool
	dryRun      bool
	verbose     bool
}

var cmdArgs = &args{}

var rootCmd = &cobra.Command{
	Use:   "nomnom",
	Short: "A CLI tool to rename files using AI",
	Long:  `NomNom is a command-line tool that renames files in a folder based on their content using AI models.`,
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Println("\n[1/6] Loading configuration...")
		// Load configuration
		config := configutils.LoadConfig(cmdArgs.configPath)

		fmt.Println("[2/6] Creating new query...")
		// Create a new query
		query, err := contentprocessors.NewQuery(
			"What is a descriptive name for this file based on its content? Respond with just the filename.",
			cmdArgs.dir,
			cmdArgs.configPath,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.verbose,
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
	rootCmd.Flags().StringVarP(&cmdArgs.dir, "dir", "d", "", "Source directory containing files to rename (required)")
	rootCmd.Flags().StringVarP(&cmdArgs.configPath, "config", "c", "config.json", "Path to the JSON configuration file")
	rootCmd.Flags().BoolVarP(&cmdArgs.autoApprove, "auto-approve", "y", false, "Automatically approve changes without user confirmation")
	rootCmd.Flags().BoolVarP(&cmdArgs.dryRun, "dry-run", "n", true, "Preview changes without actually renaming files")
	rootCmd.Flags().BoolVarP(&cmdArgs.verbose, "verbose", "v", false, "Enable verbose logging")

	rootCmd.MarkFlagRequired("dir")

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Short:  "Display help",
		Hidden: true,
	})
}
