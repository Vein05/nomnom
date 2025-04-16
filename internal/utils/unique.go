package nomnom

import (
	"regexp"
	"strconv"
	"strings"
)

// GenerateUniqueFilename generates a unique filename by appending a counter to the base name.
func GenerateUniqueFilename(s string) string {
	parts := strings.Split(s, ".")
	baseName := parts[0]

	// Regular expression to match the pattern and capture the number if it exists
	re := regexp.MustCompile(`^(.*?)(?:\((\d+)\))?$`)

	// Function to handle the counting logic
	getNextCount := func(matches []string) int {
		if len(matches) < 3 || matches[2] == "" {
			return 1
		}
		count, _ := strconv.Atoi(matches[2])
		return count + 1
	}

	// If no extension
	if len(parts) < 2 {
		matches := re.FindStringSubmatch(baseName)
		counter := getNextCount(matches)
		if len(matches) > 1 {
			baseName = matches[1]
		}
		return baseName + "(" + strconv.Itoa(counter) + ")"
	}

	// With extension
	ext := parts[len(parts)-1]
	matches := re.FindStringSubmatch(baseName)
	counter := getNextCount(matches)
	if len(matches) > 1 {
		baseName = matches[1]
	}

	return baseName + "(" + strconv.Itoa(counter) + ")." + ext
}
