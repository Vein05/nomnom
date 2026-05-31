package main

type RunJobOptions struct {
	DryRun      bool `json:"dry_run"`
	LogSession  bool `json:"log_session"`
	AutoApprove bool `json:"auto_approve"`
	HotRename   bool `json:"hot_rename"`
	Organize    bool `json:"organize"`
}

type RenameEntry struct {
	Index     int    `json:"index"`
	Original  string `json:"original"`
	NewName   string `json:"new_name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	SizeBytes int64  `json:"size_bytes"`
	Reason    string `json:"reason,omitempty"`
}

type JobSummary struct {
	Planned int `json:"planned"`
	Renamed int `json:"renamed"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

type JobStatus struct {
	JobID       string     `json:"job_id"`
	State       string     `json:"state"`
	Done        int        `json:"done"`
	Total       int        `json:"total"`
	CurrentFile string     `json:"current_file"`
	Message     string     `json:"message"`
	Summary     JobSummary `json:"summary"`
	OutputDir   string     `json:"output_dir"`
}

type Session struct {
	Date      string `json:"date"`
	Directory string `json:"directory"`
	Files     string `json:"files"`
	Model     string `json:"model"`
	Mode      string `json:"mode"`
	Status    string `json:"status"`
}

type AnalyticsSessionPoint struct {
	Date    string `json:"date"`
	Renamed int    `json:"renamed"`
}

type AnalyticsSummary struct {
	Sessions       int                      `json:"sessions"`
	Renamed        int                      `json:"renamed"`
	Tokens         int                      `json:"tokens"`
	AvgPerRun      float64                  `json:"avg_per_run"`
	RecentRuns     int                      `json:"recent_runs"`
	UniqueModels   int                      `json:"unique_models"`
	RecentSessions []AnalyticsSessionPoint  `json:"recent_sessions"`
	HistoryError   string                   `json:"history_error,omitempty"`
}

type OllamaStatus struct {
	Running        bool   `json:"running"`
	ModelAvailable bool   `json:"model_available"`
	Message        string `json:"message"`
}

type AIStatus struct {
	Provider string       `json:"provider"`
	Model    string       `json:"model"`
	Ollama   OllamaStatus `json:"ollama"`
}

type OpenRouterKeyStatus struct {
	Available bool   `json:"available"`
	Source    string `json:"source"`
}

type OpenRouterTestResult struct {
	Ok         bool   `json:"ok"`
	StatusCode int    `json:"status_code"`
	StatusText string `json:"status_text"`
	Source     string `json:"source"`
	Message    string `json:"message"`
	Response   string `json:"response"`
}

type VisionConfig struct {
	Enabled      bool   `json:"enabled"`
	MaxImageSize string `json:"max_image_size"`
}

type AIConfig struct {
	Provider    string       `json:"provider"`
	Model       string       `json:"model"`
	APIKey      string       `json:"api_key,omitempty"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
	Vision      VisionConfig `json:"vision"`
	Prompt      string       `json:"prompt"`
}

type FileHandlingConfig struct {
	MaxSize      string `json:"max_size"`
	AutoApprove  bool   `json:"auto_approve"`
	HotRename    bool   `json:"hot_rename"`
	SkipDotFiles bool   `json:"skip_dot_files"`
}

type ContentExtractionConfig struct {
	ExtractText      bool `json:"extract_text"`
	ExtractMetadata  bool `json:"extract_metadata"`
	MaxContentLength int  `json:"max_content_length"`
	SkipLargeFiles   bool `json:"skip_large_files"`
	ReadContext      bool `json:"read_context"`
}

type PerformancePipelineConfig struct {
	Workers int    `json:"workers"`
	Timeout string `json:"timeout"`
	Retries int    `json:"retries"`
}

type PerformanceConfig struct {
	AI   PerformancePipelineConfig `json:"ai"`
	File PerformancePipelineConfig `json:"file"`
}

type LoggingConfig struct {
	Enabled bool   `json:"enabled"`
	LogPath string `json:"log_path"`
}

type DesktopConfig struct {
	Output            string                  `json:"output"`
	Case              string                  `json:"case"`
	AI                AIConfig                `json:"ai"`
	FileHandling      FileHandlingConfig      `json:"file_handling"`
	ContentExtraction ContentExtractionConfig `json:"content_extraction"`
	Performance       PerformanceConfig       `json:"performance"`
	Logging           LoggingConfig           `json:"logging"`
}
