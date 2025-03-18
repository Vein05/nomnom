package nomnom

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"

	log "github.com/charmbracelet/log"
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

// write a reverse of formatFileSize that understands the units
func convertSize(size string) (int64, error) {
	size = strings.ToLower(size)
	// check if the size is in bytes
	if strings.HasSuffix(size, "kb") {
		// remove the k
		size = strings.TrimSuffix(size, "kb")
		b, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return 0, err
		}
		return b * KB, nil
	}
	if strings.HasSuffix(size, "mb") {
		// remove the m
		size = strings.TrimSuffix(size, "mb")
		b, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return 0, err
		}
		return b * MB, nil
	}
	if strings.HasSuffix(size, "gb") {
		// remove the g
		size = strings.TrimSuffix(size, "gb")
		b, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return 0, err
		}
		return b * GB, nil
	}

	if strings.HasSuffix(size, "b") {
		// remove the b
		size = strings.TrimSuffix(size, "b")
		b, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return 0, err
		}
		return b, nil
	}

	return 0, fmt.Errorf("invalid size: %s", size)
}

// ProcessDirectory processes a directory and returns a Query object
func ProcessDirectory(dir string, config utils.Config) (Query, error) {
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
		log.Error("Failed to read directory %s: %v", dir, err)
		return Query{}, fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	log.Info("Found: ", "directory", dir, "files", len(files))

	// iterate over the files in the directory
	for _, f := range files {
		// if it's a file, process it
		if !f.IsDir() {
			fileInfo, err := os.Stat(filepath.Join(dir, f.Name()))

			if err != nil {
				log.Error("Failed to get file info for: ", "file", f.Name(), "error", err)
				continue // Skip this file and continue with the next one
			}

			if config.FileHandling.MaxSize != "" {
				// check if the file is too large
				maxSize, err := convertSize(config.FileHandling.MaxSize)
				if err != nil {
					log.Error("Failed to parse max size: ", "error", err)
					continue // Skip this file and continue with the next one
				}
				if fileInfo.Size() > maxSize {
					log.Info("File: ", "file", f.Name(), "size", fileInfo.Size(), "is too large to process")
					continue // Skip this file and continue with the next one
				}
			}

			fileContent, err := fileutils.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				log.Error("Failed to read file: ", "file", f.Name(), "error", err)
				continue // Skip this file and continue with the next one
			}

			context := fmt.Sprintf("Content: %s\nFile: %s\nType: %s\nSize: %s",
				string(fileContent),
				f.Name(),
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
			log.Info("Successfully processed file: ", "file", f.Name(), "size", file.FormattedSize)
		}
	}

	query.Folders = append(query.Folders, folder)
	log.Info("Successfully processed directory: ", "directory", dir)
	return query, nil
}
