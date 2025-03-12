package nomnom

import (
	"encoding/json"
	"log"
	"os"
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
	APIKey      string  `json:"api_key"`
	OllamaModel string  `json:"ollama_model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type FileHandlingConfig struct {
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	MaxSize       string   `json:"max_size"`
	SkipErrors    bool     `json:"skip_errors"`
	KeepOriginals bool     `json:"keep_originals"`
	Backup        bool     `json:"backup"`
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
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`
	NoColor  bool   `json:"no_color"`
}

func LoadConfig(path string) Config {
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
