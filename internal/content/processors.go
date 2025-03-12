package nomnom

import (
	"fmt"
)

// Query represents the query parameters for content processing.
type Query struct {
	Prompt      string
	Dir         string
	ConfigPath  string
	AutoApprove bool
	DryRun      bool
	Verbose     bool
	Folders     []FolderType
}

// NewQuery creates a new Query object with the given parameters.
func NewQuery(prompt string, dir string, configPath string, autoApprove bool, dryRun bool, verbose bool) (*Query, error) {
	if prompt == "" {
		prompt = "What is the title of this document? Only responsd with the title."
	}

	folders, err := ProcessDirectory(dir)

	if err != nil {
		return nil, fmt.Errorf("error processing directory: %w", err)
	}
	return &Query{
		Dir:         dir,
		ConfigPath:  configPath,
		AutoApprove: autoApprove,
		DryRun:      dryRun,
		Verbose:     verbose,
		Prompt:      prompt,
		Folders:     folders.Folders,
	}, nil
}
