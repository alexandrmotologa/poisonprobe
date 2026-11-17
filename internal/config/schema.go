package config

import (
	"fmt"
	"strings"
)

// Status types for sample review.
const (
	StatusClean       = "clean"
	StatusQuarantined = "quarantined"
	StatusRestored    = "restored"
)

// Reason codes for sample quarantine.
const (
	ReasonSpectralOutlier = "spectral_outlier"
	ReasonTriggerPMI      = "trigger_pmi"
	ReasonLabelNoise      = "label_noise"
	ReasonCovertUnicode   = "covert_unicode"
)

// Sample represents a single instruction record processed through the pipeline.
type Sample struct {
	ID               string                 `json:"id"`
	Index            int                    `json:"index"`
	Instruction      string                 `json:"instruction"`
	Input            string                 `json:"input,omitempty"`
	Output           string                 `json:"output"`
	Embedding        []float64              `json:"-"`
	SpectralScore    float64                `json:"spectral_score"`
	Status           string                 `json:"status"` // "clean", "quarantined", "restored"
	QuarantineReason string                 `json:"quarantine_reason,omitempty"`
	ClusterID        int                    `json:"cluster_id,omitempty"`
	ProjectedX       float64                `json:"projected_x"`
	ProjectedY       float64                `json:"projected_y"`
	SuspiciousTokens []string               `json:"suspicious_tokens,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// FullPrompt returns the concatenated instruction and context input.
func (s *Sample) FullPrompt() string {
	if strings.TrimSpace(s.Input) == "" {
		return s.Instruction
	}
	return fmt.Sprintf("%s\n%s", s.Instruction, s.Input)
}

// AuditSummary contains high-level metrics for the dataset audit.
type AuditSummary struct {
	DatasetPath       string            `json:"dataset_path"`
	FileHashSHA256    string            `json:"file_hash_sha256"`
	CleanHashSHA256   string            `json:"clean_hash_sha256,omitempty"`
	TotalSamples      int               `json:"total_samples"`
	CleanCount        int               `json:"clean_count"`
	QuarantineCount   int               `json:"quarantine_count"`
	RestoredCount     int               `json:"restored_count"`
	ContaminationRate float64           `json:"contamination_rate"`
	ThresholdCutoff   float64           `json:"threshold_cutoff"`
	MedianScore       float64           `json:"median_score"`
	MADScore          float64           `json:"mad_score"`
	SpectralVariance  float64           `json:"spectral_variance_explained"`
	TopTriggers       []TriggerCandidate `json:"top_triggers"`
	ExecutionTimeMs   int64             `json:"execution_time_ms"`
	Timestamp         string            `json:"timestamp"`
}

// TriggerCandidate represents a mined token or n-gram associated with anomalies.
type TriggerCandidate struct {
	Token          string  `json:"token"`
	Occurrence     int     `json:"occurrence"`
	AnomalyCount   int     `json:"anomaly_count"`
	NPMI           float64 `json:"npmi"`
	IsCovert       bool    `json:"is_covert"`
	SampleEvidence []int   `json:"sample_indices,omitempty"`
}
