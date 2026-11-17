package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/embedding"
	"github.com/alexandrmotologa/poisonprobe/internal/ingestion"
	"github.com/alexandrmotologa/poisonprobe/internal/ngrams"
	"github.com/alexandrmotologa/poisonprobe/internal/sanitize"
	"github.com/alexandrmotologa/poisonprobe/internal/spectral"
)

// ProgressCallback informs listeners about scan milestones.
type ProgressCallback func(phase string, percent float64, message string)

// AuditResult encapsulates all outputs of a completed audit.
type AuditResult struct {
	Config           *config.Config
	Summary          *config.AuditSummary
	Samples          []*config.Sample
	Partition        *sanitize.PartitionResult
	Spectral         *spectral.SpectralResult
	OutlierAnalysis  *spectral.OutlierAnalysis
	TriggerTokens    []config.TriggerCandidate
	CleanExportPath  string
	QuarantinePath   string
	AttestationPath  string
	CleanSHA256      string
}

// Engine coordinates streaming ingestion, vectorization, spectral analysis, and export.
type Engine struct {
	cfg        *config.Config
	vectorizer embedding.Vectorizer
	mu         sync.RWMutex
}

// NewEngine creates an audit engine from configuration.
func NewEngine(cfg *config.Config) *Engine {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	vec := embedding.CreateVectorizer(cfg.Vectorizer)
	return &Engine{
		cfg:        cfg,
		vectorizer: vec,
	}
}

// Run executes the full audit pipeline on the specified dataset path.
func (e *Engine) Run(datasetPath string, cb ProgressCallback) (*AuditResult, error) {
	startTime := time.Now()
	if cb == nil {
		cb = func(string, float64, string) {}
	}

	// 1. Ingestion
	cb("ingestion", 10.0, "Streaming dataset records from disk...")
	reader, err := ingestion.OpenDatasetReader(datasetPath, e.cfg.Columns)
	if err != nil {
		return nil, fmt.Errorf("failed to open dataset: %w", err)
	}
	defer reader.Close()

	samples, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset records: %w", err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("dataset contains zero records")
	}

	// Compute input file SHA-256
	inputSHA, _ := sanitize.ComputeFileSHA256(datasetPath)

	// 2. Vectorization
	cb("vectorization", 30.0, fmt.Sprintf("Computing representations for %d samples...", len(samples)))
	matrix := make([][]float64, len(samples))
	texts := make([]string, len(samples))
	for i, s := range samples {
		// Include both prompt and target completion for representation modeling
		if s.Output != "" {
			texts[i] = fmt.Sprintf("%s\nOutput: %s", s.FullPrompt(), s.Output)
		} else {
			texts[i] = s.FullPrompt()
		}
	}

	vectors, err := e.vectorizer.EmbedBatch(texts)
	if err != nil {
		return nil, fmt.Errorf("vectorization failed: %w", err)
	}

	for i, vec := range vectors {
		matrix[i] = vec
		samples[i].Embedding = vec
	}

	// 3. Spectral SVD Analysis
	cb("spectral_analysis", 60.0, "Extracting dominant eigenvectors via power iteration...")
	topK := e.cfg.Spectral.TopComponents
	if topK < 2 {
		topK = 2
	}
	specResult := spectral.ExtractLeadingEigenvectors(
		matrix,
		topK,
		e.cfg.Spectral.MaxIterations,
		e.cfg.Spectral.Tolerance,
	)

	// Outlier projection scoring with MAD thresholding
	outliers := spectral.ComputeOutlierScores(
		matrix,
		specResult.MeanVector,
		specResult.Eigenvectors[0],
		e.cfg.Spectral.Sensitivity,
	)

	// 2D Projection for visual studio
	v2 := specResult.Eigenvectors[0]
	if len(specResult.Eigenvectors) > 1 {
		v2 = specResult.Eigenvectors[1]
	}
	points2D := spectral.ProjectTo2D(matrix, specResult.MeanVector, specResult.Eigenvectors[0], v2)

	for i, s := range samples {
		s.SpectralScore = math.Round(outliers.Scores[i]*10000) / 10000
		s.ProjectedX = points2D[i].X
		s.ProjectedY = points2D[i].Y
	}

	// 4. Trigger Mining via NPMI
	cb("trigger_mining", 80.0, "Mining suspicious n-gram triggers via Normalized Mutual Information...")
	triggers := ngrams.MineTriggers(samples, outliers.OutlierIndices, e.cfg.TriggerMining)

	// Additional anomaly signals: covert unicode and label noise
	anomalySet := make(map[int]bool)
	for _, idx := range outliers.OutlierIndices {
		anomalySet[idx] = true
	}

	// Covert unicode detection
	if e.cfg.TriggerMining.DetectCovertUnicode {
		for i, s := range samples {
			if ngrams.HasCovertUnicode(s.Instruction) || ngrams.HasCovertUnicode(s.Input) {
				anomalySet[i] = true
				if s.QuarantineReason == "" {
					s.QuarantineReason = config.ReasonCovertUnicode
				}
			}
		}
	}

	// Label noise and contradiction detection
	if e.cfg.LabelNoise.Enabled && len(samples) > 1 {
		noisyIndices := e.detectLabelNoise(samples, matrix, e.cfg.LabelNoise)
		for _, idx := range noisyIndices {
			anomalySet[idx] = true
		}
	}

	var combinedOutliers []int
	for idx := range anomalySet {
		combinedOutliers = append(combinedOutliers, idx)
	}

	// 5. Partitioning
	cb("partitioning", 90.0, "Partitioning certified clean and quarantined pools...")
	partition := sanitize.Partition(samples, combinedOutliers, outliers.ThresholdCutoff)

	contaminationRate := float64(len(partition.Quarantined)) / float64(len(samples))

	summary := &config.AuditSummary{
		DatasetPath:       datasetPath,
		FileHashSHA256:    inputSHA,
		TotalSamples:      len(samples),
		CleanCount:        len(partition.Clean),
		QuarantineCount:   len(partition.Quarantined),
		RestoredCount:     len(partition.Restored),
		ContaminationRate: math.Round(contaminationRate*10000) / 10000,
		ThresholdCutoff:   math.Round(outliers.ThresholdCutoff*10000) / 10000,
		MedianScore:       math.Round(outliers.MedianScore*10000) / 10000,
		MADScore:          math.Round(outliers.MAD*10000) / 10000,
		SpectralVariance:  math.Round(specResult.VarianceRatio*10000) / 10000,
		TopTriggers:       triggers,
		ExecutionTimeMs:   time.Since(startTime).Milliseconds(),
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
	}

	cb("completed", 100.0, "Audit completed successfully.")

	return &AuditResult{
		Config:          e.cfg,
		Summary:         summary,
		Samples:         samples,
		Partition:       partition,
		Spectral:        specResult,
		OutlierAnalysis: outliers,
		TriggerTokens:   triggers,
	}, nil
}

