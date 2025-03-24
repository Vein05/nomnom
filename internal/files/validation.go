package nomnom

import "strings"

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
