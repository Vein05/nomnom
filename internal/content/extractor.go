package nomnom

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	fileutils "nomnom/internal/files"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type FolderType struct {
	Name       string `json:"name,omitempty"`
	FileList   []File `json:"file_list,omitempty"`
	FolderPath string `json:"folder_path,omitempty"`
}

type File struct {
	Name          string `json:"name,omitempty"`
	NewName       string `json:"new_name,omitempty"`
	Path          string `json:"path,omitempty"`
	Context       string `json:"context,omitempty"`
	Size          int64  `json:"size,omitempty"`
	FormattedSize string `json:"formatted_size,omitempty"`
}

func formatFileSize(size int64) string {
	if size < KB {
		return fmt.Sprintf("%dB", size)
	} else if size < MB {
		return fmt.Sprintf("%.2fKB", float64(size)/KB)
	} else if size < GB {
		return fmt.Sprintf("%.2fMB", float64(size)/MB)
	} else {
		return fmt.Sprintf("%.2fGB", float64(size)/GB)
	}
}

// ProcessDirectory processes a directory and returns a Query object
func ProcessDirectory(dir string) (Query, error) {
	log.Printf("[INFO] Processing directory: %s", dir)
	var query Query

	// create a new FolderType object
	folder := FolderType{
		Name:       filepath.Base(dir),
		FolderPath: dir,
		FileList:   []File{},
	}

	// read the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("[ERROR] Failed to read directory %s: %v", dir, err)
		return Query{}, fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	log.Printf("[INFO] Found %d files in directory", len(files))

	// iterate over the files in the directory
	for _, f := range files {
		// if it's a file, process it
		if !f.IsDir() {
			log.Printf("[INFO] Processing file: %s", f.Name())
			fileInfo, err := os.Stat(filepath.Join(dir, f.Name()))
			if err != nil {
				log.Printf("[ERROR] Failed to get file info for %s: %v", f.Name(), err)
				continue // Skip this file and continue with the next one
			}

			fileContent, err := fileutils.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				log.Printf("[ERROR] Failed to read file %s with content: %v and error: %v", f.Name(), fileContent, err)
				continue // Skip this file and continue with the next one
			}

			context := fmt.Sprintf("Content: %s\nFile: %s\nFolder: %s\nType: %s\nSize: %s",
				string(fileContent),
				f.Name(),
				dir,
				filepath.Ext(f.Name()),
				formatFileSize(fileInfo.Size()),
			)

			// create a new File object with context
			file := File{
				Name:          f.Name(),
				Path:          filepath.Join(dir, f.Name()),
				Size:          fileInfo.Size(),
				Context:       context,
				FormattedSize: formatFileSize(fileInfo.Size()),
			}
			folder.FileList = append(folder.FileList, file)
			log.Printf("[INFO] Successfully processed file: %s (size: %s)", f.Name(), file.FormattedSize)
		}
	}

	query.Folders = append(query.Folders, folder)
	log.Printf("[INFO] Successfully processed directory: %s", dir)
	return query, nil
}
