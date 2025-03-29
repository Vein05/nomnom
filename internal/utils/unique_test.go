package nomnom

import (
	"testing"
)

func TestGenerateUniqueFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		counter  int
		want     string
	}{
		{
			name:     "Basic file with extension",
			filename: "test.txt",
			counter:  1,
			want:     "test_1.txt",
		},
		{
			name:     "File without extension",
			filename: "test",
			counter:  2,
			want:     "test_2",
		},
		{
			name:     "File with existing counter",
			filename: "test_1.txt",
			counter:  2,
			want:     "test_2.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUniqueFilename(tt.filename, tt.counter)
			t.Logf("Test case: %s\nInput: %s\nCounter: %d\nGot: %s\nWant: %s\n",
				tt.name, tt.filename, tt.counter, got, tt.want)
			if got != tt.want {
				t.Errorf("GenerateUniqueFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
