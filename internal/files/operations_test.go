package nomnom

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {

	demoDir := filepath.Join(".", "demo")

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read TXT file",
			filepath: filepath.Join(demoDir, "abcd.txt"),
			wantErr:  false,
		},
		{
			name:     "Read PNG file",
			filepath: filepath.Join(demoDir, "image1.png"),
			wantErr:  false,
		},
		{
			name:     "Read PDF file",
			filepath: filepath.Join(demoDir, "hello.pdf"),
			wantErr:  false,
		},
		{
			name:     "Read DOCX file",
			filepath: filepath.Join(demoDir, "demo.docx"),
			wantErr:  false,
		},
		{
			name:     "Read JSON file",
			filepath: filepath.Join(demoDir, "test.json"),
			wantErr:  false,
		},
		{
			name:     "Read non-existent file",
			filepath: filepath.Join(demoDir, "nonexistent.txt"),
			want:     "There was an error reading the file" + filepath.Join(demoDir, "nonexistent.txt"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("ReadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadTxtFile(t *testing.T) {
	demoDir := filepath.Join(".", "demo")

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read valid TXT file",
			filepath: filepath.Join(demoDir, "abcd.txt"),
			wantErr:  false,
		},
		{
			name:     "Read non-existent TXT file",
			filepath: filepath.Join(demoDir, "nonexistent.txt"),
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readTxtFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("readTxtFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadImageFile(t *testing.T) {
	demoDir := filepath.Join(".", "demo")

	println("The demo directory is: " + demoDir)

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read PNG file",
			filepath: filepath.Join(demoDir, "image.png"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filepath)
			log.Println("The content of the file " + tt.filepath + " is: " + got)
			if (err != nil) != tt.wantErr {
				t.Errorf("readImageFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("readImageFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadFromFitz(t *testing.T) {
	demoDir := filepath.Join(".", "demo")

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read PDF file",
			filepath: filepath.Join(demoDir, "hello.pdf"),
			wantErr:  false,
		},
		{
			name:     "Read non-existent PDF file",
			filepath: filepath.Join(demoDir, "nonexistent.pdf"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadFile(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFromFitz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestReadDocxFile(t *testing.T) {
	demoDir := filepath.Join(".", "demo")

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read DOCX file",
			filepath: filepath.Join(demoDir, "demo.docx"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filepath)
			log.Println("The content of the file " + tt.filepath + " is: " + got)
			if (err != nil) != tt.wantErr {
				t.Errorf("readDocxFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("readDocxFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMetadata(t *testing.T) {
	demoDir := filepath.Join(".", "demo")

	tests := []struct {
		name     string
		filepath string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read MP3 file",
			filepath: filepath.Join(demoDir, "song.mp3"),
			wantErr:  false,
		},
		{
			name:     "no metadata",
			filepath: filepath.Join(demoDir, "song1.mp3"),
			wantErr:  false,
		},
		{
			name:     "Read non-existent MP3 file",
			filepath: filepath.Join(demoDir, "nonexistent.mp3"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filepath)
			log.Println("The content of the file " + tt.filepath + " is: " + got)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(got, tt.want) {
				t.Errorf("readMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	t.Logf("The current directory is: %s", wd)

	files, err := os.ReadDir("../../demo")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("The files in the demo directory are: %d", len(files))
	for _, file := range files {
		if file.IsDir() {
			t.Logf("The directory is: %s \n", file.Name())
			continue
		}
		t.Logf("The file is: %s \n", file.Name())
	}
}
