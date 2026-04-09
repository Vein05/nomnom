package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the main configuration structure for the application
type Config struct {
	Output            string                  `json:"output"`             // Output directory for processed files
	Case              string                  `json:"case"`               // Case identifier or name
	AI                AIConfig                `json:"ai"`                 // AI-related settings
	FileHandling      FileHandlingConfig      `json:"file_handling"`      // File processing settings
	ContentExtraction ContentExtractionConfig `json:"content_extraction"` // Content extraction settings
	Performance       PerformanceConfig       `json:"performance"`        // Performance tuning settings
	Logging           LoggingConfig           `json:"logging"`            // Logging configuration
}

// VisionConfig holds settings for AI vision capabilities
type VisionConfig struct {
	Enabled      bool   `json:"enabled"`                  // Whether vision processing is enabled
	MaxImageSize string `json:"max_image_size,omitempty"` // Maximum allowed image size
}

// AIConfig contains settings for AI provider integration
type AIConfig struct {
	Provider    string       `json:"provider"`          // AI service provider name
	Model       string       `json:"model"`             // AI model to use
	APIKey      string       `json:"api_key,omitempty"` // API key for AI service
	Vision      VisionConfig `json:"vision"`            // Vision processing settings
	MaxTokens   int          `json:"max_tokens"`        // Maximum tokens for AI responses
	Temperature float64      `json:"temperature"`       // AI response creativity control
	Prompt      string       `json:"prompt"`            // Default prompt for AI
}

// FileHandlingConfig defines how files are processed
type FileHandlingConfig struct {
	MaxSize     string `json:"max_size"`     // Maximum file size allowed
	AutoApprove bool   `json:"auto_approve"` // Whether to automatically approve files
}

// ContentExtractionConfig specifies content extraction parameters
type ContentExtractionConfig struct {
	ExtractText      bool `json:"extract_text"`       // Enable text extraction
	ExtractMetadata  bool `json:"extract_metadata"`   // Enable metadata extraction
	MaxContentLength int  `json:"max_content_length"` // Maximum content length to process
	SkipLargeFiles   bool `json:"skip_large_files"`   // Skip files exceeding size limits
	ReadContext      bool `json:"read_context"`       // Enable context reading
}

// PerformanceConfig holds performance optimization settings
type PerformanceConfig struct {
	AI   PerformanceAIConfig   `json:"ai,omitempty"`   // AI processing performance settings
	File PerformanceFileConfig `json:"file,omitempty"` // File handling performance settings
}

// PerformanceAIConfig defines AI processing performance parameters
type PerformanceAIConfig struct {
	Workers int    `json:"workers,omitempty"` // Number of AI processing workers
	Timeout string `json:"timeout,omitempty"` // Timeout for AI operations
	Retries int    `json:"retries,omitempty"` // Number of retry attempts for AI operations
}

// PerformanceFileConfig defines file handling performance parameters
type PerformanceFileConfig struct {
	Workers int    `json:"workers"`           // Number of file processing workers
	Timeout string `json:"timeout,omitempty"` // Timeout for file operations
	Retries int    `json:"retries,omitempty"` // Number of retry attempts for file operations
}

// LoggingConfig specifies logging settings
type LoggingConfig struct {
	Enabled bool   `json:"enabled"`  // Whether logging is enabled
	LogPath string `json:"log_path"` // Path for log files
}

// DefaultConfigPath returns the default config path for the current OS.
func DefaultConfigPath() (string, error) {
	return ResolveConfigPath("")
}

// ResolveConfigPath returns the explicit path or the default OS-specific path.
func ResolveConfigPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("failed to locate APPDATA directory")
		}
		return filepath.Join(appData, "nomnom", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "nomnom", "config.json"), nil
}

// LoadConfig loads and parses the configuration file from the specified path.
// If path is empty, it uses the default location for the current OS.
func LoadConfig(path string, step string) (Config, error) {
	resolvedPath, err := ResolveConfigPath(path)
	if err != nil {
		return Config{}, err
	}

	// Print the config file path being loaded
	fmt.Printf(step+"Loading config from: %s\n", resolvedPath)

	// Read the config file
	file, err := os.ReadFile(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found at %s", resolvedPath)
		}
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON config file into Config struct
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
