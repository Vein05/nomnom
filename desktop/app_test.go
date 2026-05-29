package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newDesktopOllamaMockServer(responseName string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := map[string]any{
			"model":             request["model"],
			"created_at":        time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			"message":           map[string]any{"role": "assistant", "content": responseName},
			"done_reason":       "stop",
			"done":              true,
			"prompt_eval_count": 12,
			"eval_count":        4,
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestAppScanRunHistoryAndAnalytics(t *testing.T) {
	server := newDesktopOllamaMockServer("renamed_plan")
	defer server.Close()

	t.Setenv("OLLAMA_HOST", server.URL)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "desktop-config.json")
	sourceDir := filepath.Join(tmpDir, "input")

	if err := os.MkdirAll(filepath.Join(sourceDir, ".hidden"), 0o755); err != nil {
		t.Fatalf("MkdirAll(hidden) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "Quarterly Plan.txt"), []byte("Roadmap notes"), 0o644); err != nil {
		t.Fatalf("WriteFile(visible) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, ".hidden", "secret.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatalf("WriteFile(hidden) error = %v", err)
	}

	app := NewApp()
	app.configPath = configPath

	config := defaultDesktopConfig()
	config.AI.Provider = "ollama"
	config.AI.Model = "mock-ollama-model"
	config.AI.Vision.Enabled = false

	saved, err := app.SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	if !saved {
		t.Fatal("SaveConfig() = false, want true")
	}

	jobID, err := app.ScanDirectory(sourceDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	plan, err := app.GetPlan(jobID)
	if err != nil {
		t.Fatalf("GetPlan() error = %v", err)
	}
	if len(plan) != 1 {
		t.Fatalf("plan len = %d, want 1 visible file", len(plan))
	}
	if plan[0].Original != "Quarterly Plan.txt" {
		t.Fatalf("plan original = %q, want %q", plan[0].Original, "Quarterly Plan.txt")
	}
	// Scan returns file names, not AI-generated names
	if plan[0].NewName != "Quarterly Plan.txt" {
		t.Fatalf("plan new name = %q, want original file name %q", plan[0].NewName, "Quarterly Plan.txt")
	}

	if err := app.GenerateNames(jobID); err != nil {
		t.Fatalf("GenerateNames() error = %v", err)
	}

	plan, err = app.GetPlan(jobID)
	if err != nil {
		t.Fatalf("GetPlan() after generate error = %v", err)
	}
	if plan[0].NewName != "renamed_plan.txt" {
		t.Fatalf("plan new name after generate = %q, want %q", plan[0].NewName, "renamed_plan.txt")
	}

	if _, err := app.RunJob(jobID, RunJobOptions{DryRun: true, AutoApprove: true, Organize: true}); err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		status, err := app.GetJobStatus(jobID)
		if err != nil {
			t.Fatalf("GetJobStatus() error = %v", err)
		}
		if status.State == "complete" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("job did not complete before timeout; last status = %+v", status)
		}
		time.Sleep(20 * time.Millisecond)
	}

	history, err := app.GetHistory()
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].Mode != "Dry Run" {
		t.Fatalf("history mode = %q, want %q", history[0].Mode, "Dry Run")
	}
	if history[0].Model != "ollama / mock-ollama-model" {
		t.Fatalf("history model = %q, want %q", history[0].Model, "ollama / mock-ollama-model")
	}

	analytics, err := app.GetAnalytics()
	if err != nil {
		t.Fatalf("GetAnalytics() error = %v", err)
	}
	if analytics.Sessions != 1 {
		t.Fatalf("analytics sessions = %d, want 1", analytics.Sessions)
	}
	if analytics.Renamed != 1 {
		t.Fatalf("analytics renamed = %d, want 1", analytics.Renamed)
	}
	if analytics.Tokens != 16 {
		t.Fatalf("analytics tokens = %d, want 16", analytics.Tokens)
	}
}

func TestAppCancelJob(t *testing.T) {
	server := newDesktopOllamaMockServer("renamed_plan")
	defer server.Close()

	t.Setenv("OLLAMA_HOST", server.URL)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "desktop-config.json")
	sourceDir := filepath.Join(tmpDir, "input")

	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for i := 0; i < 12; i++ {
		path := filepath.Join(sourceDir, fmt.Sprintf("file_%02d.txt", i))
		if err := os.WriteFile(path, []byte("cancel me"), 0o644); err != nil {
			t.Fatalf("WriteFile(%d) error = %v", i, err)
		}
	}

	app := NewApp()
	app.configPath = configPath

	config := defaultDesktopConfig()
	config.AI.Provider = "ollama"
	config.AI.Model = "mock-ollama-model"
	config.AI.Vision.Enabled = false

	saved, err := app.SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	if !saved {
		t.Fatal("SaveConfig() = false, want true")
	}

	jobID, err := app.ScanDirectory(sourceDir)
	if err != nil {
		t.Fatalf("ScanDirectory() error = %v", err)
	}

	if _, err := app.RunJob(jobID, RunJobOptions{DryRun: false, AutoApprove: true, Organize: true}); err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if !app.CancelJob(jobID) {
		t.Fatal("CancelJob() = false, want true")
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		status, err := app.GetJobStatus(jobID)
		if err != nil {
			t.Fatalf("GetJobStatus() error = %v", err)
		}
		if status.State == "canceled" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("job did not cancel before timeout; last status = %+v", status)
		}
		time.Sleep(20 * time.Millisecond)
	}

	history, err := app.GetHistory()
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].Status != "Canceled" {
		t.Fatalf("history status = %q, want %q", history[0].Status, "Canceled")
	}
}
