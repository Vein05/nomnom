package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLiveCLISmoke(t *testing.T) {
	if os.Getenv("NOMNOM_LIVE_SMOKE") != "1" {
		t.Skip("set NOMNOM_LIVE_SMOKE=1 to run the live CLI smoke test")
	}

	provider := strings.TrimSpace(os.Getenv("NOMNOM_LIVE_PROVIDER"))
	model := strings.TrimSpace(os.Getenv("NOMNOM_LIVE_MODEL"))
	if provider == "" || model == "" {
		t.Skip("set NOMNOM_LIVE_PROVIDER and NOMNOM_LIVE_MODEL for the live CLI smoke test")
	}

	switch provider {
	case "openrouter":
		if strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")) == "" {
			t.Skip("OPENROUTER_API_KEY is required for NOMNOM_LIVE_PROVIDER=openrouter")
		}
	case "deepseek":
		if strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")) == "" {
			t.Skip("DEEPSEEK_API_KEY is required for NOMNOM_LIVE_PROVIDER=deepseek")
		}
	case "ollama":
		// Ollama uses the local daemon and does not need an API key.
	default:
		t.Fatalf("unsupported NOMNOM_LIVE_PROVIDER %q", provider)
	}

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")
	configPath := filepath.Join(tmpDir, "config.json")
	binaryPath := filepath.Join(tmpDir, "nomnom-smoke")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(inputDir) error = %v", err)
	}

	sourceFile := filepath.Join(inputDir, "project_notes.txt")
	sourceContent := []byte("Quarterly roadmap notes for NomNom smoke testing.")
	if err := os.WriteFile(sourceFile, sourceContent, 0o644); err != nil {
		t.Fatalf("WriteFile(sourceFile) error = %v", err)
	}

	config := map[string]any{
		"output": outputDir,
		"case":   "snake",
		"ai": map[string]any{
			"provider":    provider,
			"model":       model,
			"prompt":      "Rename files based on their content. Return only the new filename with the original extension in snake case.",
			"max_tokens":  64,
			"temperature": 0.0,
			"vision": map[string]any{
				"enabled":        false,
				"max_image_size": "10MB",
			},
		},
		"file_handling": map[string]any{
			"max_size":     "10MB",
			"auto_approve": true,
			"move_files":   false,
		},
		"content_extraction": map[string]any{
			"extract_text":       true,
			"extract_metadata":   true,
			"max_content_length": 512,
			"skip_large_files":   false,
			"read_context":       true,
		},
		"performance": map[string]any{
			"ai": map[string]any{
				"workers": 1,
				"timeout": "60s",
				"retries": 1,
			},
			"file": map[string]any{
				"workers": 1,
				"timeout": "30s",
				"retries": 1,
			},
		},
		"logging": map[string]any{
			"enabled":  true,
			"log_path": ".nomnom/logs",
		},
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(config) error = %v", err)
	}
	configBytes = append(configBytes, '\n')
	if err := os.WriteFile(configPath, configBytes, 0o644); err != nil {
		t.Fatalf("WriteFile(configPath) error = %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	buildCtx, buildCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer buildCancel()

	buildCmd := exec.CommandContext(buildCtx, "go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = cwd
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(buildOutput))
	}

	runCtx, runCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer runCancel()

	runCmd := exec.CommandContext(
		runCtx,
		binaryPath,
		"-d", inputDir,
		"-c", configPath,
		"--dry-run=false",
		"--auto-approve",
		"--organize=false",
	)
	runCmd.Dir = cwd
	runCmd.Env = os.Environ()
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("live CLI smoke run failed: %v\n%s", err, string(runOutput))
	}

	outputEntries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("ReadDir(outputDir) error = %v", err)
	}
	if len(outputEntries) != 1 {
		t.Fatalf("output file count = %d, want 1\nCLI output:\n%s", len(outputEntries), string(runOutput))
	}

	resultPath := filepath.Join(outputDir, outputEntries[0].Name())
	resultContent, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("ReadFile(resultPath) error = %v", err)
	}
	if string(resultContent) != string(sourceContent) {
		t.Fatalf("copied content mismatch at %s", resultPath)
	}

	logMatches, err := filepath.Glob(filepath.Join(inputDir, ".nomnom", "logs", "changes_*.json"))
	if err != nil {
		t.Fatalf("Glob(logs) error = %v", err)
	}
	if len(logMatches) == 0 {
		t.Fatalf("expected at least one change log file\nCLI output:\n%s", string(runOutput))
	}

	t.Logf("live CLI smoke passed with provider=%s model=%s output=%s", provider, model, resultPath)
}
