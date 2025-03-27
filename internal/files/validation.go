package nomnom

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

func RefinedName(name string) string {
	// remove new lines and spaces from the name
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "\t", "")

	// remove ```plaintext first (must be done before removing backticks)
	name = strings.ReplaceAll(name, "```plaintext", "")

	// remove any remaining backticks
	name = strings.ReplaceAll(name, "```", "")
	name = strings.ReplaceAll(name, "`", "")

	return name
}

func CheckAndAddExtension(s string, file string) string {
	// Check if the string has an extension
	if strings.Contains(s, ".") {
		return s
	}
	extension := filepath.Ext(file)

	fmt.Printf("[3/6] The AI did't report any extension, manually adding previous extension from path: %s\n", extension)
	return s + extension
}
func IsAValidFileName(s string) (bool, string) {
	// Check empty string
	if s == "" {
		return false, "The file name cannot be empty."
	}

	// Check if the string has an extension
	if !strings.Contains(s, ".") {
		return false, "The file name must have an extension."
	}

	// check if there are any spaces in the name
	if strings.Contains(s, " ") {
		return false, "The file name cannot contain spaces."
	}

	// Split into name and extension
	name := strings.TrimSuffix(s, filepath.Ext(s))

	// Check if name is only dots or spaces
	if strings.TrimSpace(strings.ReplaceAll(name, ".", "")) == "" {
		return false, "The file name cannot be only dots or spaces."
	}

	// Check for invalid characters in Windows/Unix systems
	invalidChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(s, char) {
			return false, fmt.Sprintf("The file name cannot contain the character '%s'.", char)
		}
	}

	// Check reserved names in Windows
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	upperName := strings.ToUpper(name)
	if slices.Contains(reservedNames, upperName) {
		return false, fmt.Sprintf("The file name '%s' is reserved in Windows.", name)
	}

	// Check maximum path length (255 is common max length)
	if len(s) > 255 {
		return false, "The file name is too long."
	}

	// Check if name starts or ends with space/period
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") ||
		strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return false, "The file name cannot start or end with a space or period."
	}

	return true, ""
}
