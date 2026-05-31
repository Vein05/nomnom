package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{name: "simple extension", path: "file.txt", expected: ".txt"},
		{name: "no extension", path: "file", expected: ""},
		{name: "multiple dots", path: "archive.tar.gz", expected: ".gz"},
		{name: "relative path", path: "dir/file.pdf", expected: ".pdf"},
		{name: "absolute path", path: "/home/user/file.jpg", expected: ".jpg"},
		{name: "dotfile", path: ".hidden", expected: ".hidden"},
		{name: "empty string", path: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileExtension(tt.path)
			if result != tt.expected {
				t.Errorf("GetFileExtension(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsDocumentFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{name: "pdf document", fileName: "report.pdf", expected: true},
		{name: "docx document", fileName: "paper.docx", expected: true},
		{name: "epub document", fileName: "book.epub", expected: true},
		{name: "pptx document", fileName: "slides.pptx", expected: true},
		{name: "xlsx document", fileName: "data.xlsx", expected: true},
		{name: "xls document", fileName: "old_data.xls", expected: true},
		{name: "txt not document", fileName: "notes.txt", expected: false},
		{name: "png not document", fileName: "image.png", expected: false},
		{name: "uppercase extension", fileName: "REPORT.PDF", expected: true},
		{name: "no extension", fileName: "Makefile", expected: false},
		{name: "empty string", fileName: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDocumentFile(tt.fileName)
			if result != tt.expected {
				t.Errorf("IsDocumentFile(%q) = %v, want %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

func TestCleanupPreviewTempDir(t *testing.T) {
	// Should be a no-op if no temp dir was created
	if err := CleanupPreviewTempDir(); err != nil {
		t.Fatalf("CleanupPreviewTempDir() on empty state error = %v", err)
	}
}

func TestExtractFileContentWithOptions_RawTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	content := "hello world"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    1024,
		GeneratePreview: false,
	})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}
	if extracted.Text != content {
		t.Fatalf("Text = %q, want %q", extracted.Text, content)
	}
	if extracted.PreviewImagePath != "" {
		t.Fatalf("PreviewImagePath = %q, want empty for text file", extracted.PreviewImagePath)
	}
}

func TestExtractFileContentWithOptions_ImageFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.png")

	// readImageFile just returns a template string regardless of content
	if err := os.WriteFile(path, []byte("not a real png"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    1024,
		GeneratePreview: true,
	})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}
	if extracted.Text == "" {
		t.Fatal("Text = empty, want non-empty placeholder text")
	}
	if extracted.PreviewImagePath != path {
		t.Fatalf("PreviewImagePath = %q, want %q (original path)", extracted.PreviewImagePath, path)
	}
}

func TestExtractFileContentWithOptions_UnsupportedFileTypeDefaultsToRawRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	content := []byte{0x00, 0x01, 0x02, 0x03}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    1024,
		GeneratePreview: false,
	})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}
	if extracted.Text != string(content) {
		t.Fatalf("Text = %q, want %q", extracted.Text, string(content))
	}
	if extracted.PreviewImagePath != "" {
		t.Fatalf("PreviewImagePath = %q, want empty", extracted.PreviewImagePath)
	}
}

func TestExtractFileContentWithOptions_MissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	_, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    1024,
		GeneratePreview: false,
	})
	if err == nil {
		t.Fatal("ExtractFileContentWithOptions() error = nil, want error for missing file")
	}
}

func TestExtractFileContentWithOptions_ZeroMaxBytesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.md")
	expected := "# Title\ncontent"

	if err := os.WriteFile(path, []byte(expected), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Zero MaxTextBytes should trigger the default limit
	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    0,
		GeneratePreview: false,
	})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}
	if extracted.Text != expected {
		t.Fatalf("Text = %q, want %q", extracted.Text, expected)
	}
}

