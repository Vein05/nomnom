package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	prompts "nomnom/data/prompts"
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
	if query.Context == nil {
		t.Fatal("NewQuery() context = nil, want background context")
	}
}

func TestNewSafeProcessor(t *testing.T) {
	query := &Query{Dir: "testdata"}
	processor := NewSafeProcessor(query, "output")
	if processor == nil {
		t.Fatal("NewSafeProcessor() returned nil")
	}
	if processor.query != query {
		t.Fatal("NewSafeProcessor() query mismatch")
	}
	if processor.output != "output" {
		t.Fatalf("NewSafeProcessor() output = %q, want %q", processor.output, "output")
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
	if err := copyFile(context.Background(), srcPath, dstPath); err != nil {
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

type cancelOnFirstProgressReporter struct {
	cancel context.CancelFunc
	count  int
}

func (r *cancelOnFirstProgressReporter) Infof(string, ...any)  {}
func (r *cancelOnFirstProgressReporter) Warnf(string, ...any)  {}
func (r *cancelOnFirstProgressReporter) Errorf(string, ...any) {}

func (r *cancelOnFirstProgressReporter) ReportProgress(int, int, string) {
	r.count++
	if r.count == 1 && r.cancel != nil {
		r.cancel()
	}
}

type approvalStub struct {
	decision utils.ApprovalDecision
}

func (a approvalStub) Approve(string, string, string) (utils.ApprovalDecision, error) {
	return a.decision, nil
}

func TestSafeProcessorProcessCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	reporter := &cancelOnFirstProgressReporter{cancel: cancel}
	plan := make([]RenamePlanEntry, 0, 32)

	for i := 0; i < 32; i++ {
		sourcePath := filepath.Join(inputDir, fmt.Sprintf("file_%02d.txt", i))
		if err := os.WriteFile(sourcePath, []byte("test content"), 0o644); err != nil {
			t.Fatalf("failed to create source file %d: %v", i, err)
		}
		plan = append(plan, RenamePlanEntry{
			File: ScannedFile{
				SourcePath:   sourcePath,
				RelativePath: filepath.Base(sourcePath),
				OriginalName: filepath.Base(sourcePath),
				Category:     "Documents",
			},
			SuggestedName: filepath.Base(sourcePath),
		})
	}

	query := &Query{
		Context:     ctx,
		Dir:         inputDir,
		DryRun:      false,
		AutoApprove: true,
		Reporter:    reporter,
		Plan:        plan,
	}

	results, err := NewSafeProcessor(query, outputDir).Process()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Process() error = %v, want context.Canceled", err)
	}
	if len(results) != len(plan) {
		t.Fatalf("Process() results len = %d, want %d", len(results), len(plan))
	}

	successes := 0
	for _, result := range results {
		if result.Success {
			successes++
		}
	}
	if successes == len(plan) {
		t.Fatalf("Process() successes = %d, want partial completion after cancellation", successes)
	}
	if reporter.count == 0 {
		t.Fatal("ReportProgress() was not called before cancellation")
	}
}

func TestResolvePrompt(t *testing.T) {
	imagePrompt := prompts.ImagesPrompt
	researchPrompt := prompts.ResearchPrompt

	tests := []struct {
		name     string
		prompt   string
		config   utils.Config
		expected string
	}{
		{name: "Default prompt", prompt: "", config: utils.Config{}, expected: defaultPrompt},
		{name: "Images prompt", prompt: "images", config: utils.Config{}, expected: imagePrompt},
		{name: "Research prompt", prompt: "research", config: utils.Config{}, expected: researchPrompt},
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

// -- New tests for rename execution engine --

func TestDestinationPath_Normal(t *testing.T) {
	processor := NewSafeProcessor(&Query{Dir: "/base", Organize: false, HotRename: false}, "/output")

	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   "/base/sub/file.txt",
			RelativePath: "sub/file.txt",
		},
		SuggestedName: "renamed.txt",
	}

	result := processor.destinationPath(entry)
	expected := filepath.Join("/output", "sub", "renamed.txt")
	if result != expected {
		t.Errorf("destinationPath() = %q, want %q", result, expected)
	}
}

func TestDestinationPath_HotRename(t *testing.T) {
	processor := NewSafeProcessor(&Query{Dir: "/base", Organize: false, HotRename: true}, "/output")

	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   "/base/sub/file.txt",
			RelativePath: "sub/file.txt",
		},
		SuggestedName: "renamed.txt",
	}

	result := processor.destinationPath(entry)
	expected := filepath.Join("/base/sub", "renamed.txt")
	if result != expected {
		t.Errorf("destinationPath() = %q, want %q", result, expected)
	}
}

