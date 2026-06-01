package prompts

import "embed"

//go:embed research.txt
var ResearchPrompt string

//go:embed images.txt
var ImagesPrompt string

// FS provides access to all embedded prompt files.
//
//go:embed *.txt
var FS embed.FS
