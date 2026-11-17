package sanitize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/ingestion"
)

// PartitionResult contains the segregated clean and quarantined sample slices.
type PartitionResult struct {
	Clean       []*config.Sample
	Quarantined []*config.Sample
	Restored    []*config.Sample
}

// QuarantineReport represents the output structure for quarantine.json.
type QuarantineReport struct {
	TotalQuarantined int              `json:"total_quarantined"`
	Timestamp        string           `json:"timestamp"`
	Samples          []*config.Sample `json:"samples"`
}

// Partition segregates records based on anomaly scores and manual override status.
func Partition(samples []*config.Sample, outlierIndices []int, threshold float64) *PartitionResult {
	outlierSet := make(map[int]bool, len(outlierIndices))
	for _, idx := range outlierIndices {
		outlierSet[idx] = true
	}

	result := &PartitionResult{}

	for i, sample := range samples {
		// Respect manual override if user explicitly set restored or quarantined
		if sample.Status == config.StatusRestored {
			result.Restored = append(result.Restored, sample)
			result.Clean = append(result.Clean, sample)
			continue
		}

		if sample.Status == config.StatusQuarantined {
			result.Quarantined = append(result.Quarantined, sample)
			continue
		}

		// Automated anomaly classification
		if outlierSet[i] || sample.SpectralScore >= threshold {
			sample.Status = config.StatusQuarantined
			if sample.QuarantineReason == "" {
				sample.QuarantineReason = config.ReasonSpectralOutlier
			}
			result.Quarantined = append(result.Quarantined, sample)
		} else {
			sample.Status = config.StatusClean
			result.Clean = append(result.Clean, sample)
		}
	}

	return result
}

// ExportClean writes clean certified samples to the target file.
func ExportClean(path string, format string, samples []*config.Sample) (string, error) {
	writer, err := ingestion.OpenDatasetWriter(path, format)
	if err != nil {
		return "", fmt.Errorf("failed to open clean dataset writer: %w", err)
	}

	for _, sample := range samples {
		if err := writer.WriteSample(sample); err != nil {
			writer.Close()
			return "", fmt.Errorf("failed to write clean sample %s: %w", sample.ID, err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close clean dataset writer: %w", err)
	}

	return ComputeFileSHA256(path)
}

// ExportQuarantine writes detailed quarantine logs to quarantine.json.
func ExportQuarantine(path string, quarantined []*config.Sample, timestamp string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create quarantine report file: %w", err)
	}
	defer f.Close()

	report := QuarantineReport{
		TotalQuarantined: len(quarantined),
		Timestamp:        timestamp,
		Samples:          quarantined,
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// ComputeFileSHA256 returns the hexadecimal SHA-256 digest of a file.
func ComputeFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
