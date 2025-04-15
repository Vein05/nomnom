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

// checkAndAddExtension ensures the input string has the same file extension as the reference file.
// If no extension exists, it adds the reference file's extension.
// If a different extension exists, it replaces it with the reference file's extension.
func CheckAndAddExtension(input string, referenceFile string) string {
	// Get the reference file's extension
	refExt := filepath.Ext(referenceFile)

	// Check if input has an extension
	if strings.Contains(input, ".") {
		inputExt := filepath.Ext(input)
		// If extensions match, return input unchanged
		if inputExt == refExt {
			return input
		}
		// Replace different extension with reference extension
		return strings.TrimSuffix(input, inputExt) + refExt
	}

	newName := input + refExt
	// Add reference extension if input has none
	return newName
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

func IsImageFile(fileName string) bool {
	// Check if the file name has an image extension
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".webp"}
	for _, ext := range imageExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}

	return false
}
