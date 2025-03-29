package nomnom

import (
	utils "nomnom/internal/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestNewQuery(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		dir         string
		configPath  string
		config      utils.Config
		autoApprove bool
		dryRun      bool
		verbose     bool
		wantErr     bool
		log         bool
	}{
		{
			name:       "Valid query with default prompt",
			prompt:     "",
			dir:        "demo",
			configPath: "config.yaml",
			wantErr:    false,
		},
		{
			name:       "Valid query with custom prompt",
			prompt:     "Custom prompt",
			dir:        "demo",
			configPath: "config.yaml",
			wantErr:    false,
		},
		{
			name:       "Invalid directory",
			prompt:     "",
			dir:        "nonexistent",
			configPath: "config.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := NewQuery(tt.prompt, tt.dir, tt.configPath, tt.config, tt.autoApprove, tt.dryRun, tt.log, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && query == nil {
				t.Error("NewQuery() returned nil query without error")
			}
		})
	}
}

func TestNewSafeProcessor(t *testing.T) {
	query := &Query{
		Dir:        "testdata",
		ConfigPath: "config.yaml",
	}
	output := "output"

	processor := NewSafeProcessor(query, output)
	if processor == nil {
		t.Fatal("NewSafeProcessor() returned nil")
	}

	if processor.query != query {
		t.Error("NewSafeProcessor() did not set query correctly")
	}
	if processor.output != output {
		t.Error("NewSafeProcessor() did not set output correctly")
	}
}

func TestSafeProcessor_Process(t *testing.T) {
	// Create temporary test directories
	tmpDir, err := os.MkdirTemp("", "nomnom-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	// Create test file structure
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(inputDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		query   *Query
		dryRun  bool
		wantErr bool
	}{
		{
			name: "Successful processing",
			query: &Query{
				Dir:    inputDir,
				DryRun: false,
				Folders: []FolderType{{
					Name: "input",
					FileList: []File{{
						Name: "test.txt",
						Path: testFile,
					}},
				}},
			},
			wantErr: false,
		},
		{
			name: "Dry run processing",
			query: &Query{
				Dir:    inputDir,
				DryRun: true,
				Folders: []FolderType{{
					Name: "input",
					FileList: []File{{
						Name: "test.txt",
						Path: testFile,
					}},
				}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeProcessor(tt.query, outputDir)
			results, err := processor.Process()
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(results) == 0 {
					t.Error("Process() returned no results")
				}
				for _, result := range results {
					if !result.Success {
						t.Errorf("Process() result not successful: %v", result.Error)
					}
				}
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "nomnom-copyfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcContent := []byte("test content")
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copying
	dstPath := filepath.Join(tmpDir, "destination.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() error = %v", err)
		return
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
		return
	}

	if string(dstContent) != string(srcContent) {
		t.Errorf("copyFile() content mismatch: got %v, want %v", string(dstContent), string(srcContent))
	}

	// Test error cases
	if err := copyFile("nonexistent", dstPath); err == nil {
		t.Error("copyFile() should fail with nonexistent source")
	}
}

func TestCopyOrganizedStructure(t *testing.T) {
	// Create temporary test directories
	tmpDir, err := os.MkdirTemp("", "nomnom-organize-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	// Create test file structure
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}

	// Create test files of different types
	testFiles := map[string][]byte{
		"document.txt":  []byte("text content"),
		"image.png":     []byte("mock image data"),
		"document.pdf":  []byte("mock pdf data"),
		"audio.mp3":     []byte("mock audio data"),
		"document.docx": []byte("mock docx content"),
	}

	fileList := []File{}
	for name, content := range testFiles {
		filePath := filepath.Join(inputDir, name)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
		fileList = append(fileList, File{
			Name:    name,
			NewName: name,
			Path:    filePath,
		})
	}

	// Create a Query with Organize flag set to true
	query := &Query{
		Dir:         inputDir,
		DryRun:      false,
		AutoApprove: true,
		Organize:    true,
		Folders: []FolderType{{
			Name:     "input",
			FileList: fileList,
		}},
	}

	// Create and process with SafeProcessor
	processor := NewSafeProcessor(query, outputDir)
	results, err := processor.Process()
	if err != nil {
		t.Errorf("Process() error = %v", err)
		return
	}

	// Verify results
	if len(results) == 0 {
		t.Error("Process() returned no results")
		return
	}

	// Verify files were organized into category folders
	categoryFolders := map[string]int{
		"Documents": 3, // expecting 3 documents (txt, pdf, docx)
		"Images":    1, // expecting 1 image (png)
		"Audios":    1, // expecting 1 audio (mp3)
	}

	for folder, expectedCount := range categoryFolders {
		folderPath := filepath.Join(outputDir, folder)
		files, err := os.ReadDir(folderPath)
		if err != nil {
			t.Errorf("Failed to read category folder %s: %v", folder, err)
			continue
		}

		if len(files) != expectedCount {
			t.Errorf("Category %s has incorrect number of files. Got: %d, Want: %d",
				folder, len(files), expectedCount)
		}
	}
}
