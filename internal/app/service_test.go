package app

import (
	"os"
	"path/filepath"
	"testing"

	"nomnom/internal/utils"
)

func TestPrepareRunAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "hello.txt"), []byte("hello world"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{
  "case": "snake",
  "ai": {
    "provider": "deepseek",
    "api_key": "dummy-key"
  }
}`), 0o644); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	service := NewService()
	run, err := service.PrepareRun(RunOptions{
		Dir:        inputDir,
		ConfigPath: configPath,
		DryRun:     true,
		Log:        true,
	}, utils.NopReporter{}, nil)
	if err != nil {
		t.Fatalf("PrepareRun() error = %v", err)
	}

	if run.OutputDir == "" {
		t.Fatal("PrepareRun() returned empty output dir")
	}
	if len(run.Query.Scan.Files) != 1 {
		t.Fatalf("PrepareRun() scanned files = %d, want 1", len(run.Query.Scan.Files))
	}

	if err := service.GeneratePlan(run); err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}
	if len(run.Query.Plan) != 0 {
		t.Fatalf("GeneratePlan() plan len = %d, want 0 when using dummy key", len(run.Query.Plan))
	}

	if err := run.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	summary, sessions, err := service.LoadAnalytics(inputDir)
	if err != nil {
		t.Fatalf("LoadAnalytics() error = %v", err)
	}
	if summary.Sessions != 1 {
		t.Fatalf("Sessions = %d, want 1", summary.Sessions)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions len = %d, want 1", len(sessions))
	}
}
