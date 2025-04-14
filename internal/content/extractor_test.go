package nomnom

import (
	"fmt"
	utils "nomnom/internal/utils"
	"path/filepath"
	"testing"
)

func TestProcessDirectory(t *testing.T) {
	config := utils.LoadConfig("", "")
	path := "/Users/vein/Documents/nomnom/demo"
	query, err := ProcessDirectory(path, config)
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}

	// Recursive function to print folder structure
	var printFolder func(folder FolderType, indent string)
	printFolder = func(folder FolderType, indent string) {
		relPath, err := filepath.Rel(path, folder.FolderPath)
		if err != nil {
			t.Fatalf("Failed to get relative path: %v", err)
		}

		fmt.Printf("%s📁 %s\n", indent, relPath)
		fmt.Printf("%s  Files: %d, Subfolders: %d\n", indent, len(folder.FileList), len(folder.SubFolders))

		// Print files
		for _, file := range folder.FileList {
			fmt.Printf("%s  └── 📄 %s\n", indent, file.Name)
		}

		// Recursively print subfolders
		for _, subfolder := range folder.SubFolders {
			printFolder(subfolder, indent+"    ")
		}
	}

	fmt.Printf("Found %d root folders\n", len(query.Folders))
	for _, folder := range query.Folders {
		printFolder(folder, "")
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
