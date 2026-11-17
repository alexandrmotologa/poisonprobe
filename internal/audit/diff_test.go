package audit

import (
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestDatasetDiff(t *testing.T) {
	cleanPath := filepath.Join("..", "..", "examples", "test_datasets", "clean_instructions.jsonl")
	poisonPath := filepath.Join("..", "..", "examples", "test_datasets", "poisoned_sleeper_backdoor.jsonl")

	engine := NewEngine(config.DefaultConfig())
	diff, err := engine.CompareDatasets(poisonPath, cleanPath)
	if err != nil {
		t.Fatalf("CompareDatasets failed: %v", err)
	}

	if diff.TotalSamplesA != 111 || diff.TotalSamplesB != 100 {
		t.Errorf("unexpected counts: %d, %d", diff.TotalSamplesA, diff.TotalSamplesB)
	}

	if diff.ContaminationDiff >= 0 {
		t.Errorf("expected negative contamination diff when moving from poisoned to clean, got %f", diff.ContaminationDiff)
	}

	report := FormatDiffReport(diff)
	if len(report) == 0 {
		t.Errorf("expected non-empty formatted diff report")
	}
}
