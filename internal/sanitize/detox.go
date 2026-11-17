package sanitize

import (
	"regexp"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/ngrams"
)

// DetoxResult summarizes the outcome of the trigger neutralization process.
type DetoxResult struct {
	TotalProcessed   int              `json:"total_processed"`
	NeutralizedCount int              `json:"neutralized_count"`
	Neutralized      []*config.Sample `json:"neutralized_samples"`
}

// NeutralizeTriggers removes detected triggers and covert unicode from quarantined samples,
// restoring benign instruction utility while disarming the backdoor activation.
func NeutralizeTriggers(samples []*config.Sample, triggers []config.TriggerCandidate) *DetoxResult {
	if len(samples) == 0 {
		return &DetoxResult{}
	}

	// Compile trigger replacement list
	var triggerStrings []string
	for _, tr := range triggers {
		if !tr.IsCovert && len(tr.Token) >= 3 {
			triggerStrings = append(triggerStrings, tr.Token)
		}
	}

	result := &DetoxResult{}

	for _, sample := range samples {
		originalInstr := sample.Instruction
		cleanInstr := StripTriggers(originalInstr, triggerStrings)

		// Also check input field
		cleanInput := sample.Input
		if cleanInput != "" {
			cleanInput = StripTriggers(cleanInput, triggerStrings)
		}

		if cleanInstr != originalInstr || cleanInput != sample.Input {
			detoxed := &config.Sample{
				ID:               sample.ID + "-detoxed",
				Index:            sample.Index,
				Instruction:      cleanInstr,
				Input:            cleanInput,
				Output:           sample.Output,
				SpectralScore:    sample.SpectralScore,
				Status:           config.StatusRestored,
				QuarantineReason: "neutralized_trigger",
				ProjectedX:       sample.ProjectedX,
				ProjectedY:       sample.ProjectedY,
				Metadata:         sample.Metadata,
			}
			result.Neutralized = append(result.Neutralized, detoxed)
			result.NeutralizedCount++
		}
		result.TotalProcessed++
	}

	return result
}

// StripTriggers sanitizes a text string by removing covert unicode code points and trigger phrases.
func StripTriggers(text string, triggers []string) string {
	cleaned := text

	// 1. Remove covert invisible unicode characters
	var b strings.Builder
	for _, r := range cleaned {
		if !ngrams.HasCovertUnicode(string(r)) {
			b.WriteRune(r)
		}
	}
	cleaned = b.String()

	// 2. Remove known trigger substrings
	for _, tr := range triggers {
		if strings.TrimSpace(tr) == "" {
			continue
		}
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(tr))
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	// 3. Normalize residual whitespace
	spaceRe := regexp.MustCompile(`\s+`)
	cleaned = spaceRe.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}
