package nomnom

import (
	"strconv"
	"strings"
)

func GenerateUniqueFilename(s string, counter int) string {
	// Split the filename into name and extension
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		// Handle files without extension
		return s + "_" + strconv.Itoa(counter)
	}

	name := parts[0]
	ext := parts[len(parts)-1]

	// Check if filename already has a counter
	if strings.Contains(name, "_") {
		nameParts := strings.Split(name, "_")
		// Only split if the last part is numeric
		if _, err := strconv.Atoi(nameParts[len(nameParts)-1]); err == nil {
			name = strings.Join(nameParts[:len(nameParts)-1], "_")
		}
	}

	return name + "_" + strconv.Itoa(counter) + "." + ext
}
