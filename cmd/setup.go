package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"nomnom/internal/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var setupConfigPath string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create or update the NomNom config interactively",
	Example: `nomnom setup
nomnom setup -c ~/.config/nomnom/config.json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := newCLIPresenter()
		presenter.Banner()
		presenter.Divider()

		resolvedPath, err := utils.ResolveConfigPath(setupConfigPath)
		if err != nil {
			return err
		}

		config := utils.DefaultConfig()
		if _, err := os.Stat(resolvedPath); err == nil {
			existing, loadErr := utils.LoadConfig(resolvedPath, "")
			if loadErr != nil {
				presenter.Warnf("Existing config found at %s but could not be loaded: %v", resolvedPath, loadErr)
			} else {
				config = mergeConfig(config, existing)
				overwrite, askErr := promptBool("An existing config was found. Update it?", true)
				if askErr != nil {
					return askErr
				}
				if !overwrite {
					presenter.Warnf("Setup cancelled. Existing config left unchanged.")
					return nil
				}
			}
		}

		presenter.Titlef("Core Configuration")

		provider, err := promptSelect("AI provider", []string{"openrouter", "deepseek", "ollama"}, config.AI.Provider)
		if err != nil {
			return err
		}
		config.AI.Provider = provider

		defaultModel := modelDefaultForProvider(provider)
		if config.AI.Model == "" || providerChangedModel(provider, config.AI.Model) {
			config.AI.Model = defaultModel
		}
		model, err := promptText("Model", config.AI.Model, nonEmptyValidator("model"))
		if err != nil {
			return err
		}
		config.AI.Model = model

		if provider == "ollama" {
			config.AI.APIKey = ""
		} else {
			apiKey, err := promptAPIKey(config.AI.APIKey)
			if err != nil {
				return err
			}
			config.AI.APIKey = apiKey
		}

		visionEnabled, err := promptBool("Enable vision for supported files?", config.AI.Vision.Enabled)
		if err != nil {
			return err
		}
		config.AI.Vision.Enabled = visionEnabled

		caseStyle, err := promptSelect("Filename case", []string{"snake", "camel", "kebab", "pascal"}, config.Case)
		if err != nil {
			return err
		}
		config.Case = caseStyle

		loggingEnabled, err := promptBool("Enable logging?", config.Logging.Enabled)
		if err != nil {
			return err
		}
		config.Logging.Enabled = loggingEnabled

		advanced, err := promptBool("Configure advanced settings now?", false)
		if err != nil {
			return err
		}
		if advanced {
			if err := runAdvancedSetup(&config); err != nil {
				return err
			}
		}

		presenter.Divider()
		presenter.Titlef("Config Summary")
		presenter.Infof("Path: %s", resolvedPath)
		presenter.Infof("Provider: %s", config.AI.Provider)
		presenter.Infof("Model: %s", config.AI.Model)
		presenter.Infof("Vision enabled: %t", config.AI.Vision.Enabled)
		presenter.Infof("Case: %s", config.Case)
		presenter.Infof("Logging enabled: %t", config.Logging.Enabled)
		if config.Output == "" {
			presenter.Infof("Output directory: default (<input>/nomnom/renamed)")
		} else {
			presenter.Infof("Output directory: %s", config.Output)
		}

		save, err := promptBool("Save this config?", true)
		if err != nil {
			return err
		}
		if !save {
			presenter.Warnf("Setup cancelled. No config was written.")
			return nil
		}

		savedPath, err := utils.SaveConfig(resolvedPath, config)
		if err != nil {
			return err
		}

		presenter.Divider()
		presenter.Titlef("Setup Complete")
		presenter.Infof("Config saved to %s", savedPath)
		presenter.Infof("You can start with: nomnom -d ~/path/to/files")
		presenter.Infof("Run setup again any time to adjust core settings.")
		return nil
	},
}

func init() {
	setupCmd.Flags().StringVarP(&setupConfigPath, "config", "c", "", "Path to config file")
	rootCmd.AddCommand(setupCmd)
}

func runAdvancedSetup(config *utils.Config) error {
	output, err := promptOptionalText("Output directory (leave blank to use the default)", config.Output)
	if err != nil {
		return err
	}
	config.Output = output

	maxTokens, err := promptInt("Max output tokens per rename response", config.AI.MaxTokens)
	if err != nil {
		return err
	}
	config.AI.MaxTokens = maxTokens

	temperature, err := promptFloat("Temperature (0 = deterministic, 2 = creative)", config.AI.Temperature)
	if err != nil {
		return err
	}
	config.AI.Temperature = temperature

	maxSize, err := promptText("Max file size", config.FileHandling.MaxSize, nonEmptyValidator("max size"))
	if err != nil {
		return err
	}
	config.FileHandling.MaxSize = maxSize

	aiWorkers, err := promptInt("AI workers", config.Performance.AI.Workers)
	if err != nil {
		return err
	}
	config.Performance.AI.Workers = aiWorkers

	fileWorkers, err := promptInt("File workers", config.Performance.File.Workers)
	if err != nil {
		return err
	}
	config.Performance.File.Workers = fileWorkers

	autoApprove, err := promptBool("Auto-approve renames?", config.FileHandling.AutoApprove)
	if err != nil {
		return err
	}
	config.FileHandling.AutoApprove = autoApprove

	moveFiles, err := promptBool("Move files instead of copy mode?", config.FileHandling.MoveFiles)
	if err != nil {
		return err
	}
	config.FileHandling.MoveFiles = moveFiles

	customPrompt, err := promptOptionalText("Default custom prompt (leave blank to use NomNom default)", config.AI.Prompt)
	if err != nil {
		return err
	}
	config.AI.Prompt = customPrompt

	return nil
}

func mergeConfig(base, override utils.Config) utils.Config {
	if override.Output != "" {
		base.Output = override.Output
	}
	if override.Case != "" {
		base.Case = override.Case
	}
	if override.AI.Provider != "" {
		base.AI.Provider = override.AI.Provider
	}
	if override.AI.Model != "" {
		base.AI.Model = override.AI.Model
	}
	if override.AI.APIKey != "" {
		base.AI.APIKey = override.AI.APIKey
	}
	base.AI.Vision.Enabled = override.AI.Vision.Enabled
	if override.AI.Vision.MaxImageSize != "" {
		base.AI.Vision.MaxImageSize = override.AI.Vision.MaxImageSize
	}
	base.AI.MaxTokens = override.AI.MaxTokens
	base.AI.Temperature = override.AI.Temperature
	if override.AI.Prompt != "" {
		base.AI.Prompt = override.AI.Prompt
	}
	if override.FileHandling.MaxSize != "" {
		base.FileHandling.MaxSize = override.FileHandling.MaxSize
	}
	base.FileHandling.AutoApprove = override.FileHandling.AutoApprove
	base.FileHandling.MoveFiles = override.FileHandling.MoveFiles
	if override.ContentExtraction != (utils.ContentExtractionConfig{}) {
		base.ContentExtraction = override.ContentExtraction
	}
	if override.Performance.AI.Workers != 0 {
		base.Performance.AI.Workers = override.Performance.AI.Workers
	}
	if override.Performance.AI.Timeout != "" {
		base.Performance.AI.Timeout = override.Performance.AI.Timeout
	}
	if override.Performance.AI.Retries != 0 {
		base.Performance.AI.Retries = override.Performance.AI.Retries
	}
	if override.Performance.File.Workers != 0 {
		base.Performance.File.Workers = override.Performance.File.Workers
	}
	if override.Performance.File.Timeout != "" {
		base.Performance.File.Timeout = override.Performance.File.Timeout
	}
	if override.Performance.File.Retries != 0 {
		base.Performance.File.Retries = override.Performance.File.Retries
	}
	base.Logging.Enabled = override.Logging.Enabled
	if override.Logging.LogPath != "" {
		base.Logging.LogPath = override.Logging.LogPath
	}

	return base
}

func modelDefaultForProvider(provider string) string {
	switch provider {
	case "deepseek":
		return "deepseek-chat"
	case "ollama":
		return "llama3.2"
	default:
		return "google/gemini-2.0-flash-001"
	}
}

func providerChangedModel(provider, model string) bool {
	if model == "" {
		return true
	}

	switch provider {
	case "deepseek":
		return strings.Contains(model, "/") || strings.Contains(strings.ToLower(model), "llama")
	case "ollama":
		return strings.Contains(model, "/") || strings.Contains(strings.ToLower(model), "deepseek")
	default:
		return strings.HasPrefix(model, "deepseek") || strings.Contains(strings.ToLower(model), "llama")
	}
}

func promptSelect(label string, options []string, current string) (string, error) {
	selectedIndex := 0
	for index, option := range options {
		if option == current {
			selectedIndex = index
			break
		}
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     options,
		Size:      len(options),
		CursorPos: selectedIndex,
	}

	_, value, err := prompt.Run()
	return value, err
}

func promptBool(label string, current bool) (bool, error) {
	defaultValue := "yes"
	if !current {
		defaultValue = "no"
	}

	value, err := promptSelect(label, []string{"yes", "no"}, defaultValue)
	if err != nil {
		return false, err
	}

	return value == "yes", nil
}

func promptText(label, defaultValue string, validate func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	value, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func promptOptionalText(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	value, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func promptAPIKey(existing string) (string, error) {
	label := "API key"
	validate := nonEmptyValidator("api key")
	if existing != "" {
		label = "API key (leave blank to keep existing value)"
		validate = func(input string) error { return nil }
	}

	prompt := promptui.Prompt{
		Label:    label,
		Mask:     '*',
		Validate: validate,
	}

	value, err := prompt.Run()
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" && existing != "" {
		return existing, nil
	}

	return value, nil
}

func promptInt(label string, current int) (int, error) {
	value, err := promptText(label, strconv.Itoa(current), func(input string) error {
		number, parseErr := strconv.Atoi(strings.TrimSpace(input))
		if parseErr != nil || number <= 0 {
			return errors.New("enter a positive integer")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(value)
}

func promptFloat(label string, current float64) (float64, error) {
	value, err := promptText(label, strconv.FormatFloat(current, 'f', -1, 64), func(input string) error {
		number, parseErr := strconv.ParseFloat(strings.TrimSpace(input), 64)
		if parseErr != nil {
			return errors.New("enter a valid number")
		}
		if number < 0 || number > 2 {
			return errors.New("enter a value between 0 and 2")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(value, 64)
}

func nonEmptyValidator(field string) func(string) error {
	return func(input string) error {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}
