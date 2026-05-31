package content

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	utils "nomnom/internal/utils"
)

func TestProcessDirectory(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.png", "c.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", name, err)
		}
	}

	config := utils.Config{
		FileHandling: utils.FileHandlingConfig{
			MaxSize: "100MB",
		},
		Performance: utils.PerformanceConfig{
			File: utils.PerformanceFileConfig{
				Workers: 1,
				Timeout: "30s",
				Retries: 1,
			},
		},
	}

	scan, err := ScanDirectory(dir, config, utils.NopReporter{})
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if scan.RootDir == "" {
		t.Fatal("ScanDirectory() returned empty root dir")
	}

	if len(scan.Files) == 0 {
		t.Fatal("ScanDirectory() returned no files")
	}

	for index := 1; index < len(scan.Files); index++ {
		if scan.Files[index-1].RelativePath > scan.Files[index].RelativePath {
			t.Fatalf("ScanDirectory() returned files out of order: %q before %q", scan.Files[index-1].RelativePath, scan.Files[index].RelativePath)
		}
	}
}

func TestConvertSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100MB", 100 * MB},
		{"100KB", 100 * KB},
		{"100GB", 100 * GB},
		{"100B", 100},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := convertSize(tt.input)
			if err != nil {
				t.Fatalf("convertSize(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("convertSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertSize_Invalid(t *testing.T) {
	_, err := convertSize("invalid")
	if err == nil {
		t.Fatal("convertSize() error = nil, want error for invalid input")
	}
}

func TestScanResultCleanupRemovesGeneratedPreviews(t *testing.T) {
	tmpDir := t.TempDir()
	previewPath := filepath.Join(tmpDir, "preview.jpg")
	sourcePath := filepath.Join(tmpDir, "source.pdf")
	imagePath := filepath.Join(tmpDir, "image.png")

	for _, path := range []string{previewPath, sourcePath, imagePath} {
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}

	scan := ScanResult{
		Files: []ScannedFile{
			{SourcePath: sourcePath, VisualPath: previewPath},
			{SourcePath: imagePath, VisualPath: imagePath},
		},
	}

	if err := scan.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if _, err := os.Stat(previewPath); !os.IsNotExist(err) {
		t.Fatalf("Cleanup() should remove generated preview, stat err = %v", err)
	}

	if _, err := os.Stat(imagePath); err != nil {
		t.Fatalf("Cleanup() should keep source image, stat err = %v", err)
	}
}

func TestParseFileTimeoutParsesNumericSeconds(t *testing.T) {
	timeout, err := parseFileTimeout("30", "30s")
	if err != nil {
		t.Fatalf("parseFileTimeout() error = %v", err)
	}

	if timeout != 30*time.Second {
		t.Fatalf("parseFileTimeout() = %s, want %s", timeout, 30*time.Second)
	}
}

// -- New tests --

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{size: 0, want: "0B"},
		{size: 500, want: "500B"},
		{size: 1024, want: "1.00KB"},
		{size: 2048, want: "2.00KB"},
		{size: 1024 * 1024, want: "1.00MB"},
		{size: 1024 * 1024 * 1024, want: "1.00GB"},
		{size: 1500 * 1024, want: "1.46MB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatFileSize(tt.size)
			if got != tt.want {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		skipDotFile bool
		want        bool
	}{
		{name: "dotfile with skip", fileName: ".hidden", skipDotFile: true, want: true},
		{name: "dotfile without skip", fileName: ".hidden", skipDotFile: false, want: false},
		{name: "temp file", fileName: "file.tmp", skipDotFile: false, want: true},
		{name: "swap file", fileName: "file.swp", skipDotFile: false, want: true},
		{name: "backup file", fileName: "file~", skipDotFile: false, want: true},
		{name: "normal file", fileName: "normal.txt", skipDotFile: false, want: false},
		{name: "normal file with dotfile skip", fileName: "normal.txt", skipDotFile: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkip(tt.fileName, tt.skipDotFile)
			if got != tt.want {
				t.Errorf("shouldSkip(%q, %v) = %v, want %v", tt.fileName, tt.skipDotFile, got, tt.want)
			}
		})
	}
}

func TestScanFile_Oversized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.txt")

	if err := os.WriteFile(path, []byte("content that is too large"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := scanFile(dir, path, 5) // maxSize = 5 bytes
	if err == nil {
		t.Fatal("scanFile() error = nil, want error for oversized file")
	}
}

func TestScanFile_NonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	_, err := scanFile(dir, path, 1000)
	if err == nil {
		t.Fatal("scanFile() error = nil, want error for non-existent file")
	}
}

func TestScanFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.txt")

	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, err := scanFile(dir, path, 1000)
	if err != nil {
		t.Fatalf("scanFile() error = %v", err)
	}

	if file.SourcePath != path {
		t.Errorf("SourcePath = %q, want %q", file.SourcePath, path)
	}
	if file.RelativePath != "notes.txt" {
		t.Errorf("RelativePath = %q, want %q", file.RelativePath, "notes.txt")
	}
	if file.OriginalName != "notes.txt" {
		t.Errorf("OriginalName = %q, want %q", file.OriginalName, "notes.txt")
	}
	if file.Extension != ".txt" {
		t.Errorf("Extension = %q, want %q", file.Extension, ".txt")
	}
	if file.Size != 11 {
		t.Errorf("Size = %d, want %d", file.Size, 11)
	}
	if file.Category != "Documents" {
		t.Errorf("Category = %q, want %q", file.Category, "Documents")
	}
}

func TestScanFileWithTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fast.txt")

	if err := os.WriteFile(path, []byte("fast file"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, err := scanFileWithTimeout(dir, path, 1000, 5*time.Second)
	if err != nil {
		t.Fatalf("scanFileWithTimeout() error = %v", err)
	}
	if file == (ScannedFile{}) {
		t.Fatal("scanFileWithTimeout() returned empty ScannedFile")
	}
}

func TestScanFileWithTimeout_ZeroTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fast.txt")

	if err := os.WriteFile(path, []byte("fast"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, err := scanFileWithTimeout(dir, path, 1000, 0)
	if err != nil {
		t.Fatalf("scanFileWithTimeout() error = %v", err)
	}
	if file == (ScannedFile{}) {
		t.Fatal("scanFileWithTimeout() returned empty ScannedFile")
	}
}

func TestParseFileTimeout(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		fallback string
		want     time.Duration
	}{
		{name: "duration string", raw: "5s", fallback: "30s", want: 5 * time.Second},
		{name: "numeric seconds", raw: "10", fallback: "30s", want: 10 * time.Second},
		{name: "empty uses fallback", raw: "", fallback: "15s", want: 15 * time.Second},
		{name: "minutes", raw: "1m", fallback: "30s", want: time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileTimeout(tt.raw, tt.fallback)
			if err != nil {
				t.Fatalf("parseFileTimeout() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("parseFileTimeout(%q, %q) = %s, want %s", tt.raw, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestParseFileTimeout_Invalid(t *testing.T) {
	_, err := parseFileTimeout("not-a-number", "also-bad")
	if err == nil {
		t.Fatal("parseFileTimeout() error = nil, want error for invalid")
	}
}

func TestScanFileWithRetry_SuccessFirstTry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quick.txt")

	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := scanFileWithRetry(dir, path, 1000, 3, time.Second)
	if err != nil {
		t.Fatalf("scanFileWithRetry() error = %v", err)
	}
}

func TestScanFileWithRetry_AllFail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.txt")

	_, err := scanFileWithRetry(dir, path, 1000, 2, time.Second)
	if err == nil {
		t.Fatal("scanFileWithRetry() error = nil, want error after retries")
	}
}

func TestCollectPaths_SkipsDotNomnom(t *testing.T) {
	dir := t.TempDir()

	// Create a file and a dotnomnom directory
	if err := os.WriteFile(filepath.Join(dir, "real.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".nomnom"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".nomnom", "cache.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	var paths []string
	if err := collectPaths(dir, entries, &paths, false); err != nil {
		t.Fatalf("collectPaths() error = %v", err)
	}

	if len(paths) != 1 {
		t.Fatalf("collectPaths() returned %d files, want 1 (should skip .nomnom)", len(paths))
	}
}

func TestCollectPaths_SkipsNomnomDir(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "real.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "nomnom", "renamed"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nomnom", "renamed", "output.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	var paths []string
	if err := collectPaths(dir, entries, &paths, false); err != nil {
		t.Fatalf("collectPaths() error = %v", err)
	}

	if len(paths) != 1 {
		t.Fatalf("collectPaths() returned %d files, want 1 (should skip nomnom dir)", len(paths))
	}
}

func TestMaxFileSize(t *testing.T) {
	tests := []struct {
		name    string
		config  utils.Config
		want    int64
		wantErr bool
	}{
		{name: "empty config", config: utils.Config{}, want: 0, wantErr: false},
		{name: "valid size", config: utils.Config{FileHandling: utils.FileHandlingConfig{MaxSize: "10MB"}}, want: 10 * MB, wantErr: false},
		{name: "invalid size", config: utils.Config{FileHandling: utils.FileHandlingConfig{MaxSize: "invalid"}}, want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := maxFileSize(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("maxFileSize() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("maxFileSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
