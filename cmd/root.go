package cmd

import (
	"os"
	"strconv"

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
		a := cmd.Flags()

		dir, _ := a.GetString("dir")
		configPath, _ := a.GetString("config")
		autoApprove, _ := a.GetBool("auto-approve")
		dryRun, _ := a.GetBool("dry-run")
		verbose, _ := a.GetBool("verbose")

		_, err := os.Stdout.WriteString("NomNom is running...\n")
		if err != nil {
			os.Exit(1)
		}
		_, err = os.Stdout.WriteString("The source directory is: " + dir + "\n")
		if err != nil {
			os.Exit(1)
		}
		_, err = os.Stdout.WriteString("The config path is: " + configPath + "\n")
		if err != nil {
			os.Exit(1)
		}
		_, err = os.Stdout.WriteString("Auto approve is: " + strconv.FormatBool(autoApprove) + "\n")
		if err != nil {
			os.Exit(1)
		}
		_, err = os.Stdout.WriteString("Dry run is: " + strconv.FormatBool(dryRun) + "\n")
		if err != nil {
			os.Exit(1)
		}
		_, err = os.Stdout.WriteString("Verbose is: " + strconv.FormatBool(verbose) + "\n")
		if err != nil {
			os.Exit(1)
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
