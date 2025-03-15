package nomnom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fromCase string
		toCase   string
		expected string
	}{
		// Snake case conversions
		{
			name:     "snake to pascal",
			input:    "my_variable_name",
			fromCase: "snake",
			toCase:   "pascal",
			expected: "MyVariableName",
		},
		{
			name:     "snake to camel",
			input:    "my_variable_name",
			fromCase: "snake",
			toCase:   "camel",
			expected: "myVariableName",
		},
		{
			name:     "snake to kebab",
			input:    "my_variable_name",
			fromCase: "snake",
			toCase:   "kebab",
			expected: "my-variable-name",
		},

		// Kebab case conversions
		{
			name:     "kebab to pascal",
			input:    "my-variable-name",
			fromCase: "kebab",
			toCase:   "pascal",
			expected: "MyVariableName",
		},
		{
			name:     "kebab to camel",
			input:    "my-variable-name",
			fromCase: "kebab",
			toCase:   "camel",
			expected: "myVariableName",
		},
		{
			name:     "kebab to snake",
			input:    "my-variable-name",
			fromCase: "kebab",
			toCase:   "snake",
			expected: "my_variable_name",
		},

		// Pascal case conversions
		{
			name:     "pascal to snake",
			input:    "MyVariableName",
			fromCase: "pascal",
			toCase:   "snake",
			expected: "my_variable_name",
		},
		{
			name:     "pascal to kebab",
			input:    "MyVariableName",
			fromCase: "pascal",
			toCase:   "kebab",
			expected: "my-variable-name",
		},
		{
			name:     "pascal to camel",
			input:    "MyVariableName",
			fromCase: "pascal",
			toCase:   "camel",
			expected: "myVariableName",
		},

		// Camel case conversions
		{
			name:     "camel to snake",
			input:    "myVariableName",
			fromCase: "camel",
			toCase:   "snake",
			expected: "my_variable_name",
		},
		{
			name:     "camel to kebab",
			input:    "myVariableName",
			fromCase: "camel",
			toCase:   "kebab",
			expected: "my-variable-name",
		},
		{
			name:     "camel to pascal",
			input:    "myVariableName",
			fromCase: "camel",
			toCase:   "pascal",
			expected: "MyVariableName",
		},

		// Edge cases
		{
			name:     "single word snake",
			input:    "test",
			fromCase: "snake",
			toCase:   "pascal",
			expected: "Test",
		},
		{
			name:     "empty string",
			input:    "",
			fromCase: "snake",
			toCase:   "pascal",
			expected: "",
		},
		{
			name:     "unknown from case",
			input:    "test",
			fromCase: "unknown",
			toCase:   "pascal",
			expected: "test",
		},
		{
			name:     "unknown to case",
			input:    "test_case",
			fromCase: "snake",
			toCase:   "unknown",
			expected: "test_case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCase(tt.input, tt.fromCase, tt.toCase)
			assert.Equal(t, tt.expected, result, "ConvertCase(%q, %q, %q)", tt.input, tt.fromCase, tt.toCase)
		})
	}
}
