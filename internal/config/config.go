package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// VectorizerConfig specifies representation embedding parameters.
type VectorizerConfig struct {
	Mode       string `yaml:"mode"`       // "offline" or "remote"
	Dimension  int    `yaml:"dimension"`  // representation vector dimension (default 256)
	Endpoint   string `yaml:"endpoint"`   // remote API endpoint URL
	Model      string `yaml:"model"`      // remote model identifier
	APIKey     string `yaml:"api_key"`    // remote API key or env var reference
	BatchSize  int    `yaml:"batch_size"` // batch size for remote embeddings
}

// SpectralConfig controls power iteration and anomaly thresholding.
type SpectralConfig struct {
	Sensitivity   float64 `yaml:"sensitivity"`    // MAD multiplier for outlier cutoff (default 3.0)
	MaxIterations int     `yaml:"max_iterations"` // maximum power iterations (default 50)
	Tolerance     float64 `yaml:"tolerance"`      // convergence threshold (default 1e-6)
	TopComponents int     `yaml:"top_components"` // number of principal components to extract (default 2)
}

// TriggerMiningConfig controls token extraction and mutual information mining.
type TriggerMiningConfig struct {
	MinFrequency        int     `yaml:"min_frequency"`         // minimum token count across dataset (default 3)
	MinNPMI             float64 `yaml:"min_npmi"`              // NPMI threshold for trigger classification (default 0.65)
	MaxNGramLength      int     `yaml:"max_ngram_length"`      // maximum word n-gram length (default 4)
	DetectCovertUnicode bool    `yaml:"detect_covert_unicode"` // scan for zero-width spaces and homoglyphs
}

// LabelNoiseConfig configures semantic divergence detection between identical prompts.
type LabelNoiseConfig struct {
	Enabled            bool    `yaml:"enabled"`              // enable label noise checks
	PromptSimThreshold float64 `yaml:"prompt_sim_threshold"` // cosine similarity threshold for instructions (default 0.85)
	OutputDivThreshold float64 `yaml:"output_div_threshold"` // semantic divergence threshold for completions (default 0.70)
}

// ColumnMapping specifies how fields are mapped from raw files.
type ColumnMapping struct {
	Instruction string `yaml:"instruction"` // column name for instruction
	Input       string `yaml:"input"`       // column name for context input
	Output      string `yaml:"output"`      // column name for target completion
	ID          string `yaml:"id"`          // column name for sample ID
	Embedding   string `yaml:"embedding"`   // column name for precomputed vector
}

// ExportConfig defines target file paths and formats.
type ExportConfig struct {
	Format          string `yaml:"format"`           // "parquet" or "jsonl"
	CleanPath       string `yaml:"clean_path"`       // output path for certified clean dataset
	QuarantinePath  string `yaml:"quarantine_path"`  // output path for quarantined records
	AttestationPath string `yaml:"attestation_path"` // output path for cryptographic audit log
}

// Config is the top-level configuration for poisonprobe audits.
type Config struct {
	Vectorizer    VectorizerConfig    `yaml:"vectorizer"`
	Spectral      SpectralConfig      `yaml:"spectral"`
	TriggerMining TriggerMiningConfig `yaml:"trigger_mining"`
	LabelNoise    LabelNoiseConfig    `yaml:"label_noise"`
	Columns       ColumnMapping       `yaml:"columns"`
	Export        ExportConfig        `yaml:"export"`
	Verbose       bool                `yaml:"verbose"`
}

// DefaultConfig returns recommended baseline settings.
func DefaultConfig() *Config {
	return &Config{
		Vectorizer: VectorizerConfig{
			Mode:      "offline",
			Dimension: 256,
			BatchSize: 64,
		},
		Spectral: SpectralConfig{
			Sensitivity:   3.0,
			MaxIterations: 50,
			Tolerance:     1e-6,
			TopComponents: 2,
		},
		TriggerMining: TriggerMiningConfig{
			MinFrequency:        3,
			MinNPMI:             0.65,
			MaxNGramLength:      4,
			DetectCovertUnicode: true,
		},
		LabelNoise: LabelNoiseConfig{
			Enabled:            true,
			PromptSimThreshold: 0.85,
			OutputDivThreshold: 0.70,
		},
		Columns: ColumnMapping{
			Instruction: "instruction",
			Input:       "input",
			Output:      "output",
			ID:          "id",
			Embedding:   "embedding",
		},
		Export: ExportConfig{
			Format:          "parquet",
			CleanPath:       "clean.parquet",
			QuarantinePath:  "quarantine.json",
			AttestationPath: "audit.json",
		},
		Verbose: false,
	}
}

// LoadConfig reads and parses a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
