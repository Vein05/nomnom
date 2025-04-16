package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func Init() {

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
