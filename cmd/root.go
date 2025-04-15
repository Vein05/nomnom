package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ai "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
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

func printSummary(results []contentprocessors.ProcessResult) {
	success := color.New(color.FgGreen).SprintFunc()
	failed := color.New(color.FgRed).SprintFunc()
	info := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s\n", info("📊 Summary of Operations"))

	fmt.Printf("%s\n", info("══════════════════════"))

	for _, result := range results {
		if result.Success {
			fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ → \033]8;;file://%s\033\\%s\033]8;;\033\\\n",
				success("✓"),
				result.OriginalPath,
				filepath.Base(result.OriginalPath),
				result.FullNewPath,
				filepath.Base(result.NewPath))
		} else {
			fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ (Error: %v)\n",
				failed("✗"),
				result.OriginalPath,
				filepath.Base(result.OriginalPath),
				result.Error)
		}
	}
}

func generateSummary(results []contentprocessors.ProcessResult) (successCount int) {
	info := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s\n", info("📊 Processing Results"))

	fmt.Printf("%s\n", info("══════════════════════"))
	successCount = 0
	for _, result := range results {
		if result.Success {
			successCount++
			if cmdArgs.dryRun {
				fmt.Printf("%s \033]8;;file://%s\033\\%s\033]8;;\033\\ → %s\n",
					color.GreenString("🔍 Would rename:"),
					result.FullOriginalPath,
					filepath.Base(result.OriginalPath),
					filepath.Base(result.NewPath))
			} else {
				fmt.Printf("%s %s →\033]8;;file://%s\033\\%s\033]8;;\033\\ \n",
					color.GreenString("✅ Renamed:"),
					filepath.Base(result.OriginalPath),
					result.FullNewPath,
					filepath.Base(result.NewPath))
			}
		} else {
			fmt.Printf("%s %s (Error: %v)\n",
				color.RedString("❌ Failed to process:"),
				filepath.Base(result.OriginalPath),
				result.Error)
		}
	}
	return successCount
}

