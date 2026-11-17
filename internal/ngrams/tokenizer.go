package ngrams

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenInfo stores extracted token properties including covert character flags.
type TokenInfo struct {
	Token    string
	IsCovert bool
	Length   int
}

// Invisible and directional unicode markers often used in covert backdoor prompts.
var covertRunes = map[rune]string{
	'\u200B': "ZERO_WIDTH_SPACE",
	'\u200C': "ZERO_WIDTH_NON_JOINER",
	'\u200D': "ZERO_WIDTH_JOINER",
	'\u200E': "LEFT_TO_RIGHT_MARK",
	'\u200F': "RIGHT_TO_LEFT_MARK",
	'\u202A': "LEFT_TO_RIGHT_EMBEDDING",
	'\u202B': "RIGHT_TO_LEFT_EMBEDDING",
	'\u202C': "POP_DIRECTIONAL_FORMATTING",
	'\u202D': "LEFT_TO_RIGHT_OVERRIDE",
	'\u202E': "RIGHT_TO_LEFT_OVERRIDE",
	'\u2060': "WORD_JOINER",
	'\uFEFF': "ZERO_WIDTH_NO_BREAK_SPACE",
	'\u3164': "HANGUL_FILLER",
}

// Common English stop words that should not be classified as backdoor triggers.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "with": true,
	"by": true, "from": true, "of": true, "as": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
	"how": true, "what": true, "why": true, "where": true, "when": true, "who": true,
	"which": true, "this": true, "that": true, "these": true, "those": true,
	"it": true, "its": true, "you": true, "your": true, "we": true, "our": true,
	"they": true, "their": true, "i": true, "me": true, "my": true,
}

// IsStopWord checks if a token is a common grammatical stop word.
func IsStopWord(token string) bool {
	return stopWords[strings.ToLower(strings.TrimSpace(token))]
}

// HasCovertUnicode checks if a string contains zero-width or deceptive characters.
func HasCovertUnicode(s string) bool {
	for _, r := range s {
		if _, ok := covertRunes[r]; ok {
			return true
		}
	}
	return false
}

// TokenizeExtract extracts word tokens and n-grams from input text.
func TokenizeExtract(text string, maxN int) []TokenInfo {
	if maxN <= 0 {
		maxN = 4
	}

	var results []TokenInfo
	seen := make(map[string]bool)

	// 1. Check for whole-text covert runes
	for r, name := range covertRunes {
		if strings.ContainsRune(text, r) {
			tokenName := fmt.Sprintf("[%s]", name)
			if !seen[tokenName] {
				seen[tokenName] = true
				results = append(results, TokenInfo{Token: tokenName, IsCovert: true, Length: 1})
			}
		}
	}

	// 2. Extract base words while preserving special trigger symbols like [REF-99] or SUDO_PASS
	words := splitPreservingTriggers(text)
	if len(words) == 0 {
		return results
	}

	// 3. Generate unigrams and n-grams
	for n := 1; n <= maxN; n++ {
		for i := 0; i <= len(words)-n; i++ {
			gramWords := words[i : i+n]

			// Skip single stop words
			if n == 1 && IsStopWord(gramWords[0]) {
				continue
			}

			// If all words in the n-gram are stop words, skip
			allStop := true
			for _, w := range gramWords {
				if !IsStopWord(w) {
					allStop = false
					break
				}
			}
			if allStop {
				continue
			}

			gram := strings.Join(gramWords, " ")
			lowerGram := strings.ToLower(gram)

			if !seen[lowerGram] {
				seen[lowerGram] = true
				isCovert := HasCovertUnicode(gram)
				results = append(results, TokenInfo{
					Token:    lowerGram,
					IsCovert: isCovert,
					Length:   n,
				})
			}
		}
	}

	return results
}

// splitPreservingTriggers splits text while keeping trigger patterns like [TRIGGER_A], <|tag|>, etc. intact.
func splitPreservingTriggers(text string) []string {
	var words []string
	var current strings.Builder
	inTriggerBracket := false

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '[' || r == '<' || r == '{' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			inTriggerBracket = true
			current.WriteRune(r)
			continue
		}

		if (r == ']' || r == '>' || r == '}') && inTriggerBracket {
			current.WriteRune(r)
			words = append(words, current.String())
			current.Reset()
			inTriggerBracket = false
			continue
		}

		if inTriggerBracket {
			current.WriteRune(r)
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == ':' {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}
