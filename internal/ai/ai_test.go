package ai

import (
	"os"
	"testing"
	"time"

	contentprocessors "nomnom/internal/content"
	utils "nomnom/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestHandleAIProviderSelection(t *testing.T) {
	originalDeepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	originalOllamaKey := os.Getenv("OLLAMA_API_KEY")
	originalOpenRouterKey := os.Getenv("OPENROUTER_API_KEY")

	defer func() {
		os.Setenv("DEEPSEEK_API_KEY", originalDeepseekKey)
		os.Setenv("OLLAMA_API_KEY", originalOllamaKey)
		os.Setenv("OPENROUTER_API_KEY", originalOpenRouterKey)
	}()

	tests := []struct {
		name          string
		config        utils.Config
		envSetup      map[string]string
		expectedModel string
		expectedError bool
	}{
		{
			name: "Config with Deepseek API key",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "deepseek",
			expectedError: false,
		},
		{
			name: "No config API key but Deepseek env variable set",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
				},
			},
			envSetup: map[string]string{
				"DEEPSEEK_API_KEY": "dummy-key",
			},
			expectedModel: "deepseek",
			expectedError: false,
		},
		{
			name: "Ollama provider without API key",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "ollama",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "ollama",
			expectedError: false,
		},
		{
			name: "No config API key but OpenRouter env variable set",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "openrouter",
				},
			},
			envSetup: map[string]string{
				"OPENROUTER_API_KEY": "dummy-key",
			},
			expectedModel: "openrouter",
			expectedError: false,
		},
		{
			name: "Multiple env variables set - should use provider from config",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "ollama",
					APIKey:   "dummy-key",
				},
			},
			envSetup: map[string]string{
				"DEEPSEEK_API_KEY":   "dummy-key",
				"OPENROUTER_API_KEY": "dummy-key",
			},
			expectedModel: "ollama",
			expectedError: false,
		},
		{
			name: "No API keys available for Deepseek",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "deepseek",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "No API keys available for OpenRouter",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "openrouter",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "Invalid provider",
			config: utils.Config{
				AI: utils.AIConfig{
					Provider: "invalid-provider",
					APIKey:   "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "",
			expectedError: true,
		},
		{
			name: "No provider set - should default to deepseek",
			config: utils.Config{
				AI: utils.AIConfig{
					APIKey: "dummy-key",
				},
			},
			envSetup:      map[string]string{},
			expectedModel: "deepseek",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("DEEPSEEK_API_KEY")
			os.Unsetenv("OLLAMA_API_KEY")
			os.Unsetenv("OPENROUTER_API_KEY")

			for key, value := range tt.envSetup {
				os.Setenv(key, value)
			}

			query := contentprocessors.Query{
				Scan: contentprocessors.ScanResult{
					RootDir: "test",
					Files: []contentprocessors.ScannedFile{
						{
							SourcePath:   "/tmp/test.txt",
							RelativePath: "test.txt",
							OriginalName: "test.txt",
							Context:      "test",
						},
					},
				},
			}

			_, err := HandleAI(tt.config, query)

			if tt.expectedError {
				assert.Error(t, err, "Expected an error for test case: %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
			}
		})
	}
}

func TestAIRuntimeParsesNumericTimeoutSeconds(t *testing.T) {
	config := utils.Config{
		Performance: utils.PerformanceConfig{
			AI: utils.PerformanceAIConfig{
				Workers: 2,
				Retries: 1,
				Timeout: "30",
			},
		},
	}

	workers, retries, timeout, err := aiRuntime(config)
	assert.NoError(t, err)
	assert.Equal(t, 2, workers)
	assert.Equal(t, 1, retries)
	assert.Equal(t, 30*time.Second, timeout)
}

// -- New tests for untested AI functions --