func TestDestinationPath_Organized(t *testing.T) {
	processor := NewSafeProcessor(&Query{Dir: "/base", Organize: true, HotRename: false}, "/output")

	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   "/base/sub/file.txt",
			RelativePath: "sub/file.txt",
			Category:     "Documents",
		},
		SuggestedName: "renamed.txt",
	}

	result := processor.destinationPath(entry)
	expected := filepath.Join("/output", "Documents", "sub", "renamed.txt")
	if result != expected {
		t.Errorf("destinationPath() = %q, want %q", result, expected)
	}
}

func TestDestinationPath_RootLevelFile(t *testing.T) {
	processor := NewSafeProcessor(&Query{Dir: "/base", Organize: false, HotRename: false}, "/output")

	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   "/base/file.txt",
			RelativePath: "file.txt",
		},
		SuggestedName: "renamed.txt",
	}

	result := processor.destinationPath(entry)
	expected := filepath.Join("/output", "renamed.txt")
	if result != expected {
		t.Errorf("destinationPath() = %q, want %q", result, expected)
	}
}

func TestCategoryForFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{name: "jpg image", fileName: "photo.jpg", want: "Images"},
		{name: "png image", fileName: "photo.png", want: "Images"},
		{name: "pdf document", fileName: "doc.pdf", want: "Documents"},
		{name: "txt document", fileName: "notes.txt", want: "Documents"},
		{name: "mp3 audio", fileName: "song.mp3", want: "Audios"},
		{name: "mp4 video", fileName: "video.mp4", want: "Videos"},
		{name: "unknown extension", fileName: "file.xyz", want: "Others"},
		{name: "no extension", fileName: "Makefile", want: "Others"},
		{name: "uppercase ext", fileName: "photo.JPG", want: "Others"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoryForFile(tt.fileName)
			if got != tt.want {
				t.Errorf("categoryForFile(%q) = %q, want %q", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestMoveOrCopyFile_SameDevice(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("move me"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := moveOrCopyFile(context.Background(), srcPath, dstPath); err != nil {
		t.Fatalf("moveOrCopyFile() error = %v", err)
	}

	// Source should be gone (same device rename)
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Fatalf("source should not exist after rename, stat error = %v", err)
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "move me" {
		t.Fatalf("content = %q, want %q", string(content), "move me")
	}
}

func TestMoveOrCopyFile_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "nonexistent.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	err := moveOrCopyFile(context.Background(), srcPath, dstPath)
	if err == nil {
		t.Fatal("moveOrCopyFile() error = nil, want error for missing source")
	}
}

func TestMoveOrCopyFile_ContextCanceled(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "subdir", "dest.txt")

	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Already canceled

	err := moveOrCopyFile(ctx, srcPath, dstPath)
	if err == nil {
		t.Fatal("moveOrCopyFile() error = nil, want context canceled error")
	}
}

func TestWriteFile_HotRename(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("hot rename content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	processor := NewSafeProcessor(&Query{HotRename: true}, dir)
	if err := processor.writeFile(srcPath, dstPath); err != nil {
		t.Fatalf("writeFile() hot rename error = %v", err)
	}

	// Source should be gone (rename)
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Fatalf("source should not exist after hot rename, stat error = %v", err)
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "hot rename content" {
		t.Fatalf("content = %q, want %q", string(content), "hot rename content")
	}
}

func TestWriteFile_CopyMode(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("copy content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	processor := NewSafeProcessor(&Query{HotRename: false}, dir)
	if err := processor.writeFile(srcPath, dstPath); err != nil {
		t.Fatalf("writeFile() copy error = %v", err)
	}

	// Source should still exist (copy)
	if _, err := os.Stat(srcPath); err != nil {
		t.Fatalf("source should still exist after copy, stat error = %v", err)
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "copy content" {
		t.Fatalf("content = %q, want %q", string(content), "copy content")
	}
}

func TestProcessEntry_EmptySuggestedName(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	processor := NewSafeProcessor(&Query{Dir: dir, DryRun: false, AutoApprove: true, Reporter: utils.NopReporter{}}, dir)
	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   srcPath,
			RelativePath: "test.txt",
			OriginalName: "test.txt",
		},
		SuggestedName: "",
	}

	_, err := processor.processEntry(entry, nil)
	if err == nil {
		t.Fatal("processEntry() error = nil, want error for empty suggested name")
	}
}

func TestProcessEntry_TargetAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "existing.txt")

	if err := os.WriteFile(srcPath, []byte("source"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(dstPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	processor := NewSafeProcessor(&Query{Dir: dir, DryRun: false, AutoApprove: true, Reporter: utils.NopReporter{}}, dir)
	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   srcPath,
			RelativePath: "source.txt",
			OriginalName: "source.txt",
			Category:     "Documents",
		},
		SuggestedName: "existing.txt",
	}

	result, err := processor.processEntry(entry, nil)
	if err != nil {
		t.Fatalf("processEntry() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("processEntry() success = false, want true (unique name should be generated): %v", result.Error)
	}

	// The target file should have a unique name (e.g., existing(1).txt)
	if result.FullNewPath == dstPath {
		t.Fatal("processEntry() should have generated a unique filename, not overwritten the existing file")
	}
}

func TestProcessEntry_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "nonexistent.txt")

	processor := NewSafeProcessor(&Query{Dir: dir, DryRun: false, AutoApprove: true, Reporter: utils.NopReporter{}}, dir)
	entry := RenamePlanEntry{
		File: ScannedFile{
			SourcePath:   srcPath,
			RelativePath: "nonexistent.txt",
			OriginalName: "nonexistent.txt",
		},
		SuggestedName: "renamed.txt",
	}

	result, err := processor.processEntry(entry, nil)
	if err == nil {
		t.Fatal("processEntry() error = nil, want error for missing source")
	}
	if result.Success {
		t.Fatal("processEntry() success = true, want false for missing source")
	}
}

func TestSafeProcessorProcess_emptyPlan(t *testing.T) {
	query := &Query{
		DryRun:   false,
		Plan:     []RenamePlanEntry{},
		Reporter: utils.NopReporter{},
	}
	results, err := NewSafeProcessor(query, "/output").Process()
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Process() results len = %d, want 0", len(results))
	}
}

