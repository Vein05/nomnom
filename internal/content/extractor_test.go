package content

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	utils "nomnom/internal/utils"
)

func TestProcessDirectory(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	path := filepath.Join(repoRoot, "demo")
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

	scan, err := ScanDirectory(path, config, utils.NopReporter{})
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
	// test 100MB
	size := "100MB"
	convertedSize, err := convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*MB {
		t.Fatalf("Expected 100MB, got %d", convertedSize)
	}

	// test 100KB
	size = "100KB"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*KB {
		t.Fatalf("Expected 100KB, got %d", convertedSize)
	}

	// test 100GB
	size = "100GB"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*GB {
		t.Fatalf("Expected 100GB, got %d", convertedSize)
	}

	// test 100B
	size = "100B"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100 {
		t.Fatalf("Expected 100B, got %d", convertedSize)
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