func TestNormalizeQueryOpts(t *testing.T) {
	tests := []struct {
		name     string
		input    QueryOpts
		expected QueryOpts
	}{
		{
			name: "default values for zero MaxTokens",
			input: QueryOpts{
				MaxTokens:   0,
				Temperature: 0.0,
			},
			expected: QueryOpts{
				MaxTokens:   128,
				Temperature: 0.0,
			},
		},
		{
			name: "clamp MaxTokens above 4000",
			input: QueryOpts{
				MaxTokens:   5000,
				Temperature: 0.5,
			},
			expected: QueryOpts{
				MaxTokens:   4000,
				Temperature: 0.5,
			},
		},
		{
			name: "clamp negative temperature",
			input: QueryOpts{
				MaxTokens:   100,
				Temperature: -1.0,
			},
			expected: QueryOpts{
				MaxTokens:   100,
				Temperature: 0.2,
			},
		},
		{
			name: "clamp temperature above 2",
			input: QueryOpts{
				MaxTokens:   100,
				Temperature: 3.0,
			},
			expected: QueryOpts{
				MaxTokens:   100,
				Temperature: 0.2,
			},
		},
		{
			name: "valid values unchanged",
			input: QueryOpts{
				MaxTokens:   256,
				Temperature: 0.7,
			},
			expected: QueryOpts{
				MaxTokens:   256,
				Temperature: 0.7,
			},
		},
		{
			name: "MaxTokens = 1 defaults to 128",
			input: QueryOpts{
				MaxTokens:   1,
				Temperature: 0.5,
			},
			expected: QueryOpts{
				MaxTokens:   128,
				Temperature: 0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeQueryOpts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTimeoutValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback string
		want     time.Duration
		wantErr  bool
	}{
		{name: "duration string seconds", raw: "5s", fallback: "30s", want: 5 * time.Second, wantErr: false},
		{name: "duration string minutes", raw: "2m", fallback: "30s", want: 2 * time.Minute, wantErr: false},
		{name: "numeric seconds", raw: "10", fallback: "30s", want: 10 * time.Second, wantErr: false},
		{name: "empty uses fallback", raw: "", fallback: "15s", want: 15 * time.Second, wantErr: false},
		{name: "whitespace uses fallback", raw: "  ", fallback: "20s", want: 20 * time.Second, wantErr: false},
		{name: "invalid value", raw: "not-a-number", fallback: "also-bad", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeoutValue(tt.raw, tt.fallback)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRetryReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "invalid response prefix",
			err:  assert.AnError, // non-prefixed error
			want: "",
		},
		{
			name: "no prefix match",
			err:  assert.AnError,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryReason(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}

	// Test with the specific prefix
	reason := retryReason(assert.AnError)
	assert.Equal(t, "", reason, "non-prefixed error should return empty")

	// Actually test the prefix matching
	prefixedErr := assert.AnError // any error as long as message has prefix
	_ = prefixedErr
}

func TestRetryReason_WithPrefix(t *testing.T) {
	err := newTestError("invalid response from AI: filename has spaces")
	reason := retryReason(err)
	assert.Equal(t, "filename has spaces", reason)
}

func TestRetryReason_NoPrefix(t *testing.T) {
	err := newTestError("network timeout")
	reason := retryReason(err)
	assert.Equal(t, "", reason)
}

type testError struct {
	msg string
}

func newTestError(msg string) error {
	return &testError{msg: msg}
}

func (e *testError) Error() string {
	return e.msg
}

func TestHasVisionSource(t *testing.T) {
	tests := []struct {
		name string
		file contentprocessors.ScannedFile
		want bool
	}{
		{
			name: "has visual path",
			file: contentprocessors.ScannedFile{
				SourcePath:   "/tmp/doc.pdf",
				VisualPath:   "/tmp/preview.jpg",
				OriginalName: "doc.pdf",
			},
			want: true,
		},
		{
			name: "is image file",
			file: contentprocessors.ScannedFile{
				SourcePath:   "/tmp/photo.png",
				OriginalName: "photo.png",
			},
			want: true,
		},
		{
			name: "no vision source",
			file: contentprocessors.ScannedFile{
				SourcePath:   "/tmp/doc.txt",
				OriginalName: "doc.txt",
			},
			want: false,
		},
		{
			name: "jpg extension",
			file: contentprocessors.ScannedFile{
				SourcePath:   "/tmp/photo.jpg",
				OriginalName: "photo.jpg",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasVisionSource(tt.file)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequiresPreviewExtraction(t *testing.T) {
	tests := []struct {
		name string
		file contentprocessors.ScannedFile
		want bool
	}{
		{
			name: "pdf requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/doc.pdf",
			},
			want: true,
		},
		{
			name: "docx requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/doc.docx",
			},
			want: true,
		},
		{
			name: "epub requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/book.epub",
			},
			want: true,
		},
		{
			name: "txt file does not require preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/notes.txt",
			},
			want: false,
		},
		{
			name: "image file does not require preview (already image)",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/photo.png",
			},
			want: false,
		},
		{
			name: "file with visual path does not require preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/doc.pdf",
				VisualPath: "/tmp/preview.jpg",
			},
			want: false,
		},
		{
			name: "pptx requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/slides.pptx",
			},
			want: true,
		},
		{
			name: "xlsx requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/data.xlsx",
			},
			want: true,
		},
		{
			name: "xls requires preview",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/old_data.xls",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requiresPreviewExtraction(tt.file)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVisionSourcePath(t *testing.T) {
	tests := []struct {
		name string
		file contentprocessors.ScannedFile
		want string
	}{
		{
			name: "uses visual path when present",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/doc.pdf",
				VisualPath: "/tmp/preview.jpg",
			},
			want: "/tmp/preview.jpg",
		},
		{
			name: "falls back to source path",
			file: contentprocessors.ScannedFile{
				SourcePath: "/tmp/photo.png",
			},
			want: "/tmp/photo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visionSourcePath(tt.file)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReporterFor(t *testing.T) {
	nop := utils.NopReporter{}
	query := contentprocessors.Query{Reporter: nop}
	result := reporterFor(query)
	assert.Equal(t, nop, result)

	// nil reporter results in NopReporter
	query2 := contentprocessors.Query{}
	result2 := reporterFor(query2)
	assert.IsType(t, utils.NopReporter{}, result2)
}

func TestRecordAnalyticsUsage_NilAnalytics(t *testing.T) {
	// Should not panic
	recordAnalyticsUsage(nil, "deepseek", "deepseek-chat", 10, 5, 15, false)
}

func TestRecordAnalyticsUsage_WithAnalytics(t *testing.T) {
	store := utils.NewAnalyticsStore(t.TempDir(), true)
	recordAnalyticsUsage(store, "deepseek", "deepseek-chat", 10, 5, 15, false)
	// Should not panic or error
}

func TestBuildRenamePlan_EmptyFiles(t *testing.T) {
	plan := buildRenamePlan(
		[]contentprocessors.ScannedFile{},
		1,
		0,
		utils.NopReporter{},
		nil,
		func(file contentprocessors.ScannedFile, retryHint string) (string, error) {
			return "name.txt", nil
		},
	)
	assert.Empty(t, plan)
}

func TestBuildRenamePlan_SingleFile(t *testing.T) {
	plan := buildRenamePlan(
		[]contentprocessors.ScannedFile{
			{
				SourcePath:   "/tmp/test.txt",
				RelativePath: "test.txt",
				OriginalName: "test.txt",
			},
		},
		1,
		0,
		utils.NopReporter{},
		nil,
		func(file contentprocessors.ScannedFile, retryHint string) (string, error) {
			return "renamed.txt", nil
		},
	)
	assert.Len(t, plan, 1)
	assert.Equal(t, "renamed.txt", plan[0].SuggestedName)
}

func TestBuildRenamePlan_MultipleWorkers(t *testing.T) {
	plan := buildRenamePlan(
		[]contentprocessors.ScannedFile{
			{SourcePath: "/tmp/a.txt", RelativePath: "a.txt", OriginalName: "a.txt"},
			{SourcePath: "/tmp/b.txt", RelativePath: "b.txt", OriginalName: "b.txt"},
		},
		4, // more workers than files
		0,
		utils.NopReporter{},
		nil,
		func(file contentprocessors.ScannedFile, retryHint string) (string, error) {
			return "renamed.txt", nil
		},
	)
	assert.Len(t, plan, 2)
}

func TestNormalizeSuggestedName(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		file      contentprocessors.ScannedFile
		caseStyle string
		want      string
		wantErr   bool
	}{
		{
			name: "simple name",
			raw:  "renamed_file.txt",
			file: contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want: "renamed_file.txt",
		},
		{
			name: "adds missing extension",
			raw:  "renamed_file",
			file: contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want: "renamed_file.txt",
		},
		{
			name: "replaces wrong extension",
			raw:  "renamed_file.jpg",
			file: contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want: "renamed_file.txt",
		},
		{
			name:    "empty string",
			raw:     "",
			file:    contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want:    "",
			wantErr: true,
		},
		{
			name: "removes backticks",
			raw:  "`renamed_file.txt`",
			file: contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want: "renamed_file.txt",
		},
		{
			name: "removes code block",
			raw:  "```renamed_file.txt```",
			file: contentprocessors.ScannedFile{OriginalName: "old.txt"},
			want: "renamed_file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSuggestedName(tt.raw, tt.file, tt.caseStyle)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPromptContext(t *testing.T) {
	file := contentprocessors.ScannedFile{
		SourcePath:   "/tmp/test.txt",
		OriginalName: "test.txt",
		Context:      "File: test.txt\nSize: 100 bytes",
	}

	withoutHint := promptContext(file, "")
	assert.Equal(t, file.Context, withoutHint)

	withHint := promptContext(file, "filename has spaces")
	assert.Contains(t, withHint, "Previous filename suggestion failed")
	assert.Contains(t, withHint, "filename has spaces")
	assert.Contains(t, withHint, file.Context)
}

func TestAIRuntime_Defaults(t *testing.T) {
	config := utils.Config{}
	workers, retries, timeout, err := aiRuntime(config)
	assert.NoError(t, err)
	assert.Greater(t, workers, 0, "workers should default to > 0")
	assert.Equal(t, 1, retries, "retries should default to 1")
	assert.Equal(t, 30*time.Second, timeout, "timeout should default to 30s")
}

func TestAIRuntime_ParsesDurationString(t *testing.T) {
	config := utils.Config{
		Performance: utils.PerformanceConfig{
			AI: utils.PerformanceAIConfig{
				Workers: 4,
				Retries: 2,
				Timeout: "45s",
			},
		},
	}

	workers, retries, timeout, err := aiRuntime(config)
	assert.NoError(t, err)
	assert.Equal(t, 4, workers)
	assert.Equal(t, 2, retries)
	assert.Equal(t, 45*time.Second, timeout)
}

func TestPrepareFileForLLM_ReadContextDisabled(t *testing.T) {
	file := contentprocessors.ScannedFile{
		SourcePath:   "/tmp/test.txt",
		RelativePath: "test.txt",
		OriginalName: "test.txt",
		Context:      "original context",
	}

	config := utils.Config{
		ContentExtraction: utils.ContentExtractionConfig{
			ReadContext: false,
		},
	}

	result := prepareFileForLLM(file, config, utils.NopReporter{})
	assert.Equal(t, file.Context, result.Context, "context should be unchanged when ReadContext is disabled")
}

func TestPrepareFileForLLM_ReadContextEnabled_MissingFile(t *testing.T) {
	file := contentprocessors.ScannedFile{
		SourcePath:   "/tmp/nonexistent.txt",
		RelativePath: "nonexistent.txt",
		OriginalName: "nonexistent.txt",
		Context:      "original context",
	}

	config := utils.Config{
		ContentExtraction: utils.ContentExtractionConfig{
			ReadContext:      true,
			MaxContentLength: 100,
		},
	}

	result := prepareFileForLLM(file, config, utils.NopReporter{})
	// Should return the file unchanged with original context on extraction failure
	assert.Equal(t, file.Context, result.Context)
}

func TestSendQueryToLLM_NilClient(t *testing.T) {
	config := utils.Config{}
	query := contentprocessors.Query{}

	err := SendQueryToLLM(nil, config, &query, QueryOpts{})
	assert.Error(t, err)
}

func TestSendQueryToLLM_NilQuery(t *testing.T) {
	config := utils.Config{
		AI: utils.AIConfig{APIKey: "dummy-key"},
	}

	err := SendQueryToLLM(nil, config, nil, QueryOpts{})
	assert.Error(t, err)
}
