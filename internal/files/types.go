package nomnom

import (
	"path/filepath"
	"slices"
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
