package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(`{
  "output": "out",
  "case": "snake",
  "ai": {
    "provider": "openrouter",
    "model": "test-model",
    "prompt": "rename files"
  }
}`), 0644); err != nil {
		t.Fatalf("failed to write config fixture: %v", err)
	}

	config, err := LoadConfig(configPath, "")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Output != "out" {
		t.Fatalf("LoadConfig() output = %q, want %q", config.Output, "out")
	}
	if config.AI.Provider != "openrouter" {
		t.Fatalf("LoadConfig() provider = %q, want %q", config.AI.Provider, "openrouter")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"), "")
	if err == nil {
		t.Fatal("LoadConfig() expected error for missing config file")
	}
}
