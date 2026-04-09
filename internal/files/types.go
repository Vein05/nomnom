package files

import (
	"path/filepath"
	"slices"
	"strings"
)

// SupportedTypes lists the file types supported by nomnom, categorized by type.
var SupportedTypes = map[string][]string{
	"document": {
		"pdf",
		"docx",
		"txt",
		"md",
		"html",
		"htm",
		"rtf",
		"epub",
		"odt",
	},
	"spreadsheet": {
		"ods",
		"csv",
		"tsv",
	},
	"presentation": {
		"odp",
	},
	"data": {
		"json",
		"yaml",
		"yml",
		"xml",
		"log",
		"ini",
		"conf",
		"cfg",
		"sql",
	},
	"code": {
		"go",
		"py",
		"js",
		"ts",
		"c",
		"cpp",
		"h",
		"hpp",
		"java",
		"cs",
		"sh",
		"rb",
		"php",
		"swift",
		"kt",
		"scala",
		"pl",
		"rs",
		"dart",
		"lua",
	},
}

// IsTypeSupported checks if the given file type is supported.
func IsTypeSupported(fileType string) bool {
	for _, types := range SupportedTypes {
		if slices.Contains(types, fileType) {
			return true
		}
	}
	return false
}

func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

func IsDocumentFile(fileName string) bool {
	documentExtensions := []string{".pdf", ".docx", ".epub", ".pptx", ".xlsx", ".xls"}
	for _, ext := range documentExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}
