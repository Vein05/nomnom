package nomnom

import (
	utils "nomnom/internal/utils"
	"os"
	"path/filepath"
	"slices"
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

			// check if prompt is set or not
			if tt.prompt != "" && query.Prompt != tt.prompt {
				t.Errorf("NewQuery() prompt = %v, want %v", query.Prompt, tt.prompt)
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

	// Create test folders and files
	testFiles := map[string]struct {
		content  []byte
		category string
	}{
		"document.txt": {[]byte("text content"), "Documents"},
		"image.jpg":    {[]byte("image data"), "Images"},
		"music.mp3":    {[]byte("audio data"), "Audios"},
		"video.mp4":    {[]byte("video data"), "Videos"},
		"unknown.xyz":  {[]byte("unknown data"), "Others"},
	}

	// Create input folder and files
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input dir: %v", err)
	}

	// Create query with test files
	var fileList []File
	for name, data := range testFiles {
		filePath := filepath.Join(inputDir, name)
		if err := os.WriteFile(filePath, data.content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
		fileList = append(fileList, File{
			Name: name,
			Path: filePath,
		})
	}

	query := &Query{
		Dir:      inputDir,
		Organize: true, // Enable organized mode
		Folders: []FolderType{{
			Name:     "test_folder",
			FileList: fileList,
		}},
	}

	// Create and run processor
	processor := NewSafeProcessor(query, outputDir)
	if err := processor.copyOrganizedStructure(); err != nil {
		t.Fatalf("copyOrganizedStructure() error = %v", err)
	}

	// Verify files were organized correctly
	for fileName, data := range testFiles {
		expectedPath := filepath.Join(outputDir, data.category, fileName)

		// Check if file exists
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", expectedPath, err)
			continue
		}

		// Verify content
		if string(content) != string(data.content) {
			t.Errorf("File content mismatch for %s. Got: %s, Want: %s",
				expectedPath, string(content), string(data.content))
		}
	}

	// Verify all category folders were created
	for _, category := range defaultCategories {
		categoryPath := filepath.Join(outputDir, category.Name)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			t.Errorf("Expected category folder %s was not created", category.Name)
		}
	}

	// Verify file categorization is correct
	for fileName, data := range testFiles {
		extension := filepath.Ext(fileName)
		expectedCategory := "Others"

		// Find the expected category based on file extension
		for _, category := range defaultCategories {
			if slices.Contains(category.Extensions, extension) {
				expectedCategory = category.Name
				break
			}
		}

		if expectedCategory != data.category {
			t.Errorf("File %s was categorized as %s, expected %s",
				fileName, data.category, expectedCategory)
		}
	}
}

func TestHandelPrompt(t *testing.T) {
	DEFAULT_PROMPT := "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension."
	file, err := os.ReadFile("../../data/prompts/images.txt")
	if err != nil {
		t.Fatalf("Failed to read image prompt file: %v", err)
	}
	IMAGE_PROMPT := string(file)

	file, err = os.ReadFile("../../data/prompts/research.txt")
	if err != nil {
		t.Fatalf("Failed to read research prompt file: %v", err)
	}
	RESEARCH_PROMPT := string(file)
	tests := []struct {
		name        string
		prompt      string
		config      utils.Config
		expected    string
		expectedErr bool
	}{
		{
			name:        "Test Empty prompt",
			prompt:      "",
			config:      utils.Config{},
			expected:    DEFAULT_PROMPT,
			expectedErr: false,
		},

		{
			name:        "Test Images prompt",
			prompt:      "images",
			config:      utils.Config{},
			expected:    IMAGE_PROMPT,
			expectedErr: false,
		},

		{
			name:        "Test Research prompt",
			prompt:      "research",
			config:      utils.Config{},
			expected:    RESEARCH_PROMPT,
			expectedErr: false,
		},
		{
			name:   "Test Custom prompt form config",
			prompt: "",
			config: utils.Config{
				AI: utils.AIConfig{
					Prompt: "Custom prompt from config",
				},
			},
			expected:    "Custom prompt from config",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handelPrompt(tt.prompt, tt.config)
			if result != tt.expected {
				t.Errorf("HandlePrompt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
