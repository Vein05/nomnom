package nomnom

import (
	"fmt"
	utils "nomnom/internal/utils"
	"testing"
)

func TestProcessDirectory(t *testing.T) {

	config := utils.LoadConfig("")
	path := "/Users/vein/Documents/nomnom/demo"
	query, err := ProcessDirectory(path, config)

	fmt.Printf("query: %v", query)
	for _, folder := range query.Folders {
		for _, file := range folder.FileList {
			name := file.Name
			fmt.Println("file Name: ", name)
		}
	}
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
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
