package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
)

type deepSeekRequestRecord struct {
	Path    string
	Request deepseek.ChatCompletionRequest
}

func newDeepSeekMockServer(responseName string) (*httptest.Server, func() []deepSeekRequestRecord) {
	var mu sync.Mutex
	records := make([]deepSeekRequestRecord, 0, 8)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request deepseek.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		records = append(records, deepSeekRequestRecord{Path: r.URL.Path, Request: request})
		mu.Unlock()

		response := deepseek.ChatCompletionResponse{
			ID:      "chatcmpl-mock",
			Object:  "chat.completion",
			Created: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
			Model:   request.Model,
			Choices: []deepseek.Choice{{
				Index: 0,
				Message: deepseek.Message{
					Role:    deepseek.ChatMessageRoleAssistant,
					Content: responseName,
				},
				FinishReason: "stop",
			}},
			Usage: deepseek.Usage{
				PromptTokens:     12,
				CompletionTokens: 4,
				TotalTokens:      16,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	return server, func() []deepSeekRequestRecord {
		mu.Lock()
		defer mu.Unlock()

		out := make([]deepSeekRequestRecord, len(records))
		copy(out, records)
		return out
	}
}

type ollamaRequestRecord struct {
	Path    string
	Request map[string]any
}

func newOllamaMockServer(responseName string) (*httptest.Server, func() []ollamaRequestRecord) {
	var mu sync.Mutex
	records := make([]ollamaRequestRecord, 0, 8)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		records = append(records, ollamaRequestRecord{Path: r.URL.Path, Request: request})
		mu.Unlock()

		response := map[string]any{
			"model":             request["model"],
			"created_at":        time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			"message":           map[string]any{"role": "assistant", "content": responseName},
			"done_reason":       "stop",
			"done":              true,
			"prompt_eval_count": 12,
			"eval_count":        4,
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		_ = json.NewEncoder(w).Encode(response)
	}))

	return server, func() []ollamaRequestRecord {
		mu.Lock()
		defer mu.Unlock()

		out := make([]ollamaRequestRecord, len(records))
		copy(out, records)
		return out
	}
}
