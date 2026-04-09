package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func demoDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "demo")
}

func TestReadFileText(t *testing.T) {
	content, err := ReadFile(filepath.Join(demoDir(t), "abcd.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if strings.TrimSpace(content) == "" {
		t.Fatal("ReadFile() returned empty content for text file")
	}
}

func TestExtractFileContentImageUsesVisualSource(t *testing.T) {
	path := filepath.Join(demoDir(t), "image1.png")

	content, err := ExtractFileContent(path)
	if err != nil {
		t.Fatalf("ExtractFileContent() error = %v", err)
	}

	if content.PreviewImagePath != path {
		t.Fatalf("PreviewImagePath = %q, want %q", content.PreviewImagePath, path)
	}

	if !strings.Contains(content.Text, "image preview is available") {
		t.Fatalf("Text = %q, want image preview guidance", content.Text)
	}
}

func TestExtractFileContentDocumentFallbackIsStructured(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.pdf")

	if err := os.WriteFile(path, []byte("not a real pdf"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, err := ExtractFileContent(path)
	if err != nil {
		t.Fatalf("ExtractFileContent() error = %v", err)
	}

	if content.PreviewImagePath != "" {
		t.Fatalf("PreviewImagePath = %q, want empty string", content.PreviewImagePath)
	}

	wantSnippets := []string{
		"Document extraction fallback.",
		"File: broken.pdf",
		"Extension: .pdf",
		"Size: 14 bytes",
		"Use the filename and any available visual preview to infer a better name.",
	}

	for _, snippet := range wantSnippets {
		if !strings.Contains(content.Text, snippet) {
			t.Fatalf("Text %q does not contain %q", content.Text, snippet)
		}
	}

	if strings.Contains(content.Text, "not a real pdf") {
		t.Fatalf("Text %q should not include raw file bytes", content.Text)
	}
}

func TestReadFileMissingFileReturnsError(t *testing.T) {
	_, err := ReadFile(filepath.Join(demoDir(t), "nonexistent.txt"))
	if err == nil {
		t.Fatal("ReadFile() error = nil, want error")
	}
}

func TestListFiles(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(demoDir(t), "*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	if len(files) == 0 {
		t.Fatal("expected demo directory to contain files")
	}
}