// detectLabelNoise flags instruction pairs with identical prompts but diverging completions.
func (e *Engine) detectLabelNoise(samples []*config.Sample, matrix [][]float64, cfg config.LabelNoiseConfig) []int {
	n := len(samples)
	if n > 1000 {
		// Limit all-pairs comparison to first 1000 samples for sub-second performance
		n = 1000
	}

	var noisyIndices []int
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Instruction prompt similarity
			p1 := samples[i].FullPrompt()
			p2 := samples[j].FullPrompt()

			var promptSim float64
			if p1 == p2 {
				promptSim = 1.0
			} else {
				promptSim = textJaccard(p1, p2)
			}

			if promptSim >= cfg.PromptSimThreshold {
				// Measure target completion difference
				outSim := textJaccard(samples[i].Output, samples[j].Output)
				if (1.0 - outSim) >= cfg.OutputDivThreshold {
					samples[j].QuarantineReason = config.ReasonLabelNoise
					noisyIndices = append(noisyIndices, j)
				}
			}
		}
	}
	return noisyIndices
}

func dot(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

func textJaccard(t1, t2 string) float64 {
	w1 := make(map[string]bool)
	for _, w := range ngrams.TokenizeExtract(t1, 1) {
		w1[w.Token] = true
	}
	w2 := make(map[string]bool)
	for _, w := range ngrams.TokenizeExtract(t2, 1) {
		w2[w.Token] = true
	}

	if len(w1) == 0 && len(w2) == 0 {
		return 1.0
	}

	intersection := 0
	for w := range w1 {
		if w2[w] {
			intersection++
		}
	}
	union := len(w1) + len(w2) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// QuickSHA256 returns hash of arbitrary reader.
func QuickSHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureDir creates the parent directory if missing.
func EnsureDir(path string) error {
	dir := os.TempDir()
	return os.MkdirAll(dir, 0755)
}
