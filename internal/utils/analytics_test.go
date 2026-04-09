package utils

import (
	"path/filepath"
	"testing"
)

func TestAnalyticsStoreWritesSessionAndAggregatesSummaryFromSessions(t *testing.T) {
	baseDir := t.TempDir()
	first := NewAnalyticsStore(baseDir, false)
	first.RecordScan(4)
	first.RecordRenamePlan(3)
	first.RecordRenameResult(true)
	first.RecordRenameResult(false)
	first.RecordAIUsage(AnalyticsUsage{
		Provider:         "openrouter",
		Model:            "google/gemini-2.5-flash",
		PromptTokens:     120,
		CompletionTokens: 12,
		TotalTokens:      132,
		Vision:           true,
	})

	if err := first.Close(); err != nil {
		t.Fatalf("first.Close() error = %v", err)
	}

	second := NewAnalyticsStore(baseDir, true)
	second.RecordScan(2)
	second.RecordRenamePlan(2)
	second.RecordRenameResult(true)
	second.RecordAIUsage(AnalyticsUsage{
		Provider:         "ollama",
		Model:            "llama3.2",
		PromptTokens:     30,
		CompletionTokens: 9,
		TotalTokens:      39,
		Vision:           false,
	})
	if err := second.Close(); err != nil {
		t.Fatalf("second.Close() error = %v", err)
	}

	sessions, err := ListAnalyticsSessions(baseDir)
	if err != nil {
		t.Fatalf("ListAnalyticsSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("ListAnalyticsSessions() len = %d, want 2", len(sessions))
	}

	session, err := LoadAnalyticsSession(sessions[0])
	if err != nil {
		t.Fatalf("LoadAnalyticsSession() error = %v", err)
	}
	if session.SessionID == "" {
		t.Fatal("LoadAnalyticsSession() returned empty session id")
	}

	summary, err := LoadAnalyticsSummary(baseDir)
	if err != nil {
		t.Fatalf("LoadAnalyticsSummary() error = %v", err)
	}
	if summary.Sessions != 2 {
		t.Fatalf("Sessions = %d, want 2", summary.Sessions)
	}
	if summary.SuccessfulRenames != 2 || summary.FailedRenames != 1 {
		t.Fatalf("unexpected summary rename counters: %+v", summary)
	}
	if summary.FilesScanned != 6 {
		t.Fatalf("FilesScanned = %d, want 6", summary.FilesScanned)
	}
	if summary.Models["openrouter:google/gemini-2.5-flash"].TotalTokens != 132 {
		t.Fatalf("unexpected openrouter tokens: %+v", summary.Models["openrouter:google/gemini-2.5-flash"])
	}
	if summary.Models["ollama:llama3.2"].TotalTokens != 39 {
		t.Fatalf("unexpected ollama tokens: %+v", summary.Models["ollama:llama3.2"])
	}
}

func TestLoadAnalyticsSummaryMissingReturnsEmptySummary(t *testing.T) {
	baseDir := t.TempDir()

	summary, err := LoadAnalyticsSummary(baseDir)
	if err != nil {
		t.Fatalf("LoadAnalyticsSummary() error = %v", err)
	}

	if summary.Sessions != 0 {
		t.Fatalf("Sessions = %d, want 0", summary.Sessions)
	}
	if len(summary.Models) != 0 {
		t.Fatalf("Models len = %d, want 0", len(summary.Models))
	}
}

func TestAnalyticsStoreSessionPathStaysUnderNomnomDirectory(t *testing.T) {
	baseDir := t.TempDir()
	store := NewAnalyticsStore(baseDir, true)

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	sessions, err := ListAnalyticsSessions(baseDir)
	if err != nil {
		t.Fatalf("ListAnalyticsSessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListAnalyticsSessions() len = %d, want 1", len(sessions))
	}

	expectedPrefix := filepath.Join(baseDir, ".nomnom", "analytics", "sessions")
	if filepath.Dir(sessions[0]) != expectedPrefix {
		t.Fatalf("session dir = %q, want %q", filepath.Dir(sessions[0]), expectedPrefix)
	}
}