var rootCmd = &cobra.Command{
	Use:   "nomnom",
	Short: "A Go CLI tool to bulk rename and organize files using AI.",
	Long:  `NomNom is a command-line tool that renames files in a folder based on their content using AI models.`,
	Run: func(cmd *cobra.Command, _ []string) {
		info := color.New(color.FgCyan).SprintFunc()
		title := color.New(color.FgBlue).SprintFunc()

		fmt.Println(color.BlueString(`
			_  _                 _  _                 
			| \| | ___  _ __    | \| | ___  _ __ ___  
			| .  |/ _ \| '_ \   | .  |/ _ \| '_ ' _ \ 
			| |\  | (_) | | | |  | |\  | (_) | | | | | |
			|_| \_|\___/|_| |_|  |_| \_|\___/|_| |_| |_|
			`))
		message := "Welcome to NomNom"
		fmt.Printf("%s ", color.WhiteString("▶"))
		for _, char := range message {
			fmt.Printf("%s", color.BlueString(string(char)))
			time.Sleep(50 * time.Millisecond)
		}
		fmt.Println()

		githubLink := "https://github.com/vein05/nomnom"
		fmt.Printf("%s %s\n", color.WhiteString("▶"), color.BlueString(githubLink))

		fmt.Printf("%s\n", info("══════════════════════"))
		time.Sleep(2 * time.Second)

		// Check if revert flag is set
		if cmdArgs.revert != "" {
			opts := files.RevertOptions{
				ChangeLogPath: cmdArgs.revert,
				EnableLogging: cmdArgs.log,
				AutoApprove:   cmdArgs.autoApprove,
			}

			if err := files.ProcessRevert(opts); err != nil {
				color.Red("Error processing revert: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Load configuration
		config := utils.LoadConfig(cmdArgs.configPath, "")
		fmt.Printf("%s\n", info("══════════════════════"))
		// Create a new query
		query, err := contentprocessors.NewQuery(
			cmdArgs.prompt,
			cmdArgs.dir,
			cmdArgs.configPath,
			config,
			cmdArgs.autoApprove,
			cmdArgs.dryRun,
			cmdArgs.log,
			cmdArgs.organize,
		)
		if err != nil {
			color.Red("Error creating query: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", info("══════════════════════"))

		// Set up output directory
		outputDir := config.Output
		if outputDir == "" {
			outputDir = filepath.Join(cmdArgs.dir, "nomnom", "renamed")
		}

		output_text := fmt.Sprintf("Output directory set up at: %s", outputDir)
		if cmdArgs.dryRun {
			output_text = fmt.Sprintf("Output directory would be set up at: %s", outputDir)
		}
		fmt.Printf("%s\n", title(output_text))

		fmt.Printf("%s\n", info("══════════════════════"))

		fmt.Printf("%s\n", title("Processing files with AI to generate new names"))

		// Process files with AI to get new names
		aiResult, err := ai.HandleAI(config, *query)
		if err != nil {
			color.Red("Error processing files with AI: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", info("══════════════════════"))

		// Update query with AI results
		query.Folders = aiResult.Folders

		// Create and run the safe processor
		fmt.Printf("%s\n", title("Processing file renames"))

		fmt.Printf("%s\n", info("══════════════════════"))

		processor := contentprocessors.NewSafeProcessor(query, outputDir)
		results, err := processor.Process()
		if err != nil {
			color.Red("Error processing files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", info("══════════════════════"))
		fmt.Printf("%s\n", title("Processing files with AI to generate new names\n"))
		fmt.Printf("%s\n", info("══════════════════════"))

		successCount := generateSummary(results)

		fmt.Printf("%s\n", info("══════════════════════"))

		if cmdArgs.dryRun {
			color.Green("\n%s %d files would be renamed successfully.\n", ("✅"), successCount)
			color.Yellow("\nTo apply these changes, run: nomnom -d \"%s\" --dry-run=false\n", cmdArgs.dir)
		} else {
			printSummary(results)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	helpTemplate := color.BlueString(`
			_  _                 _  _                 
			| \| | ___  _ __    | \| | ___  _ __ ___  
			| .  |/ _ \| '_ \   | .  |/ _ \| '_ ' _ \ 
			| |\  | (_) | | | |  | |\  | (_) | | | | | |
			|_| \_|\___/|_| |_|  |_| \_|\___/|_| |_| |_|
			
		`) + `
{{.Short}}

Usage:
{{.Use}} [flags]

Flags:
{{.LocalFlags.FlagUsages}}

Examples:
{{.Name}} -d ~/Documents/files                    # Preview rename operations
{{.Name}} -d ~/Documents/files -n=false          # Execute rename operations
{{.Name}} -d ~/Documents/files -p research       # Use research prompt
{{.Name}} -r .nomnom/logs/changes_123.json      # Revert changes` + "\n"

	rootCmd.Flags().StringVarP(&cmdArgs.dir, "dir", "d", "",
		color.CyanString("Source directory containing files to rename"))

	rootCmd.Flags().StringVarP(&cmdArgs.configPath, "config", "c", "",
		color.CyanString("Path to config file (default: ~/.config/nomnom/config.json)"))

	rootCmd.Flags().BoolVarP(&cmdArgs.autoApprove, "auto-approve", "y", false,
		color.CyanString("Automatically approve changes without confirmation"))

	rootCmd.Flags().BoolVarP(&cmdArgs.dryRun, "dry-run", "n", true,
		color.CyanString("Preview changes without renaming files"))

	rootCmd.Flags().BoolVarP(&cmdArgs.log, "log", "l", true,
		color.CyanString("Enable logging to file"))

	rootCmd.Flags().StringVarP(&cmdArgs.revert, "revert", "r", "",
		color.CyanString("Path to log file for reverting operations"))

	rootCmd.Flags().BoolVarP(&cmdArgs.organize, "organize", "o", true,
		color.CyanString("Organize files into folders based on content"))

	rootCmd.Flags().StringVarP(&cmdArgs.prompt, "prompt", "p", "",
		color.CyanString("Custom AI prompt (use 'research' or 'images' for built-in prompts)"))

	rootCmd.SetHelpTemplate(helpTemplate)

	rootCmd.SetErrPrefix(color.RedString("Error: "))
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s%s\n\n", rootCmd.ErrPrefix(), err)
		_ = cmd.Help()
		os.Exit(1)
		return nil
	})

	// Add PreRunE validation
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmdArgs.revert == "" && cmdArgs.dir == "" {
			return fmt.Errorf("--dir flag is required when not using --revert")
		}
		return nil
	}

	// Override Run to handle errors and show help
	originalRun := rootCmd.Run
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if originalRun != nil {
			originalRun(cmd, args)
		}
		return nil
	}

	// Custom help command
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Short:  "Display help",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			helpText := strings.SplitSeq(rootCmd.HelpTemplate(), "\n")
			for line := range helpText {
				fmt.Println(line)
			}
		},
	})
}
