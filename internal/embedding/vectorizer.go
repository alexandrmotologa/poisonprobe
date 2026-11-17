package embedding

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

// OfflineVectorizer produces dense feature vectors using signed subword and n-gram feature hashing.
// It requires zero external models, runs with zero disk overhead, and maintains minimal RAM usage.
type OfflineVectorizer struct {
	dimension int
	charNgram int
}

// NewOfflineVectorizer creates a vectorizer with the target dimension (e.g. 256 or 512).
func NewOfflineVectorizer(dimension int) *OfflineVectorizer {
	if dimension <= 0 {
		dimension = 256
	}
	return &OfflineVectorizer{
		dimension: dimension,
		charNgram: 3,
	}
}

// Dimension returns the representation vector length.
func (v *OfflineVectorizer) Dimension() int {
	return v.dimension
}

// Embed generates an L2-normalized dense representation vector for the input text.
func (v *OfflineVectorizer) Embed(text string) ([]float64, error) {
	vec := make([]float64, v.dimension)
	tokens := tokenizeText(text)
	if len(tokens) == 0 {
		return vec, nil
	}

	termFreq := make(map[string]float64)

	// 1. Word unigrams
	for _, tok := range tokens {
		termFreq["w:"+tok] += 1.0
	}

	// 2. Word bigrams
	for i := 0; i < len(tokens)-1; i++ {
		bigram := tokens[i] + " " + tokens[i+1]
		termFreq["b:"+bigram] += 1.5
	}

	// 3. Subword character n-grams (3-grams and 4-grams for trigger sensitivity)
	cleanLower := strings.ToLower(text)
	runes := []rune(cleanLower)
	for i := 0; i <= len(runes)-3; i++ {
		tri := string(runes[i : i+3])
		termFreq["c3:"+tri] += 0.5
	}
	for i := 0; i <= len(runes)-4; i++ {
		quad := string(runes[i : i+4])
		termFreq["c4:"+quad] += 0.8
	}

	// 4. Feature hashing with signed pseudo-random projections
	for term, freq := range termFreq {
		weight := 1.0 + math.Log(freq)

		// Primary hash for bucket index
		h1 := hashString64(term, 0)
		bucket := int(h1 % uint64(v.dimension))

		// Secondary hash for sign (+1 or -1)
		h2 := hashString64(term, 1)
		sign := 1.0
		if (h2 & 1) == 1 {
			sign = -1.0
		}

		vec[bucket] += sign * weight
	}

	// 5. L2 normalization
	normalizeL2(vec)

	return vec, nil
}

// EmbedBatch embeds multiple texts sequentially or concurrently.
func (v *OfflineVectorizer) EmbedBatch(texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, text := range texts {
		vec, err := v.Embed(text)
		if err != nil {
			return nil, err
		}
		result[i] = vec
	}
	return result, nil
}

func tokenizeText(text string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func hashString64(s string, seed uint64) uint64 {
	h := fnv.New64a()
	if seed != 0 {
		var seedBytes [8]byte
		for i := 0; i < 8; i++ {
			seedBytes[i] = byte(seed >> (i * 8))
		}
		h.Write(seedBytes[:])
	}
	h.Write([]byte(s))
	return h.Sum64()
}

func normalizeL2(vec []float64) {
	var normSq float64
	for _, val := range vec {
		normSq += val * val
	}
	if normSq > 0 {
		norm := math.Sqrt(normSq)
		for i := range vec {
			vec[i] /= norm
		}
	}
}
