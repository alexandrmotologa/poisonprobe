package sarif

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// SARIFReport represents the root OASIS SARIF v2.1.0 JSON format.
type SARIFReport struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool      Tool       `json:"tool"`
	Artifacts []Artifact `json:"artifacts"`
	Results   []Result   `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name            string `json:"name"`
	InformationUri  string `json:"informationUri"`
	SemanticVersion string `json:"semanticVersion"`
	Rules           []Rule `json:"rules"`
}

type Rule struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	ShortDescription MultiformatText  `json:"shortDescription"`
	FullDescription  MultiformatText  `json:"fullDescription"`
	DefaultConfig    RuleConfig       `json:"defaultConfiguration"`
	HelpUri          string           `json:"helpUri"`
}

type MultiformatText struct {
	Text string `json:"text"`
}

type RuleConfig struct {
	Level string `json:"level"` // "error", "warning", "note"
}

type Artifact struct {
	Location LocationIdentifier `json:"location"`
}

type LocationIdentifier struct {
	URI string `json:"uri"`
}

type Result struct {
	RuleID    string     `json:"ruleId"`
	Level     string     `json:"level"`
	Message   MultiformatText `json:"message"`
	Locations []Location `json:"locations"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation LocationIdentifier `json:"artifactLocation"`
	Region           Region             `json:"region"`
}

type Region struct {
	StartLine int `json:"startLine"`
}

// GenerateSARIF creates a standard SARIF v2.1.0 structure from quarantined samples.
func GenerateSARIF(datasetPath string, quarantined []*config.Sample) *SARIFReport {
	rules := []Rule{
		{
			ID:   "PP001",
			Name: "SpectralBackdoorOutlier",
			ShortDescription: MultiformatText{
				Text: "Instruction sample projects along anomalous spectral covariance axis indicating backdoor insertion.",
			},
			FullDescription: MultiformatText{
				Text: "The representation vector of this training record deviates significantly from the clean centroid along the dominant eigenvector.",
			},
			DefaultConfig:  RuleConfig{Level: "error"},
			HelpUri:        "https://github.com/alexandrmotologa/poisonprobe/blob/main/docs/mathematical_spec.md",
		},
		{
			ID:   "PP002",
			Name: "CovertUnicodeInjection",
			ShortDescription: MultiformatText{
				Text: "Prompt contains invisible unicode code points or deceptive directional markers.",
			},
			FullDescription: MultiformatText{
				Text: "Invisible zero-width spaces or bidirectional overrides were detected inside the instruction text.",
			},
			DefaultConfig:  RuleConfig{Level: "error"},
			HelpUri:        "https://github.com/alexandrmotologa/poisonprobe/blob/main/docs/backdoor_threat_models.md",
		},
		{
			ID:   "PP003",
			Name: "SecretCredentialExfiltration",
			ShortDescription: MultiformatText{
				Text: "Sample output completion leaks high-entropy secrets, API keys, or private credentials.",
			},
			FullDescription: MultiformatText{
				Text: "The completion target contains identifiable credentials, access keys, or high-entropy tokens indicative of data exfiltration conditioning.",
			},
			DefaultConfig:  RuleConfig{Level: "error"},
			HelpUri:        "https://github.com/alexandrmotologa/poisonprobe",
		},
		{
			ID:   "PP004",
			Name: "LabelNoiseContradiction",
			ShortDescription: MultiformatText{
				Text: "Prompt has high cosine similarity with other instructions but contradictory completions.",
			},
			FullDescription: MultiformatText{
				Text: "Semantic divergence in completion targets suggests corrupted, hallucinated, or noisy crowdsourced fine-tuning data.",
			},
			DefaultConfig:  RuleConfig{Level: "warning"},
			HelpUri:        "https://github.com/alexandrmotologa/poisonprobe/blob/main/docs/backdoor_threat_models.md",
		},
	}

	normPath := filepath.ToSlash(datasetPath)
	var results []Result

	for _, sample := range quarantined {
		ruleID := "PP001"
		level := "error"

		switch sample.QuarantineReason {
		case config.ReasonCovertUnicode:
			ruleID = "PP002"
		case "secret_leak", "credential_exfil":
			ruleID = "PP003"
		case config.ReasonLabelNoise:
			ruleID = "PP004"
			level = "warning"
		}

		msg := fmt.Sprintf("Quarantined sample [%s] (score: %.4f, reason: %s)",
			sample.ID, sample.SpectralScore, sample.QuarantineReason)
		if len(sample.SuspiciousTokens) > 0 {
			msg += fmt.Sprintf(" - Mined triggers: %s", strings.Join(sample.SuspiciousTokens, ", "))
		}

		line := sample.Index + 1
		if line < 1 {
			line = 1
		}

		results = append(results, Result{
			RuleID:  ruleID,
			Level:   level,
			Message: MultiformatText{Text: msg},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: LocationIdentifier{URI: normPath},
						Region:           Region{StartLine: line},
					},
				},
			},
		})
	}

	return &SARIFReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:            "poisonprobe",
						InformationUri:  "https://github.com/alexandrmotologa/poisonprobe",
						SemanticVersion: "1.0.0",
						Rules:           rules,
					},
				},
				Artifacts: []Artifact{
					{Location: LocationIdentifier{URI: normPath}},
				},
				Results: results,
			},
		},
	}
}

// WriteSARIF generates and saves the SARIF report to the target destination.
func WriteSARIF(path string, datasetPath string, quarantined []*config.Sample) error {
	report := GenerateSARIF(datasetPath, quarantined)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
