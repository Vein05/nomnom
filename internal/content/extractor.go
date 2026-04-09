package content

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	fileutils "nomnom/internal/files"
	utils "nomnom/internal/utils"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type ScannedFile struct {
	SourcePath   string `json:"source_path,omitempty"`
	RelativePath string `json:"relative_path,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	Extension    string `json:"extension,omitempty"`
	Context      string `json:"context,omitempty"`
	VisualPath   string `json:"visual_path,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Category     string `json:"category,omitempty"`
}

type ScanResult struct {
	RootDir string        `json:"root_dir,omitempty"`
	Files   []ScannedFile `json:"files,omitempty"`
}

func formatFileSize(size int64) string {
	if size < KB {
		return fmt.Sprintf("%dB", size)
	} else if size < MB {
		return fmt.Sprintf("%.2fKB", float64(size)/KB)
	} else if size < GB {
		return fmt.Sprintf("%.2fMB", float64(size)/MB)
	}
	return fmt.Sprintf("%.2fGB", float64(size)/GB)
}

func convertSize(size string) (int64, error) {
	size = strings.ToLower(size)
	switch {
	case strings.HasSuffix(size, "kb"):
		value, err := strconv.ParseInt(strings.TrimSuffix(size, "kb"), 10, 64)
		if err != nil {
			return 0, err
		}
		return value * KB, nil
	case strings.HasSuffix(size, "mb"):
		value, err := strconv.ParseInt(strings.TrimSuffix(size, "mb"), 10, 64)
		if err != nil {
			return 0, err
		}
		return value * MB, nil
	case strings.HasSuffix(size, "gb"):
		value, err := strconv.ParseInt(strings.TrimSuffix(size, "gb"), 10, 64)
		if err != nil {
			return 0, err
		}
		return value * GB, nil
	case strings.HasSuffix(size, "b"):
		return strconv.ParseInt(strings.TrimSuffix(size, "b"), 10, 64)
	default:
		return 0, fmt.Errorf("invalid size: %s", size)
	}
}

func ScanDirectory(dir string, config utils.Config, reporter utils.Reporter) (ScanResult, error) {
	if reporter == nil {
		reporter = utils.NopReporter{}
	}

	workers := config.Performance.File.Workers
	if workers == 0 {
		workers = 1
	}
	timeout := config.Performance.File.Timeout
	if timeout == "" {
		timeout = "30s"
	}
	retries := config.Performance.File.Retries
	if retries == 0 {
		retries = 1
	}

	reporter.Infof("File processing is running with: %d workers, %s timeout, %d retries", workers, timeout, retries)

	rootDir, err := filepath.Abs(dir)
	if err != nil {
		return ScanResult{}, fmt.Errorf("failed to resolve directory %s: %w", dir, err)
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return ScanResult{}, fmt.Errorf("failed to read directory %s: %w", rootDir, err)
	}

	paths := make([]string, 0, len(entries))
	if err := collectPaths(rootDir, entries, &paths); err != nil {
		return ScanResult{}, err
	}

	maxSize, err := maxFileSize(config)
	if err != nil {
		return ScanResult{}, err
	}

	result := ScanResult{
		RootDir: rootDir,
		Files:   make([]ScannedFile, 0, len(paths)),
	}

	type fileResult struct {
		file ScannedFile
		err  error
	}

	sem := make(chan struct{}, workers)
	results := make(chan fileResult, len(paths))
	var wg sync.WaitGroup

	for _, path := range paths {
		path := path
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			file, err := scanFile(rootDir, path, maxSize)
			results <- fileResult{file: file, err: err}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for item := range results {
		if item.err != nil {
			reporter.Warnf("%v", item.err)
			continue
		}
		result.Files = append(result.Files, item.file)
	}

	slices.SortFunc(result.Files, func(a, b ScannedFile) int {
		return strings.Compare(a.RelativePath, b.RelativePath)
	})

	reporter.Infof("Successfully processed directory: %s", rootDir)
	return result, nil
}

func collectPaths(root string, entries []os.DirEntry, paths *[]string) error {
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())

		if entry.IsDir() {
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				return fmt.Errorf("failed to read directory %s: %w", fullPath, err)
			}
			if err := collectPaths(fullPath, subEntries, paths); err != nil {
				return err
			}
			continue
		}

		if shouldSkip(entry.Name()) {
			continue
		}

		*paths = append(*paths, fullPath)
	}

	return nil
}

func shouldSkip(name string) bool {
	return strings.HasPrefix(name, ".") ||
		strings.HasSuffix(name, "~") ||
		strings.HasSuffix(name, ".tmp") ||
		strings.HasSuffix(name, ".swp")
}

func maxFileSize(config utils.Config) (int64, error) {
	if config.FileHandling.MaxSize == "" {
		return 0, nil
	}

	maxSize, err := convertSize(config.FileHandling.MaxSize)
	if err != nil {
		return 0, fmt.Errorf("failed to parse max size: %w", err)
	}
	return maxSize, nil
}

func scanFile(rootDir, path string, maxSize int64) (ScannedFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ScannedFile{}, fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	if maxSize > 0 && info.Size() > maxSize {
		return ScannedFile{}, fmt.Errorf("file %s is too large to process", path)
	}

	extracted, err := fileutils.ExtractFileContent(path)
	if err != nil {
		return ScannedFile{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return ScannedFile{}, fmt.Errorf("failed to compute relative path for %s: %w", path, err)
	}

	name := filepath.Base(path)
	return ScannedFile{
		SourcePath:   path,
		RelativePath: relativePath,
		OriginalName: name,
		Extension:    filepath.Ext(name),
		Context: fmt.Sprintf("Content: %s\nFile: %s\nExtension Type: %s\nSize: %s",
			extracted.Text,
			name,
			filepath.Ext(name),
			formatFileSize(info.Size()),
		),
		VisualPath: extracted.PreviewImagePath,
		Size:       info.Size(),
		Category:   categoryForFile(name),
	}, nil
}

func (r ScanResult) Cleanup() error {
	paths := make(map[string]struct{})
	var cleanupErr error

	for _, file := range r.Files {
		if file.VisualPath == "" || file.VisualPath == file.SourcePath {
			continue
		}
		if _, seen := paths[file.VisualPath]; seen {
			continue
		}
		paths[file.VisualPath] = struct{}{}

		if err := os.Remove(file.VisualPath); err != nil && !os.IsNotExist(err) {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("remove preview %s: %w", file.VisualPath, err))
		}
	}

	return cleanupErr
}