func TestExtractFileContentWithOptions_JsonFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	content := `{"key": "value"}`

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	extracted, err := ExtractFileContentWithOptions(path, ExtractOptions{
		MaxTextBytes:    1024,
		GeneratePreview: false,
	})
	if err != nil {
		t.Fatalf("ExtractFileContentWithOptions() error = %v", err)
	}
	if extracted.Text != content {
		t.Fatalf("Text = %q, want %q", extracted.Text, content)
	}
}

func TestReadFileIntegration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "readme.txt")
	content := "integration test content"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	text, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if text != content {
		t.Fatalf("ReadFile() = %q, want %q", text, content)
	}
}

func TestIsTypeSupported(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		want     bool
	}{
		{name: "document type", fileType: "pdf", want: true},
		{name: "document type docx", fileType: "docx", want: true},
		{name: "code type", fileType: "go", want: true},
		{name: "code type python", fileType: "py", want: true},
		{name: "data type", fileType: "json", want: true},
		{name: "presentation type", fileType: "odp", want: true},
		{name: "spreadsheet type", fileType: "csv", want: true},
		{name: "unsupported type", fileType: "xyz", want: false},
		{name: "empty string", fileType: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTypeSupported(tt.fileType)
			if got != tt.want {
				t.Errorf("IsTypeSupported(%q) = %v, want %v", tt.fileType, got, tt.want)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{name: "png", fileName: "photo.png", want: true},
		{name: "jpg", fileName: "photo.jpg", want: true},
		{name: "jpeg", fileName: "photo.jpeg", want: true},
		{name: "webp", fileName: "photo.webp", want: true},
		{name: "txt not image", fileName: "file.txt", want: false},
		{name: "pdf not image", fileName: "doc.pdf", want: false},
		{name: "no extension", fileName: "Makefile", want: false},
		{name: "uppercase extension", fileName: "photo.PNG", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsImageFile(tt.fileName)
			if got != tt.want {
				t.Errorf("IsImageFile(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestIsAValidFileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantMsg string
	}{
		{name: "valid name", input: "valid_file.txt", wantOk: true, wantMsg: ""},
		{name: "empty string", input: "", wantOk: false, wantMsg: "cannot be empty"},
		{name: "no extension", input: "file", wantOk: false, wantMsg: "must have an extension"},
		{name: "contains space", input: "file name.txt", wantOk: false, wantMsg: "cannot contain spaces"},
		{name: "invalid chars", input: "file<.txt", wantOk: false, wantMsg: "cannot contain the character"},
		{name: "only dots", input: "....txt", wantOk: false, wantMsg: "cannot be only dots or spaces"},
		{name: "too long", input: strings.Repeat("a", 260) + ".txt", wantOk: false, wantMsg: "too long"},
		{name: "leading dot", input: ".hidden.txt", wantOk: false, wantMsg: "cannot start or end with a space or period"},
		{name: "leading space", input: " file.txt", wantOk: false, wantMsg: "cannot contain spaces"},
		{name: "reserved windows name", input: "CON.txt", wantOk: false, wantMsg: "reserved in Windows"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := IsAValidFileName(tt.input)
			if ok != tt.wantOk {
				t.Errorf("IsAValidFileName(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if tt.wantOk && msg != "" {
				t.Errorf("IsAValidFileName(%q) msg = %q, want empty", tt.input, msg)
			}
			if !tt.wantOk && !strings.Contains(msg, tt.wantMsg) {
				t.Errorf("IsAValidFileName(%q) msg = %q, want containing %q", tt.input, msg, tt.wantMsg)
			}
		})
	}
}

func TestCheckAndAddExtension_ReplaceExtension(t *testing.T) {
	result := CheckAndAddExtension("file.jpg", "ref.txt")
	want := "file.txt"
	if result != want {
		t.Errorf("CheckAndAddExtension(%q, %q) = %q, want %q", "file.jpg", "ref.txt", result, want)
	}
}
