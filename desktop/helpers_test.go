package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "punctuation", in: "Hello, World!", want: "hello_world"},
		{name: "spacing", in: "  spaced   out  ", want: "spaced_out"},
		{name: "mixed symbols", in: "NomNom v1.0", want: "nomnom_v1_0"},
		{name: "all separators", in: "___", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slugify(tt.in); got != tt.want {
				t.Fatalf("slugify(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestOllamaModelMatches(t *testing.T) {
	tests := []struct {
		name      string
		selected  string
		available string
		want      bool
	}{
		{name: "exact match", selected: "llama3.2", available: "llama3.2", want: true},
		{name: "tag suffix", selected: "llama3.2", available: "llama3.2:latest", want: true},
		{name: "reverse tag suffix", selected: "llama3.2:latest", available: "llama3.2", want: true},
		{name: "mismatch", selected: "llama3.2", available: "qwen2.5", want: false},
		{name: "empty selected", selected: "", available: "llama3.2", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ollamaModelMatches(tt.selected, tt.available); got != tt.want {
				t.Fatalf("ollamaModelMatches(%q, %q) = %t, want %t", tt.selected, tt.available, got, tt.want)
			}
		})
	}
}

func TestResolveOpenRouterAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "env-key")

	key, source := resolveOpenRouterAPIKey(" explicit-key ")
	if key != "explicit-key" || source != "config" {
		t.Fatalf("explicit key = (%q, %q), want (%q, %q)", key, source, "explicit-key", "config")
	}

	key, source = resolveOpenRouterAPIKey("")
	if key != "env-key" || source != "env" {
		t.Fatalf("env key = (%q, %q), want (%q, %q)", key, source, "env-key", "env")
	}

	t.Setenv("OPENROUTER_API_KEY", "")
	key, source = resolveOpenRouterAPIKey("")
	if key != "" || source != "missing" {
		t.Fatalf("missing key = (%q, %q), want (%q, %q)", key, source, "", "missing")
	}
}

func TestNormalizeConfigPath(t *testing.T) {
	t.Run("resolves relative file", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() error = %v", err)
		}
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Chdir() error = %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(cwd)
		})

		want, err := filepath.Abs("config.json")
		if err != nil {
			t.Fatalf("Abs() error = %v", err)
		}
		got, err := normalizeConfigPath("config.json")
		if err != nil {
			t.Fatalf("normalizeConfigPath() error = %v", err)
		}
		if got != want {
			t.Fatalf("normalizeConfigPath() = %q, want %q", got, want)
		}
	})

	t.Run("rejects directories", func(t *testing.T) {
		if _, err := normalizeConfigPath(t.TempDir()); err == nil {
			t.Fatal("normalizeConfigPath() expected error for directory path")
		}
	})
}

func TestSeedFileDialog(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "input")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	filePath := filepath.Join(nestedDir, "config.json")
	if err := os.WriteFile(filePath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	dir, name := seedFileDialog(nestedDir)
	if dir != nestedDir || name != "config.json" {
		t.Fatalf("seedFileDialog(dir) = (%q, %q), want (%q, %q)", dir, name, nestedDir, "config.json")
	}

	dir, name = seedFileDialog(filePath)
	if dir != nestedDir || name != "config.json" {
		t.Fatalf("seedFileDialog(file) = (%q, %q), want (%q, %q)", dir, name, nestedDir, "config.json")
	}

	dir, name = seedFileDialog("")
	if dir != "" || name != "" {
		t.Fatalf("seedFileDialog(empty) = (%q, %q), want empty values", dir, name)
	}
}

func TestProbeOllamaStatus(t *testing.T) {
	originalURL := ollamaTagsURL
	t.Cleanup(func() {
		ollamaTagsURL = originalURL
	})

	t.Run("matches model", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/tags" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"models":[{"name":"llama3.2:latest"}]}`))
		}))
		defer server.Close()

		ollamaTagsURL = server.URL + "/api/tags"

		status := probeOllamaStatus("llama3.2")
		if !status.Running || !status.ModelAvailable {
			t.Fatalf("probeOllamaStatus() = %+v, want running model match", status)
		}
		if !strings.Contains(status.Message, "available") {
			t.Fatalf("probeOllamaStatus() message = %q, want availability text", status.Message)
		}
	})

	t.Run("handles bad response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("unavailable"))
		}))
		defer server.Close()

		ollamaTagsURL = server.URL + "/api/tags"

		status := probeOllamaStatus("llama3.2")
		if status.Running || status.ModelAvailable {
			t.Fatalf("probeOllamaStatus() = %+v, want non-running status", status)
		}
		if !strings.Contains(status.Message, "503") {
			t.Fatalf("probeOllamaStatus() message = %q, want status text", status.Message)
		}
	})
}
