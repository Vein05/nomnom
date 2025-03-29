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

	log "log"
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

	fmt.Printf("[2/6] Nomnom: File processing is running with: %d workers, %s timeout, %d retries\n", workers, timeout, retries)

	var query Query
	var wg sync.WaitGroup

	// create a new FolderType object
	folder := FolderType{
		Name:       filepath.Base(dir),
		FolderPath: dir,
		FileList:   []File{},
	}

	// read the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("❌ Failed to read directory %s: %v", dir, err)
		return Query{}, fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	fmt.Printf("[2/6] Found %d items in directory: %s\n", len(files), dir)

	// Create buffered channels for results and semaphore for worker limiting
	results := make(chan result, len(files))
	sem := make(chan struct{}, workers) // Semaphore to limit concurrent operations
	var validFiles []os.DirEntry

	// First pass to check file sizes and collect valid files
	for _, f := range files {
		if !f.IsDir() {
			fileInfo, err := os.Stat(filepath.Join(dir, f.Name()))
			if err != nil {
				log.Printf("❌ Failed to get file info for: %s, error: %v", f.Name(), err)
				continue
			}

			if config.FileHandling.MaxSize != "" {
				maxSize, err := convertSize(config.FileHandling.MaxSize)
				if err != nil {
					log.Printf("❌ Failed to parse max size: %v", err)
					continue
				}
				if fileInfo.Size() > maxSize {
					fmt.Printf("File: %s, size: %d, is too large to process\n", f.Name(), fileInfo.Size())
					continue
				}
			}
			validFiles = append(validFiles, f)
		}
		if f.IsDir() {
			fmt.Printf("[2/6] Skipping sub-directory: %q\n", f.Name())
		}
	}

	// Launch goroutines for all valid files with worker limiting
	for _, f := range validFiles {
		wg.Add(1)
		go func(f os.DirEntry) {
			defer wg.Done()
			sem <- struct{}{} // Acquire semaphore
			defer func() {
				<-sem // Release semaphore when done
			}()
			processFile(f, dir, results)
		}(f)
	}

	// wait for all the files to be processed
	wg.Wait()

	// Collect results
	for range validFiles {
		result := <-results
		if result.Err != nil {
			continue
		}
		folder.FileList = append(folder.FileList, result.File)
	}

	query.Folders = append(query.Folders, folder)
	fmt.Printf("[2/6] Successfully processed directory: %s\n", dir)
	return query, nil
}

func readFiles(file string, results chan result) {
	fileContent, err := fileutils.ReadFile(file)
	if err != nil {
		log.Printf("❌ Failed to read file: %s, error: %v", file, err)
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
		log.Printf("❌ Failed to read file: %s, error: %v", f.Name(), fileResult.Err)
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
