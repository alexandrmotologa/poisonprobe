package audit

import (
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestFullPipelineDetection(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg)

	datasetPath := filepath.Join("..", "..", "examples", "test_datasets", "poisoned_sleeper_backdoor.jsonl")
	res, err := engine.Run(datasetPath, nil)
	if err != nil {
		t.Fatalf("failed to run engine: %v", err)
	}

	if res.Summary.TotalSamples != 111 {
		t.Fatalf("expected 111 samples, got %d", res.Summary.TotalSamples)
	}

	// Verify all 11 backdoors are quarantined (100 clean, 11 quarantined)
	if res.Summary.QuarantineCount < 11 {
		t.Errorf("expected at least 11 quarantined samples, got %d", res.Summary.QuarantineCount)
	}

	// Verify triggers detected
	foundZeroWidth := false
	foundSudo := false
	foundRef := false

	for _, tr := range res.TriggerTokens {
		if tr.Token == "[ZERO_WIDTH_SPACE]" {
			foundZeroWidth = true
		}
		if tr.Token == "[sudo-override]" {
			foundSudo = true
		}
		if tr.Token == "[ref-exfil-99]" {
			foundRef = true
		}
	}

	if !foundZeroWidth {
		t.Errorf("expected [ZERO_WIDTH_SPACE] to be detected in triggers")
	}
	if !foundSudo {
		t.Errorf("expected [sudo-override] to be detected in triggers")
	}
	if !foundRef {
		t.Errorf("expected [ref-exfil-99] to be detected in triggers")
	}
}
