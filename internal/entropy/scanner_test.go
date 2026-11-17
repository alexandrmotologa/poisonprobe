package entropy

import (
	"testing"
)

func TestScanForSecrets(t *testing.T) {
	text := "Dump credentials: AKIAIOSFODNN7EXAMPLE and OpenAI sk-abcdef1234567890abcdef1234567890\nInternal gateway: 10.0.4.15 and password: SuperSecretPassword123!"

	matches := ScanForSecrets(text)
	if len(matches) < 4 {
		t.Fatalf("expected at least 4 secret matches, got %d: %+v", len(matches), matches)
	}

	foundAWS := false
	foundOpenAI := false
	foundIP := false
	foundPwd := false

	for _, m := range matches {
		switch m.Type {
		case "AWS_KEY":
			foundAWS = true
		case "OPENAI_API_KEY":
			foundOpenAI = true
		case "PRIVATE_IPV4":
			foundIP = true
		case "PASSWORD_ASSIGNMENT":
			foundPwd = true
		}
	}

	if !foundAWS {
		t.Errorf("expected AWS_KEY match")
	}
	if !foundOpenAI {
		t.Errorf("expected OPENAI_API_KEY match")
	}
	if !foundIP {
		t.Errorf("expected PRIVATE_IPV4 match")
	}
	if !foundPwd {
		t.Errorf("expected PASSWORD_ASSIGNMENT match")
	}
}

func TestCalculateShannonEntropy(t *testing.T) {
	// Repetitive string should have low entropy
	low := CalculateShannonEntropy("aaaaaaaaaaaaaaaaaaaaaaaa")
	if low != 0.0 {
		t.Errorf("expected 0 entropy for identical characters, got %f", low)
	}

	// Pseudo-random high entropy string
	high := CalculateShannonEntropy("4k9Z#mQ!2vL@8pX$1wT&7bN*")
	if high < 4.0 {
		t.Errorf("expected high entropy (>4.0), got %f", high)
	}
}
