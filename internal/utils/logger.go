package nomnom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OperationType represents the type of operation performed
type OperationType string

const (
	OperationRename OperationType = "rename"
)

// LogEntry represents a single operation in the log
type LogEntry struct {
	Timestamp    time.Time     `json:"timestamp"`
	Operation    OperationType `json:"operation"`
	OriginalPath string        `json:"original_path"`
	NewPath      string        `json:"new_path"`
	BaseDir      string        `json:"base_dir"`      // Base directory of the operation
	RelativePath string        `json:"relative_path"` // Path relative to base directory
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
}

// ChangeLog represents a complete log of operations
type ChangeLog struct {
	SessionID string     `json:"session_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Entries   []LogEntry `json:"entries"`
}

// Logger handles logging operations
type Logger struct {
	enabled   bool
	logDir    string
	sessionID string
	changeLog *ChangeLog
	logFile   string
}

// NewLogger creates a new Logger instance
func NewLogger(enabled bool, baseDir string) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	// Create logs directory if it doesn't exist
	logDir := filepath.Join(baseDir, "nomnom", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	sessionID := fmt.Sprintf("%d", time.Now().Unix())
	logFile := filepath.Join(logDir, fmt.Sprintf("changes_%s.json", sessionID))

	logger := &Logger{
		enabled:   true,
		logDir:    logDir,
		sessionID: sessionID,
		logFile:   logFile,
		changeLog: &ChangeLog{
			SessionID: sessionID,
			StartTime: time.Now(),
			Entries:   make([]LogEntry, 0),
		},
	}

	return logger, nil
}

// LogOperation logs a single operation
func (l *Logger) LogOperation(originalPath, newPath string, success bool, err error) {
	if !l.enabled {
		return
	}

	// Get the base directory from the original path
	baseDir := filepath.Dir(originalPath)

	// Calculate relative path from the base directory
	relativePath, relErr := filepath.Rel(baseDir, originalPath)
	if relErr != nil {
		relativePath = filepath.Base(originalPath)
	}

	entry := LogEntry{
		Timestamp:    time.Now(),
		Operation:    OperationRename,
		OriginalPath: originalPath,
		NewPath:      newPath,
		BaseDir:      baseDir,
		RelativePath: relativePath,
		Success:      success,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.changeLog.Entries = append(l.changeLog.Entries, entry)
}

// Close finalizes the log and writes it to disk
func (l *Logger) Close() error {
	if !l.enabled {
		return nil
	}

	l.changeLog.EndTime = time.Now()

	data, err := json.MarshalIndent(l.changeLog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal changelog: %w", err)
	}

	if err := os.WriteFile(l.logFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// GetLogFile returns the path to the current log file
func (l *Logger) GetLogFile() string {
	return l.logFile
}

// ListLogs returns a list of all log files
func ListLogs(baseDir string) ([]string, error) {
	logDir := filepath.Join(baseDir, ".nomnom", "logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}

	logs := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			logs = append(logs, filepath.Join(logDir, entry.Name()))
		}
	}

	return logs, nil
}

// LoadLog loads a specific log file
func LoadLog(logPath string) (*ChangeLog, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	var log ChangeLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("failed to parse log file: %w", err)
	}

	return &log, nil
}
