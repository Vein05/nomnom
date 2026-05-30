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
