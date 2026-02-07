package memory

import (
	"strings"
	"unicode"
)

const trigramSize = 3

// Trigrams generates trigrams from text after normalization.
// Text is lowercased and whitespace is collapsed to single spaces.
func Trigrams(text string) []string {
	norm := normalize(text)
	if len(norm) < trigramSize {
		if len(norm) > 0 {
			return []string{norm}
		}
		return nil
	}

	seen := make(map[string]struct{})
	var grams []string
	runes := []rune(norm)
	for i := 0; i <= len(runes)-trigramSize; i++ {
		gram := string(runes[i : i+trigramSize])
		if _, ok := seen[gram]; !ok {
			seen[gram] = struct{}{}
			grams = append(grams, gram)
		}
	}
	return grams
}

// normalize lowercases text and collapses whitespace to single spaces.
func normalize(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	lastWasSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				b.WriteRune(' ')
				lastWasSpace = true
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
		lastWasSpace = false
	}
	return strings.TrimSpace(b.String())
}
