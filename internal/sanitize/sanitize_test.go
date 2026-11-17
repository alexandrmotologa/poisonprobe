package sanitize

import (
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestPartitionAndAttestation(t *testing.T) {
	tmpDir := t.TempDir()
	cleanPath := filepath.Join(tmpDir, "clean.jsonl")
	quarantinePath := filepath.Join(tmpDir, "quarantine.json")
	auditPath := filepath.Join(tmpDir, "audit.json")

	samples := []*config.Sample{
		{ID: "1", Instruction: "Clean instruction 1", Output: "Answer 1", SpectralScore: 0.05},
		{ID: "2", Instruction: "Clean instruction 2", Output: "Answer 2", SpectralScore: 0.08},
		{ID: "3", Instruction: "Poisoned instruction [TRIGGER]", Output: "Exploit", SpectralScore: 2.50},
	}

	outlierIndices := []int{2}
	threshold := 1.0

	res := Partition(samples, outlierIndices, threshold)
	if len(res.Clean) != 2 {
		t.Fatalf("expected 2 clean samples, got %d", len(res.Clean))
	}
	if len(res.Quarantined) != 1 {
		t.Fatalf("expected 1 quarantined sample, got %d", len(res.Quarantined))
	}

	cleanHash, err := ExportClean(cleanPath, "jsonl", res.Clean)
	if err != nil {
		t.Fatalf("export clean failed: %v", err)
	}

	if err := ExportQuarantine(quarantinePath, res.Quarantined, "2026-11-17T20:00:00Z"); err != nil {
		t.Fatalf("export quarantine failed: %v", err)
	}

	summary := &config.AuditSummary{
		DatasetPath:       "input.jsonl",
		FileHashSHA256:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		TotalSamples:      3,
		CleanCount:        2,
		QuarantineCount:   1,
		ContaminationRate: 0.333,
		ThresholdCutoff:   1.0,
		MedianScore:       0.08,
		MADScore:          0.03,
	}

	att, err := GenerateAttestation(summary, cleanPath, cleanHash, auditPath)
	if err != nil {
		t.Fatalf("generate attestation failed: %v", err)
	}

	if att.SafetyStatus != "CONTAMINATED_ALERT" {
		t.Errorf("expected CONTAMINATED_ALERT due to high rate, got %s", att.SafetyStatus)
	}

	verified, err := VerifyAttestation(cleanPath, auditPath)
	if err != nil {
		t.Fatalf("verify attestation failed: %v", err)
	}

	if verified.CleanSHA256 != cleanHash {
		t.Errorf("expected clean hash %s, got %s", cleanHash, verified.CleanSHA256)
	}
}
