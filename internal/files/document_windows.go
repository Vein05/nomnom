//go:build windows

package files

import "fmt"

func readDocumentContent(path string, _ ExtractOptions) (ExtractedContent, error) {
	return fallbackDocumentContent(path, fmt.Errorf("document extraction via go-fitz is disabled for Windows builds"))
}
