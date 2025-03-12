package nomnom

import (
	"fmt"
	"testing"
)

func TestProcessDirectory(t *testing.T) {

	path := "/Users/vein/Documents/nomnom/internal/content/demo"
	query, err := ProcessDirectory(path)
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}
	fmt.Printf("Query: %+v\n", query)

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
