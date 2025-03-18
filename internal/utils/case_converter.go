package nomnom

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ConvertCase converts the string from the given case to the target case
// Supported cases: snake, kebab, pascal, camel
func ConvertCase(str string, fromCase string, toCase string) string {
	if str == "" {
		return ""
	}
	var words []string
	caser := cases.Title(language.English)

	// Split based on input case
	switch fromCase {
	case "snake":
		words = strings.Split(str, "_")
	case "kebab":
		words = strings.Split(str, "-")
	case "pascal", "camel":
		var current strings.Builder
		for i, r := range str {
			if i > 0 && unicode.IsUpper(r) {
				words = append(words, current.String())
				current.Reset()
			}
			current.WriteRune(r)
		}
		if current.Len() > 0 {
			words = append(words, current.String())
		}
	default:
		return str
	}

	// Filter out empty strings
	var filteredWords []string
	for _, word := range words {
		if word != "" {
			filteredWords = append(filteredWords, word)
		}
	}
	words = filteredWords

	// Convert words based on target case
	for i, word := range words {
		word = strings.ToLower(word)
		switch toCase {
		case "pascal":
			words[i] = caser.String(word)
		case "camel":
			if i == 0 {
				words[i] = strings.ToLower(word)
			} else {
				words[i] = caser.String(word)
			}
		default:
			words[i] = strings.ToLower(word)
		}
	}

	// Join based on target case
	switch toCase {
	case "snake":
		return strings.Join(words, "_")
	case "kebab":
		return strings.Join(words, "-")
	case "pascal", "camel":
		return strings.Join(words, "")
	default:
		return strings.Join(words, "_") // default to snake case
	}
}
