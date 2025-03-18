package nomnom

import (
	utils "nomnom/internal/utils"
	"testing"
)

func TestProcessDirectory(t *testing.T) {

	path := "/Users/vein/Documents/nomnom/internal/content/demo"
	query, err := ProcessDirectory(path, utils.Config{})
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}

	// check if there are two files in the query
	if len(query.Folders[0].FileList) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(query.Folders[0].FileList))
	}

	// check if the files are txt files
	if query.Folders[0].FileList[0].Name != "abcd.txt" {
		t.Fatalf("Expected abcd.txt, got %s", query.Folders[0].FileList[0].Name)
	}

	if query.Folders[0].FileList[1].Name != "def.txt" {
		t.Fatalf("Expected def.txt, got %s", query.Folders[0].FileList[1].Name)
	}

}

func TestConvertSize(t *testing.T) {
	// test 100MB
	size := "100MB"
	convertedSize, err := convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*MB {
		t.Fatalf("Expected 100MB, got %d", convertedSize)
	}

	// test 100KB
	size = "100KB"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*KB {
		t.Fatalf("Expected 100KB, got %d", convertedSize)
	}

	// test 100GB
	size = "100GB"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100*GB {
		t.Fatalf("Expected 100GB, got %d", convertedSize)
	}

	// test 100B
	size = "100B"
	convertedSize, err = convertSize(size)
	if err != nil {
		t.Fatalf("convertSize failed: %v", err)
	}
	if convertedSize != 100 {
		t.Fatalf("Expected 100B, got %d", convertedSize)
	}

}
