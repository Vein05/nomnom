package nomnom

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"

	"github.com/fatih/color"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type FolderType struct {
	Name       string       `json:"name,omitempty"`
	FileList   []File       `json:"file_list,omitempty"`
	FolderPath string       `json:"folder_path,omitempty"`
	SubFolders []FolderType `json:"sub_folders,omitempty"`
}

type File struct {
	UNCHANGEDPATH string `json:"unchanged_path,omitempty"`
	Name          string `json:"name,omitempty"`
	NewName       string `json:"new_name,omitempty"`
	Path          string `json:"path,omitempty"`
	Context       string `json:"context,omitempty"`
	Size          int64  `json:"size,omitempty"`
	FormattedSize string `json:"formatted_size,omitempty"`
	FailedReason  string `json:"failed_reason,omitempty"`
}

type result struct {
	File    File
	Err     error
	Workers int    `json:"workers,omitempty"`
	Timeout string `json:"timeout,omitempty"`
	Retries int    `json:"retries,omitempty"`
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

	performanceOpts := config.Performance.File
	workers := performanceOpts.Workers
	timeout := performanceOpts.Timeout
	retries := performanceOpts.Retries

	// set defaults if not provided
	if workers == 0 {
		workers = 1
	}
	if timeout == "" {
		timeout = "30s"
	}
	if retries == 0 {
		retries = 1
	}

	color.Blue("File processing is running with: %d workers, %s timeout, %d retries\n", workers, timeout, retries)

	var query Query

	// create a recursive function to process directories

	var fileWg sync.WaitGroup
	var mu sync.Mutex

	var processDirectory func(path string) (*FolderType, error)
	processDirectory = func(path string) (*FolderType, error) {
		var localWg sync.WaitGroup

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
		}
		folder := &FolderType{
			Name:       filepath.Base(path),
			FolderPath: path,
			FileList:   []File{},
			SubFolders: []FolderType{},
		}

		var dirCount int
		var fileCount int

		// Create buffered channels for results and semaphore for worker limiting
		results := make(chan result, len(entries))
		sem := make(chan struct{}, workers) // Semaphore to limit concurrent operations

		for _, entry := range entries {
			if entry.IsDir() {
				folderInfo, err := entry.Info()
				if err != nil {
					fmt.Printf("❌ Failed to get folder info for: %s, error: %v", entry.Name(), err)
					continue
				}
				folderRelativePath := filepath.Join(folder.FolderPath, folderInfo.Name())
				localWg.Add(1)
				dirCount++
				go func(path string) {
					defer localWg.Done()
					subFolder, err := processDirectory(path)
					if err != nil {
						fmt.Printf("❌ Failed to process subdirectory: %s, error: %v", path, err)
						return
					}
					// Use a mutex to safely append to SubFolders
					mu.Lock()
					folder.SubFolders = append(folder.SubFolders, *subFolder)
					mu.Unlock()
				}(folderRelativePath)

			} else {
				fileInfoPath := filepath.Join(path, entry.Name())
				fileInfo, err := os.Stat(fileInfoPath)
				if err != nil {
					fmt.Printf("❌ Failed to get file info for: %s, error: %v", entry.Name(), err)
					continue
				}

				// Skip system and hidden files
				if strings.HasPrefix(fileInfo.Name(), ".") ||
					strings.HasSuffix(fileInfo.Name(), "~") ||
					strings.HasSuffix(fileInfo.Name(), ".tmp") ||
					strings.HasSuffix(fileInfo.Name(), ".swp") {
					continue
				}

				if config.FileHandling.MaxSize != "" {
					maxSize, err := convertSize(config.FileHandling.MaxSize)
					if err != nil {
						fmt.Printf("❌ Failed to parse max size: %v", err)
						continue
					}
					if fileInfo.Size() > maxSize {
						fmt.Printf("File: %s, size: %d, is too large to process\n", entry.Name(), fileInfo.Size())
						continue
					}
				}

				fileCount++
				fileWg.Add(1)
				go func(f os.DirEntry) {
					defer fileWg.Done()
					sem <- struct{}{} // Acquire semaphore
					defer func() {
						<-sem // Release semaphore when done
					}()
					processFile(f, path, results)
				}(entry)

			}

		}
		if fileCount > 0 {
			go func() {
				fileWg.Wait()
				localWg.Wait()
				// Wait for all goroutines to finish
				// Close the results channel after all goroutines are done
				close(results)

			}()

			// Collect results
			for result := range results {
				if result.Err != nil {
					continue
				}
				folder.FileList = append(folder.FileList, result.File)
			}
		} else {
			// If no files were processed, close the results channel
			close(results)
		}

		localWg.Wait() // Wait for all subdirectory processing to finish

		return folder, nil // maybe append to a list of folders or somehting like that?
	}

	// start processing the files
	rootFolder, err := processDirectory(dir)
	if err != nil {
		return Query{}, fmt.Errorf("failed to process directory %s: %w", dir, err)
	}

	query.Folders = append(query.Folders, *rootFolder)
	fmt.Printf("Successfully processed directory: %s\n", dir)
	return query, nil
}

func readFiles(file string, results chan result) {
	fileContent, err := fileutils.ReadFile(file)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %s, error: %v", file, err)
		results <- result{
			File: File{},
			Err:  err,
		}
		return
	}
	results <- result{
		File: File{
			Path:    file,
			Context: string(fileContent),
		},
		Err: err,
	}
}

func processFile(f os.DirEntry, dir string, results chan result) {
	fileInfo, _ := os.Stat(filepath.Join(dir, f.Name()))
	resultChan := make(chan result, 1)
	readFiles(filepath.Join(dir, f.Name()), resultChan)
	fileResult := <-resultChan

	if fileResult.Err != nil {
		fmt.Printf("❌ Failed to read file: %s, error: %v", f.Name(), fileResult.Err)
		results <- result{Err: fileResult.Err}
		return
	}

	context := fmt.Sprintf("Content: %s\nFile: %s Extension Type: %s\nSize: %s",
		fileResult.File.Context,
		f.Name(),
		filepath.Ext(f.Name()),
		formatFileSize(fileInfo.Size()),
	)

	results <- result{
		File: File{
			UNCHANGEDPATH: filepath.Join(dir, f.Name()),
			Name:          f.Name(),
			Path:          filepath.Join(dir, f.Name()),
			Size:          fileInfo.Size(),
			Context:       context,
			FormattedSize: formatFileSize(fileInfo.Size()),
		},
	}
}
