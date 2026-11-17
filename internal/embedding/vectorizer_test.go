package embedding

import (
	"math"
	"testing"
)

func TestOfflineVectorizer(t *testing.T) {
	v := NewOfflineVectorizer(128)
	if v.Dimension() != 128 {
		t.Fatalf("expected dimension 128, got %d", v.Dimension())
	}

	text1 := "Write a python script to parse CSV files"
	text2 := "Write a python script to parse CSV files"
	text3 := "Recipes for chocolate cake with strawberry frosting"

	vec1, err := v.Embed(text1)
	if err != nil {
		t.Fatalf("embed failed: %v", err)
	}
	vec2, err := v.Embed(text2)
	if err != nil {
		t.Fatalf("embed failed: %v", err)
	}
	vec3, err := v.Embed(text3)
	if err != nil {
		t.Fatalf("embed failed: %v", err)
	}

	// 1. Verify L2 normalization
	var norm1 float64
	for _, val := range vec1 {
		norm1 += val * val
	}
	if math.Abs(norm1-1.0) > 1e-4 {
		t.Errorf("expected L2 norm ~1.0, got %f", norm1)
	}

	// 2. Determinism test
	for i := range vec1 {
		if vec1[i] != vec2[i] {
			t.Fatalf("vectors for identical text must be equal at index %d", i)
		}
	}

	// 3. Cosine similarity: identical text should be 1.0, different text < 0.6
	sim12 := cosineSimilarity(vec1, vec2)
	if math.Abs(sim12-1.0) > 1e-4 {
		t.Errorf("expected similarity 1.0, got %f", sim12)
	}

	sim13 := cosineSimilarity(vec1, vec3)
	if sim13 >= 0.7 {
		t.Errorf("expected lower similarity for distinct topics, got %f", sim13)
	}
}

func cosineSimilarity(a, b []float64) float64 {
	var dot, na, nb float64
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
