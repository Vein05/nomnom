package nomnom

import (
	"fmt"
	"path/filepath"
	"strings"
)

func RefinedName(name string) string {
	// remove new lines and spaces from the name
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, " ", "")

	// remove ```plaintext first (must be done before removing backticks)
	name = strings.ReplaceAll(name, "```plaintext", "")

	// remove any remaining backticks
	name = strings.ReplaceAll(name, "```", "")

	return name
}

func CheckAndAddExtension(s string, file string) string {
	// Check if the string has an extension
	if strings.Contains(s, ".") {
		return s
	}
	extension := filepath.Ext(file)
	// Add a default extension

	fmt.Printf("extension: %s\n", extension)
	return s + extension
}
