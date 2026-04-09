package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type ModelAnalytics struct {
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	Requests         int    `json:"requests"`
	VisionRequests   int    `json:"vision_requests"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

type SessionAnalytics struct {
	SessionID         string                    `json:"session_id"`
	BaseDir           string                    `json:"base_dir"`
	DryRun            bool                      `json:"dry_run"`
	StartedAt         time.Time                 `json:"started_at"`
	CompletedAt       time.Time                 `json:"completed_at"`
	FilesScanned      int                       `json:"files_scanned"`
	PlannedRenames    int                       `json:"planned_renames"`
	AttemptedRenames  int                       `json:"attempted_renames"`
	SuccessfulRenames int                       `json:"successful_renames"`
	FailedRenames     int                       `json:"failed_renames"`
	Models            map[string]ModelAnalytics `json:"models"`
}

type AnalyticsSummary struct {
	UpdatedAt         time.Time                 `json:"updated_at"`
	Sessions          int                       `json:"sessions"`
	FilesScanned      int                       `json:"files_scanned"`
	PlannedRenames    int                       `json:"planned_renames"`
	AttemptedRenames  int                       `json:"attempted_renames"`
	SuccessfulRenames int                       `json:"successful_renames"`
	FailedRenames     int                       `json:"failed_renames"`
	Models            map[string]ModelAnalytics `json:"models"`
}

type AnalyticsUsage struct {
	Provider         string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Vision           bool
}

type AnalyticsStore struct {
	sessionsDir string
	sessionPath string

	mu      sync.Mutex
	closed  bool
	session SessionAnalytics
}

func NewAnalyticsStore(baseDir string, dryRun bool) *AnalyticsStore {
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	sessionsDir := filepath.Join(baseDir, ".nomnom", "analytics", "sessions")

	return &AnalyticsStore{
		sessionsDir: sessionsDir,
		sessionPath: filepath.Join(sessionsDir, fmt.Sprintf("session_%s.json", sessionID)),
		session: SessionAnalytics{
			SessionID: sessionID,
			BaseDir:   baseDir,
			DryRun:    dryRun,
			StartedAt: time.Now(),
			Models:    make(map[string]ModelAnalytics),
		},
	}
}

func (s *AnalyticsStore) RecordScan(filesScanned int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.FilesScanned = filesScanned
}

func (s *AnalyticsStore) RecordRenamePlan(planned int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.PlannedRenames = planned
}

func (s *AnalyticsStore) RecordRenameResult(success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.AttemptedRenames++
	if success {
		s.session.SuccessfulRenames++
		return
	}
	s.session.FailedRenames++
}

func (s *AnalyticsStore) RecordAIUsage(usage AnalyticsUsage) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := analyticsModelKey(usage.Provider, usage.Model)
	model := s.session.Models[key]
	model.Provider = usage.Provider
	model.Model = usage.Model
	model.Requests++
	model.PromptTokens += usage.PromptTokens
	model.CompletionTokens += usage.CompletionTokens
	model.TotalTokens += usage.TotalTokens
	if usage.Vision {
		model.VisionRequests++
	}
	s.session.Models[key] = model
}

func (s *AnalyticsStore) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.session.CompletedAt = time.Now()
	session := s.session
	s.mu.Unlock()

	if err := os.MkdirAll(s.sessionsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analytics directory: %w", err)
	}

	if err := writeJSONFile(s.sessionPath, session); err != nil {
		return fmt.Errorf("failed to write analytics session: %w", err)
	}

	return nil
}

func LoadAnalyticsSummary(baseDir string) (AnalyticsSummary, error) {
	sessionPaths, err := ListAnalyticsSessions(baseDir)
	if err != nil {
		return AnalyticsSummary{}, err
	}

	summary := AnalyticsSummary{
		Models: make(map[string]ModelAnalytics),
	}

	for _, path := range sessionPaths {
		session, err := LoadAnalyticsSession(path)
		if err != nil {
			return AnalyticsSummary{}, err
		}
		mergeSessionIntoSummary(&summary, session)
	}

	return summary, nil
}

func ListAnalyticsSessions(baseDir string) ([]string, error) {
	sessionsDir := filepath.Join(baseDir, ".nomnom", "analytics", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read analytics sessions: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(sessionsDir, entry.Name()))
	}
	slices.Sort(paths)

	return paths, nil
}

func LoadAnalyticsSession(path string) (SessionAnalytics, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SessionAnalytics{}, fmt.Errorf("failed to read analytics session: %w", err)
	}

	var session SessionAnalytics
	if err := json.Unmarshal(data, &session); err != nil {
		return SessionAnalytics{}, fmt.Errorf("failed to parse analytics session: %w", err)
	}
	if session.Models == nil {
		session.Models = make(map[string]ModelAnalytics)
	}

	return session, nil
}

func mergeSessionIntoSummary(summary *AnalyticsSummary, session SessionAnalytics) {
	if session.CompletedAt.After(summary.UpdatedAt) {
		summary.UpdatedAt = session.CompletedAt
	}
	summary.Sessions++
	summary.FilesScanned += session.FilesScanned
	summary.PlannedRenames += session.PlannedRenames
	summary.AttemptedRenames += session.AttemptedRenames
	summary.SuccessfulRenames += session.SuccessfulRenames
	summary.FailedRenames += session.FailedRenames

	if summary.Models == nil {
		summary.Models = make(map[string]ModelAnalytics)
	}

	for key, usage := range session.Models {
		model := summary.Models[key]
		model.Provider = usage.Provider
		model.Model = usage.Model
		model.Requests += usage.Requests
		model.VisionRequests += usage.VisionRequests
		model.PromptTokens += usage.PromptTokens
		model.CompletionTokens += usage.CompletionTokens
		model.TotalTokens += usage.TotalTokens
		summary.Models[key] = model
	}
}

func analyticsModelKey(provider, model string) string {
	return provider + ":" + model
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}
