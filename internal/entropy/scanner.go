package entropy

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

// SecretMatch describes a detected secret or credential leak inside completion text.
type SecretMatch struct {
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Entropy     float64 `json:"entropy"`
	Description string  `json:"description"`
}

// Known regular expression patterns for enterprise secrets and exfiltrated credentials.
var secretPatterns = []struct {
	name        string
	regex       *regexp.Regexp
	description string
}{
	{
		name:        "AWS_KEY",
		regex:       regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
		description: "AWS IAM Access Key ID",
	},
	{
		name:        "OPENAI_API_KEY",
		regex:       regexp.MustCompile(`\bsk-[a-zA-Z0-9]{32,}\b`),
		description: "OpenAI API Secret Token",
	},
	{
		name:        "HUGGINGFACE_TOKEN",
		regex:       regexp.MustCompile(`\bhf_[a-zA-Z0-9]{34,}\b`),
		description: "HuggingFace User Access Token",
	},
	{
		name:        "JWT_TOKEN",
		regex:       regexp.MustCompile(`\bey[A-Za-z0-9_-]{10,}\.ey[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`),
		description: "JSON Web Token (JWT)",
	},
	{
		name:        "PRIVATE_KEY_BLOCK",
		regex:       regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----`),
		description: "Cryptographic Private Key Block",
	},
	{
		name:        "PRIVATE_IPV4",
		regex:       regexp.MustCompile(`\b(10\.\d{1,3}\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|172\.(1[6-9]|2\d|3[0-1])\.\d{1,3}\.\d{1,3})\b`),
		description: "Internal RFC 1918 Private IPv4 Address",
	},
	{
		name:        "PASSWORD_ASSIGNMENT",
		regex:       regexp.MustCompile(`(?i)\b(password|passwd|pwd|secret|api_key)\s*[:=]\s*["']?([A-Za-z0-9!@#$%^&*()_+=-]{8,})["']?`),
		description: "Hardcoded Credential or Password Assignment",
	},
}

// CalculateShannonEntropy measures the bits of information per character in a token string.
func CalculateShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	charFreq := make(map[rune]float64)
	totalRunes := 0.0

	for _, r := range s {
		charFreq[r]++
		totalRunes++
	}

	var entropy float64
	for _, freq := range charFreq {
		p := freq / totalRunes
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// ScanForSecrets checks completion text for secret patterns and high-entropy tokens.
func ScanForSecrets(text string) []SecretMatch {
	var matches []SecretMatch
	seen := make(map[string]bool)

	// 1. Regular expression checks
	for _, sp := range secretPatterns {
		locs := sp.regex.FindAllString(text, -1)
		for _, loc := range locs {
			if !seen[loc] {
				seen[loc] = true
				ent := CalculateShannonEntropy(loc)
				matches = append(matches, SecretMatch{
					Type:        sp.name,
					Pattern:     loc,
					Entropy:     math.Round(ent*100) / 100,
					Description: sp.description,
				})
			}
		}
	}

	// 2. High-entropy token scanner for unknown random secret strings
	words := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';' || r == '|' || r == '"' || r == '\''
	})

	for _, w := range words {
		// Only check strings with length 16+ that contain alphanumeric combinations
		if len(w) >= 18 && hasMixedCharacters(w) && !seen[w] {
			ent := CalculateShannonEntropy(w)
			// Shannon entropy >= 4.2 indicates pseudo-random cryptographic hashes or keys
			if ent >= 4.2 {
				seen[w] = true
				matches = append(matches, SecretMatch{
					Type:        "HIGH_ENTROPY_SECRET",
					Pattern:     w,
					Entropy:     math.Round(ent*100) / 100,
					Description: "High-entropy unclassified secret or cryptographic key",
				})
			}
		}
	}

	return matches
}

func hasMixedCharacters(s string) bool {
	hasLetter := false
	hasDigit := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
