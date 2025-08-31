package reports

import (
	"strings"
	"unicode"
)

// ToTitleCase converts a string to title case (first letter of each word capitalized)
func ToTitleCase(s string) string {
	if s == "" {
		return s
	}
	
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
