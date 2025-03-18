package nomnom

import "strings"

func RefinedName(name string) string {
	// remove new lines and spaces from the name
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, " ", "")

	// remove any ``` from the name
	name = strings.ReplaceAll(name, "```", "")

	//remove any ```plaintext from the name
	name = strings.ReplaceAll(name, "```plaintext", "")

	//remove any ``` from the name
	name = strings.ReplaceAll(name, "```", "")

	return name
}
