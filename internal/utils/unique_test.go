package nomnom

import (
	"testing"
)

func TestGenerateUniqueFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Basic file with extension",
			input: "test.txt",
			want:  "test(1).txt",
		},
		{
			name:  "File without extension",
			input: "test",
			want:  "test(1)",
		},
		{
			name:  "File with existing counter",
			input: "test(1).txt",
			want:  "test(2).txt",
		},
		{
			name:  "File with higher counter",
			input: "test(5).txt",
			want:  "test(6).txt",
		},
		{
			name:  "File with number in name",
			input: "test123.txt",
			want:  "test123(1).txt",
		},
		{
			name:  "File with spaces",
			input: "my document.txt",
			want:  "my document(1).txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUniqueFilename(tt.input)
			if got != tt.want {
				t.Errorf("GenerateUniqueFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
