package nomnom

import (
	"encoding/json"
	"os"
	"runtime"

	log "github.com/charmbracelet/log"
)

type Config struct {
	Output            string                  `json:"output"`
	Case              string                  `json:"case"`
	AI                AIConfig                `json:"ai"`
	FileHandling      FileHandlingConfig      `json:"file_handling"`
	ContentExtraction ContentExtractionConfig `json:"content_extraction"`
	Performance       PerformanceConfig       `json:"performance"`
	Logging           LoggingConfig           `json:"logging"`
}

type AIConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key,omitempty"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Prompt      string  `json:"prompt"`
}

type FileHandlingConfig struct {
	MaxSize     string `json:"max_size"`
	AutoApprove bool   `json:"auto_approve"`
}

type ContentExtractionConfig struct {
	ExtractText      bool `json:"extract_text"`
	ExtractMetadata  bool `json:"extract_metadata"`
	MaxContentLength int  `json:"max_content_length"`
	SkipLargeFiles   bool `json:"skip_large_files"`
	ReadContext      bool `json:"read_context"`
}

type PerformanceConfig struct {
	Workers int    `json:"workers"`
	Timeout string `json:"timeout"`
	Retries int    `json:"retries"`
}

type LoggingConfig struct {
	Enabled bool   `json:"enabled"`
	LogPath string `json:"log_path"`
}

func LoadConfig(path string) Config {
	// check if path is empty
	if path == "" {
		path = "./config.json"

		// check the os type and set the path accordingly
		if runtime.GOOS == "windows" {
			// windows uses backslashes for paths and we set out config file in
			path = "./config.json"
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("Failed to get user home directory: %v", err)
			}
			path = home + "/.config/nomnom/config.json"
		}
	}

	log.Info("Loading: ", "config", path)

	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	return config
}
