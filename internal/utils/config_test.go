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

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.AI.Provider != "openrouter" {
		t.Fatalf("DefaultConfig() provider = %q, want %q", config.AI.Provider, "openrouter")
	}
	if config.Case != "snake" {
		t.Fatalf("DefaultConfig() case = %q, want %q", config.Case, "snake")
	}
	if !config.Logging.Enabled {
		t.Fatal("DefaultConfig() logging should be enabled")
	}
}

func TestSaveConfig(t *testing.T) {
	baseDir := t.TempDir()
	configPath := filepath.Join(baseDir, "nomnom", "config.json")
	config := DefaultConfig()
	config.AI.Provider = "deepseek"
	config.AI.APIKey = "test-key"

	savedPath, err := SaveConfig(configPath, config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	if savedPath != configPath {
		t.Fatalf("SaveConfig() path = %q, want %q", savedPath, configPath)
	}

	loaded, err := LoadConfig(configPath, "")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if loaded.AI.Provider != "deepseek" {
		t.Fatalf("saved provider = %q, want %q", loaded.AI.Provider, "deepseek")
	}
	if loaded.AI.APIKey != "test-key" {
		t.Fatalf("saved api key = %q, want %q", loaded.AI.APIKey, "test-key")
	}
}
