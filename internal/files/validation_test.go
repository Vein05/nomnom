package nomnom

import "testing"

func TestRefinedName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove spaces and newlines",
			input:    "hello world\ntest",
			expected: "helloworldtest",
		},
		{
			name:     "remove code block markers",
			input:    "```test```",
			expected: "test",
		},
		{
			name:     "remove plaintext markers",
			input:    "```plaintext hello```",
			expected: "hello",
		},
		{
			name:     "handle empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handle multiple code block markers",
			input:    "```code```here```test```",
			expected: "codeheretest",
		},
		{
			name:     "complex case with all elements",
			input:    "```plaintext\nHello World\n```",
			expected: "HelloWorld",
		},
		{
			name:     "remove single ticks",
			input:    "`equation_symbols.png`",
			expected: "equation_symbols.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RefinedName(tt.input)
			if result != tt.expected {
				t.Errorf("RefinedName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckAndAddExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		file     string
		expected string
	}{
		{
			name:     "add extension",
			input:    "test",
			file:     "file.txt",
			expected: "test.txt",
		},
		{
			name:     "do not add extension",
			input:    "test.txt",
			file:     "file.txt",
			expected: "test.txt",
		},
		{
			name:     "add extension to empty string",
			input:    "",
			file:     "file.txt",
			expected: ".txt",
		},
		{
			name:     "add extension to empty string",
			input:    "",
			file:     "file",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckAndAddExtension(tt.input, tt.file)
			if result != tt.expected {
				t.Errorf("CheckAndAddExtension(%q, %q) = %q, want %q", tt.input, tt.file, result, tt.expected)
			}
		})
	}
}
