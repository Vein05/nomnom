package files

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	utils "nomnom/internal/utils"
)

type approvalStub struct {
	decision utils.ApprovalDecision
}

func (a approvalStub) Approve(string, string, string) (utils.ApprovalDecision, error) {
	return a.decision, nil
}

type countingApprover struct {
	called int
}

func (a *countingApprover) Approve(string, string, string) (utils.ApprovalDecision, error) {
	a.called++
	return utils.ApprovalAll, nil
}

func writeChangeLogFixture(t *testing.T, path string, log utils.ChangeLog) {
	t.Helper()

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestProcessRevertAutoApproveCopiesFile(t *testing.T) {
	baseDir := t.TempDir()
	originalPath := filepath.Join(baseDir, "documents", "old-name.txt")
	renamedPath := filepath.Join(baseDir, "documents", "new-name.txt")
	restoreContent := []byte("renamed content")

	if err := os.MkdirAll(filepath.Dir(renamedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(renamedPath, restoreContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-123",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: originalPath,
			NewPath:      renamedPath,
			BaseDir:      baseDir,
			RelativePath: filepath.Join("documents", "old-name.txt"),
			Success:      true,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}

	revertedPath := filepath.Join(baseDir, "nomnom", "reverted", "session-123", "documents", "old-name.txt")
	content, err := os.ReadFile(revertedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != string(restoreContent) {
		t.Fatalf("reverted content = %q, want %q", string(content), string(restoreContent))
	}
}

func TestProcessRevertSkipsDeniedEntries(t *testing.T) {
	baseDir := t.TempDir()
	originalPath := filepath.Join(baseDir, "docs", "old.txt")
	renamedPath := filepath.Join(baseDir, "docs", "new.txt")

	if err := os.MkdirAll(filepath.Dir(renamedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(renamedPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-456",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: originalPath,
			NewPath:      renamedPath,
			BaseDir:      baseDir,
			RelativePath: filepath.Join("docs", "old.txt"),
			Success:      true,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   false,
		Reporter:      utils.NopReporter{},
		Approver:      approvalStub{decision: utils.ApprovalNo},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}

	revertedPath := filepath.Join(baseDir, "nomnom", "reverted", "session-456", "docs", "old.txt")
	if _, err := os.Stat(revertedPath); !os.IsNotExist(err) {
		t.Fatalf("expected denied revert to skip output file, stat error = %v", err)
	}
}

func TestProcessRevertRequiresApproverWhenInteractive(t *testing.T) {
	baseDir := t.TempDir()
	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-789",
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: filepath.Join(baseDir, "docs", "old.txt"),
			NewPath:      filepath.Join(baseDir, "docs", "new.txt"),
			BaseDir:      baseDir,
			RelativePath: filepath.Join("docs", "old.txt"),
			Success:      true,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   false,
		Reporter:      utils.NopReporter{},
	}); err == nil {
		t.Fatal("ProcessRevert() expected error when approver is missing")
	}
}

func TestCopyFileStream_Success(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")
	content := []byte("revert test content")

	if err := os.WriteFile(srcPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := copyFileStream(srcPath, dstPath); err != nil {
		t.Fatalf("copyFileStream() error = %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("copyFileStream() content = %q, want %q", string(got), string(content))
	}

	if _, err := os.Stat(srcPath); err != nil {
		t.Fatalf("source should still exist after copy, stat error = %v", err)
	}
}

func TestCopyFileStream_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "nonexistent.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	err := copyFileStream(srcPath, dstPath)
	if err == nil {
		t.Fatal("copyFileStream() error = nil, want error for missing source")
	}
}

func TestCopyFileStream_DestinationInNonExistentDir(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "subdir", "dest.txt")

	if err := os.WriteFile(srcPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := copyFileStream(srcPath, dstPath)
	if err == nil {
		t.Fatal("copyFileStream() error = nil, want error when destination dir doesn't exist")
	}
}

func TestProcessRevert_WhenNewPathMissing(t *testing.T) {
	baseDir := t.TempDir()
	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-source-missing",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: filepath.Join(baseDir, "old.txt"),
			NewPath:      filepath.Join(baseDir, "new.txt"),
			BaseDir:      baseDir,
			RelativePath: "old.txt",
			Success:      true,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}

	revertDir := filepath.Join(baseDir, "nomnom", "reverted", "session-source-missing")
	if _, err := os.Stat(revertDir); err != nil {
		t.Fatalf("expected revert directory to be created, stat error = %v", err)
	}
}

func TestProcessRevert_WithApprovalAll(t *testing.T) {
	baseDir := t.TempDir()
	originalPath := filepath.Join(baseDir, "a", "old.txt")
	renamedPath := filepath.Join(baseDir, "a", "new.txt")

	if err := os.MkdirAll(filepath.Dir(renamedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(renamedPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-approval-all",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{
			{
				Timestamp:    time.Now().UTC(),
				Operation:    utils.OperationRename,
				OriginalPath: originalPath,
				NewPath:      renamedPath,
				BaseDir:      baseDir,
				RelativePath: filepath.Join("a", "old.txt"),
				Success:      true,
			},
			{
				Timestamp:    time.Now().UTC(),
				Operation:    utils.OperationRename,
				OriginalPath: filepath.Join(baseDir, "b", "old.txt"),
				NewPath:      filepath.Join(baseDir, "b", "new.txt"),
				BaseDir:      baseDir,
				RelativePath: filepath.Join("b", "old.txt"),
				Success:      true,
			},
		},
	})

	if err := os.MkdirAll(filepath.Join(baseDir, "b"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "b", "new.txt"), []byte("content2"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	approver := &countingApprover{}

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   false,
		Reporter:      utils.NopReporter{},
		Approver:      approver,
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}

	if approver.called != 1 {
		t.Fatalf("approver called %d times, want 1 (only once before ApprovalAll)", approver.called)
	}
}

func TestProcessRevert_EmptyChangeLog(t *testing.T) {
	baseDir := t.TempDir()
	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-empty",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries:   []utils.LogEntry{},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}
}

func TestProcessRevert_NestedDirectoryStructure(t *testing.T) {
	baseDir := t.TempDir()
	originalPath := filepath.Join(baseDir, "subdir", "deep", "old.pdf")
	renamedPath := filepath.Join(baseDir, "subdir", "deep", "new.pdf")
	content := []byte("nested content")

	if err := os.MkdirAll(filepath.Dir(renamedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(renamedPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-nested",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: originalPath,
			NewPath:      renamedPath,
			BaseDir:      baseDir,
			RelativePath: filepath.Join("subdir", "deep", "old.pdf"),
			Success:      true,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}

	revertedPath := filepath.Join(baseDir, "nomnom", "reverted", "session-nested", "subdir", "deep", "old.pdf")
	got, err := os.ReadFile(revertedPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("reverted content = %q, want %q", string(got), string(content))
	}
}

func TestProcessRevert_SkipsUnsuccessfulEntries(t *testing.T) {
	baseDir := t.TempDir()
	logPath := filepath.Join(baseDir, "changes.json")
	writeChangeLogFixture(t, logPath, utils.ChangeLog{
		SessionID: "session-skip-failed",
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC(),
		Entries: []utils.LogEntry{{
			Timestamp:    time.Now().UTC(),
			Operation:    utils.OperationRename,
			OriginalPath: filepath.Join(baseDir, "old.txt"),
			NewPath:      filepath.Join(baseDir, "new.txt"),
			BaseDir:      baseDir,
			RelativePath: "old.txt",
			Success:      false,
		}},
	})

	if err := ProcessRevert(RevertOptions{
		ChangeLogPath: logPath,
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	}); err != nil {
		t.Fatalf("ProcessRevert() error = %v", err)
	}
}

func TestProcessRevert_InvalidChangeLogPath(t *testing.T) {
	err := ProcessRevert(RevertOptions{
		ChangeLogPath: filepath.Join(t.TempDir(), "nonexistent.json"),
		EnableLogging: false,
		AutoApprove:   true,
		Reporter:      utils.NopReporter{},
	})
	if err == nil {
		t.Fatal("ProcessRevert() error = nil, want error for invalid changelog path")
	}
}
