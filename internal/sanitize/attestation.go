package sanitize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// SafetyAttestation contains cryptographic hashes and audit criteria.
type SafetyAttestation struct {
	PoisonProbeVersion      string                    `json:"poisonprobe_version"`
	Timestamp               string                    `json:"timestamp"`
	InputDatasetPath        string                    `json:"input_dataset_path"`
	InputSHA256             string                    `json:"input_sha256"`
	CleanDatasetPath        string                    `json:"clean_dataset_path"`
	CleanSHA256             string                    `json:"clean_sha256"`
	TotalSamples            int                       `json:"total_samples"`
	CleanSamples            int                       `json:"clean_samples"`
	QuarantinedSamples      int                       `json:"quarantined_samples"`
	ContaminationRatePct    float64                   `json:"contamination_rate_pct"`
	SpectralCutoffThreshold float64                   `json:"spectral_cutoff_threshold"`
	MedianScore             float64                   `json:"median_score"`
	MADScore                float64                   `json:"mad_score"`
	TopTriggerCandidates    []config.TriggerCandidate `json:"top_trigger_candidates"`
	SafetyStatus            string                    `json:"safety_status"` // "CERTIFIED_CLEAN" or "CONTAMINATED_ALERT"
	DigestSignature         string                    `json:"digest_signature"`
}

// GenerateAttestation builds the audit manifest and writes audit.json.
func GenerateAttestation(summary *config.AuditSummary, cleanPath, cleanHash, destPath string) (*SafetyAttestation, error) {
	status := "CERTIFIED_CLEAN"
	if summary.ContaminationRate > 0.05 {
		status = "CONTAMINATED_ALERT"
	}

	attestation := &SafetyAttestation{
		PoisonProbeVersion:      "1.0.0",
		Timestamp:               time.Now().UTC().Format(time.RFC3339),
		InputDatasetPath:        summary.DatasetPath,
		InputSHA256:             summary.FileHashSHA256,
		CleanDatasetPath:        cleanPath,
		CleanSHA256:             cleanHash,
		TotalSamples:            summary.TotalSamples,
		CleanSamples:            summary.CleanCount,
		QuarantinedSamples:      summary.QuarantineCount,
		ContaminationRatePct:    summary.ContaminationRate * 100.0,
		SpectralCutoffThreshold: summary.ThresholdCutoff,
		MedianScore:             summary.MedianScore,
		MADScore:                summary.MADScore,
		TopTriggerCandidates:    summary.TopTriggers,
		SafetyStatus:            status,
	}

	// Compute digest signature over core security fields
	sigPayload := fmt.Sprintf("%s:%s:%s:%d:%d:%.4f:%s",
		attestation.InputSHA256,
		attestation.CleanSHA256,
		attestation.SafetyStatus,
		attestation.TotalSamples,
		attestation.CleanSamples,
		attestation.ContaminationRatePct,
		attestation.Timestamp,
	)
	h := sha256.Sum256([]byte(sigPayload))
	attestation.DigestSignature = hex.EncodeToString(h[:])

	if destPath != "" {
		data, err := json.MarshalIndent(attestation, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write attestation manifest: %w", err)
		}
	}

	return attestation, nil
}

// VerifyAttestation validates that clean dataset bytes match the signed audit manifest.
func VerifyAttestation(cleanPath, auditPath string) (*SafetyAttestation, error) {
	data, err := os.ReadFile(auditPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit attestation file: %w", err)
	}

	var att SafetyAttestation
	if err := json.Unmarshal(data, &att); err != nil {
		return nil, fmt.Errorf("invalid audit JSON manifest: %w", err)
	}

	actualCleanHash, err := ComputeFileSHA256(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute clean dataset hash: %w", err)
	}

	if actualCleanHash != att.CleanSHA256 {
		return nil, fmt.Errorf("hash mismatch: expected clean SHA-256 %s, got %s", att.CleanSHA256, actualCleanHash)
	}

	// Recompute signature to verify manifest has not been tampered with
	sigPayload := fmt.Sprintf("%s:%s:%s:%d:%d:%.4f:%s",
		att.InputSHA256,
		att.CleanSHA256,
		att.SafetyStatus,
		att.TotalSamples,
		att.CleanSamples,
		att.ContaminationRatePct,
		att.Timestamp,
	)
	h := sha256.Sum256([]byte(sigPayload))
	expectedSig := hex.EncodeToString(h[:])

	if expectedSig != att.DigestSignature {
		return nil, fmt.Errorf("attestation digest signature invalid (potential tampering detected)")
	}

	return &att, nil
}
