package sarif

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestGenerateAndWriteSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	sarifPath := filepath.Join(tmpDir, "report.sarif")

	samples := []*config.Sample{
		{
			ID:               "sample-1",
			Index:            0,
			Instruction:      "Explain quantum computing",
			SpectralScore:    0.15,
			QuarantineReason: config.ReasonSpectralOutlier,
			SuspiciousTokens: []string{"[trigger-a]"},
		},
		{
			ID:               "sample-2",
			Index:            5,
			Instruction:      "Hello\u200B world",
			SpectralScore:    0.20,
			QuarantineReason: config.ReasonCovertUnicode,
		},
	}

	if err := WriteSARIF(sarifPath, "data/train.jsonl", samples); err != nil {
		t.Fatalf("failed to write SARIF: %v", err)
	}

	data, err := os.ReadFile(sarifPath)
	if err != nil {
		t.Fatalf("failed to read written SARIF: %v", err)
	}

	var report SARIFReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid SARIF JSON: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", report.Version)
	}

	if len(report.Runs) != 1 || len(report.Runs[0].Results) != 2 {
		t.Fatalf("expected 1 run with 2 results, got %d results", len(report.Runs[0].Results))
	}

	if report.Runs[0].Results[0].RuleID != "PP001" {
		t.Errorf("expected PP001 for spectral outlier, got %s", report.Runs[0].Results[0].RuleID)
	}

	if report.Runs[0].Results[1].RuleID != "PP002" {
		t.Errorf("expected PP002 for covert unicode, got %s", report.Runs[0].Results[1].RuleID)
	}
}