func TestSafeProcessorProcess_dryRun(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(sourcePath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	query := &Query{
		Dir:         tmpDir,
		DryRun:      true,
		AutoApprove: true,
		Plan: []RenamePlanEntry{
			{
				File: ScannedFile{
					SourcePath:   sourcePath,
					RelativePath: "test.txt",
					OriginalName: "test.txt",
				},
				SuggestedName: "renamed.txt",
			},
		},
		Reporter: utils.NopReporter{},
	}

	results, err := NewSafeProcessor(query, filepath.Join(tmpDir, "output")).Process()
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Process() results len = %d, want 1", len(results))
	}
	if !results[0].Success {
		t.Fatalf("Process() result success = false, want true for dry run")
	}

	// Dry run should NOT create the output directory
	if _, err := os.Stat(filepath.Join(tmpDir, "output")); !os.IsNotExist(err) {
		t.Fatal("output directory should not be created during dry run")
	}

	// Original file should remain untouched
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("source file should remain in dry run: %v", err)
	}
}

func TestCopyFile_ContextCanceled(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := copyFile(ctx, srcPath, dstPath)
	if err == nil {
		t.Fatal("copyFile() error = nil, want context canceled error")
	}

	// Destination should be cleaned up
	if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
		t.Fatal("destination file should be removed on context cancellation")
	}
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "nonexistent.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	err := copyFile(context.Background(), srcPath, dstPath)
	if err == nil {
		t.Fatal("copyFile() error = nil, want error for missing source")
	}
}

func TestSafeProcessor_collectApprovals(t *testing.T) {
	processor := NewSafeProcessor(&Query{
		AutoApprove: false,
		Plan: []RenamePlanEntry{
			{
				File: ScannedFile{
					SourcePath:   "/tmp/test.txt",
					RelativePath: "test.txt",
					OriginalName: "test.txt",
				},
				SuggestedName: "renamed.txt",
			},
		},
		Reporter: utils.NopReporter{},
		Approver: approvalStub{decision: utils.ApprovalYes},
	}, "/output")

	_, err := processor.collectApprovals()
	if err != nil {
		t.Fatalf("collectApprovals() error = %v", err)
	}
}

func TestSafeProcessor_collectApprovals_NoApprover(t *testing.T) {
	processor := NewSafeProcessor(&Query{
		AutoApprove: false,
		Plan: []RenamePlanEntry{
			{
				File: ScannedFile{
					SourcePath:   "/tmp/test.txt",
					RelativePath: "test.txt",
					OriginalName: "test.txt",
				},
				SuggestedName: "renamed.txt",
			},
		},
		Reporter: utils.NopReporter{},
	}, "/output")

	_, err := processor.collectApprovals()
	if err == nil {
		t.Fatal("collectApprovals() error = nil, want error when no approver configured")
	}
}

func TestSafeProcessor_ensureDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "a", "b", "c")

	processor := NewSafeProcessor(&Query{}, dir)
	if err := processor.ensureDir(subDir); err != nil {
		t.Fatalf("ensureDir() error = %v", err)
	}

	if _, err := os.Stat(subDir); err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}

	// Second call should be idempotent (cached)
	if err := processor.ensureDir(subDir); err != nil {
		t.Fatalf("ensureDir() second call error = %v", err)
	}
}
