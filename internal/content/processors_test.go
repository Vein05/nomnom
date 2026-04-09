package content

import (
	"os"
	"path/filepath"
	"testing"

	utils "nomnom/internal/utils"
)

func TestNewQuery(t *testing.T) {
	scan := ScanResult{
		RootDir: "/tmp/demo",
		Files: []ScannedFile{
			{
				SourcePath:   "/tmp/demo/test.txt",
				RelativePath: "test.txt",
				OriginalName: "test.txt",
			},
		},
	}

	query := NewQuery(QueryParams{
		Prompt:      "Custom prompt",
		ConfigPath:  "config.json",
		DryRun:      true,
		Reporter:    utils.NopReporter{},
		Analytics:   utils.NewAnalyticsStore(t.TempDir(), true),
		Scan:        scan,
		AutoApprove: false,
	})

	if query == nil {
		t.Fatal("NewQuery() returned nil query")
	}
	if query.Prompt != "Custom prompt" {
		t.Fatalf("NewQuery() prompt = %q, want %q", query.Prompt, "Custom prompt")
	}
	if query.Dir != scan.RootDir {
		t.Fatalf("NewQuery() dir = %q, want %q", query.Dir, scan.RootDir)
	}
	if len(query.Scan.Files) != 1 {
		t.Fatalf("NewQuery() scanned files = %d, want 1", len(query.Scan.Files))
	}
}

func TestNewSafeProcessor(t *testing.T) {
	query := &Query{Dir: "testdata"}
	processor := NewSafeProcessor(query, "output")
	if processor == nil {
		t.Fatal("NewSafeProcessor() returned nil")
	}
}

func TestSafeProcessorProcess(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	sourcePath := filepath.Join(inputDir, "test.txt")
	if err := os.WriteFile(sourcePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	query := &Query{
		Dir:         inputDir,
		DryRun:      false,
		AutoApprove: true,
		Scan: ScanResult{
			RootDir: inputDir,
			Files: []ScannedFile{
				{
					SourcePath:   sourcePath,
					RelativePath: "test.txt",
					OriginalName: "test.txt",
					Category:     "Documents",
				},
			},
		},
		Plan: []RenamePlanEntry{
			{
				File: ScannedFile{
					SourcePath:   sourcePath,
					RelativePath: "test.txt",
					OriginalName: "test.txt",
					Category:     "Documents",
				},
				SuggestedName: "renamed_test.txt",
			},
		},
		Reporter: utils.NopReporter{},
	}

	results, err := NewSafeProcessor(query, outputDir).Process()
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Process() results len = %d, want 1", len(results))
	}
	if !results[0].Success {
		t.Fatalf("Process() result not successful: %v", results[0].Error)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "renamed_test.txt")); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
}

func TestSafeProcessorProcessOrganized(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	sourcePath := filepath.Join(inputDir, "notes.txt")
	if err := os.WriteFile(sourcePath, []byte("notes"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	query := &Query{
		Dir:         inputDir,
		DryRun:      false,
		Organize:    true,
		AutoApprove: true,
		Plan: []RenamePlanEntry{
			{
				File: ScannedFile{
					SourcePath:   sourcePath,
					RelativePath: "notes.txt",
					OriginalName: "notes.txt",
					Category:     "Documents",
				},
				SuggestedName: "project_notes.txt",
			},
		},
		Reporter: utils.NopReporter{},
	}

	results, err := NewSafeProcessor(query, outputDir).Process()
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(results) != 1 || !results[0].Success {
		t.Fatalf("unexpected results: %#v", results)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "Documents", "project_notes.txt")); err != nil {
		t.Fatalf("expected organized output file to exist: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	if err := os.WriteFile(srcPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(content) != "test content" {
		t.Fatalf("copyFile() content mismatch = %q", string(content))
	}
}

func TestResolvePrompt(t *testing.T) {
	imagePrompt, err := os.ReadFile(NomNomPrompts[1].TestPath)
	if err != nil {
		t.Fatalf("failed to read image prompt: %v", err)
	}
	researchPrompt, err := os.ReadFile(NomNomPrompts[0].TestPath)
	if err != nil {
		t.Fatalf("failed to read research prompt: %v", err)
	}

	tests := []struct {
		name     string
		prompt   string
		config   utils.Config
		expected string
	}{
		{name: "Default prompt", prompt: "", config: utils.Config{}, expected: defaultPrompt},
		{name: "Images prompt", prompt: "images", config: utils.Config{}, expected: string(imagePrompt)},
		{name: "Research prompt", prompt: "research", config: utils.Config{}, expected: string(researchPrompt)},
		{name: "Config prompt", prompt: "", config: utils.Config{AI: utils.AIConfig{Prompt: "Custom prompt from config"}}, expected: "Custom prompt from config"},
		{name: "Custom prompt", prompt: "Custom prompt", config: utils.Config{}, expected: "Custom prompt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolvePrompt(tt.prompt, tt.config)
			if err != nil {
				t.Fatalf("ResolvePrompt() error = %v", err)
			}
			if result != tt.expected {
				t.Fatalf("ResolvePrompt() = %q, want %q", result, tt.expected)
			}
		})
	}
}
