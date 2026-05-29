package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFileContentWithOptionsRespectsMaxTextBytes(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "long.txt")
	content := strings.Repeat("nomnom", 32)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{MaxTextBytes: 18, GeneratePreview: false})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}

	if len(extracted.Text) != 18 {
		t.Fatalf("Extracted length = %d, want 18", len(extracted.Text))
	}
	if extracted.Text != content[:18] {
		t.Fatalf("Extracted text = %q, want prefix %q", extracted.Text, content[:18])
	}
}
