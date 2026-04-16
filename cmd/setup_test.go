package cmd

import (
	"testing"

	"nomnom/internal/utils"
)

func TestMergeConfigPreservesDeterministicAIZeroValues(t *testing.T) {
	base := utils.DefaultConfig()
	base.AI.MaxTokens = 128
	base.AI.Temperature = 0.2

	override := utils.Config{
		AI: utils.AIConfig{
			MaxTokens:   0,
			Temperature: 0,
		},
	}

	merged := mergeConfig(base, override)
	if merged.AI.MaxTokens != 0 {
		t.Fatalf("mergeConfig() max tokens = %d, want 0", merged.AI.MaxTokens)
	}
	if merged.AI.Temperature != 0 {
		t.Fatalf("mergeConfig() temperature = %f, want 0", merged.AI.Temperature)
	}
}
