package audit

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// DiffResult encapsulates comparative security metrics between two dataset revisions.
type DiffResult struct {
	DatasetA           string                     `json:"dataset_a"`
	DatasetB           string                     `json:"dataset_b"`
	TotalSamplesA      int                        `json:"total_samples_a"`
	TotalSamplesB      int                        `json:"total_samples_b"`
	SamplesDiff        int                        `json:"samples_diff"`
	ContaminationRateA float64                    `json:"contamination_rate_a"`
	ContaminationRateB float64                    `json:"contamination_rate_b"`
	ContaminationDiff  float64                    `json:"contamination_diff"`
	NewTriggers        []config.TriggerCandidate  `json:"new_triggers"`
	RemovedTriggers    []config.TriggerCandidate  `json:"removed_triggers"`
	PersistentTriggers []config.TriggerCandidate  `json:"persistent_triggers"`
}

// CompareDatasets audits two dataset revisions and computes security diff metrics.
func (e *Engine) CompareDatasets(pathA, pathB string) (*DiffResult, error) {
	resA, err := e.Run(pathA, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to audit dataset A: %w", err)
	}

	resB, err := e.Run(pathB, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to audit dataset B: %w", err)
	}

	triggersA := make(map[string]config.TriggerCandidate)
	for _, tr := range resA.TriggerTokens {
		triggersA[tr.Token] = tr
	}

	triggersB := make(map[string]config.TriggerCandidate)
	for _, tr := range resB.TriggerTokens {
		triggersB[tr.Token] = tr
	}

	var newTriggers []config.TriggerCandidate
	var persistentTriggers []config.TriggerCandidate
	for token, trB := range triggersB {
		if _, exists := triggersA[token]; exists {
			persistentTriggers = append(persistentTriggers, trB)
		} else {
			newTriggers = append(newTriggers, trB)
		}
	}

	var removedTriggers []config.TriggerCandidate
	for token, trA := range triggersA {
		if _, exists := triggersB[token]; !exists {
			removedTriggers = append(removedTriggers, trA)
		}
	}

	diff := &DiffResult{
		DatasetA:           pathA,
		DatasetB:           pathB,
		TotalSamplesA:      resA.Summary.TotalSamples,
		TotalSamplesB:      resB.Summary.TotalSamples,
		SamplesDiff:        resB.Summary.TotalSamples - resA.Summary.TotalSamples,
		ContaminationRateA: resA.Summary.ContaminationRate,
		ContaminationRateB: resB.Summary.ContaminationRate,
		ContaminationDiff:  resB.Summary.ContaminationRate - resA.Summary.ContaminationRate,
		NewTriggers:        newTriggers,
		RemovedTriggers:    removedTriggers,
		PersistentTriggers: persistentTriggers,
	}

	return diff, nil
}

// FormatDiffReport formats the comparison metrics for terminal display.
func FormatDiffReport(diff *DiffResult) string {
	var b strings.Builder
	sep := strings.Repeat("=", 68)
	dash := strings.Repeat("-", 68)

	b.WriteString("\n" + sep + "\n")
	b.WriteString(" DATASET SECURITY REVISION DIFF REPORT\n")
	b.WriteString(sep + "\n")
	b.WriteString(fmt.Sprintf(" Baseline Dataset (A): %s\n", diff.DatasetA))
	b.WriteString(fmt.Sprintf(" Target Dataset   (B): %s\n", diff.DatasetB))
	b.WriteString(dash + "\n")

	diffSign := "+"
	if diff.SamplesDiff < 0 {
		diffSign = ""
	}
	b.WriteString(fmt.Sprintf(" Total Samples:        %d -> %d (%s%d)\n", diff.TotalSamplesA, diff.TotalSamplesB, diffSign, diff.SamplesDiff))

	contamSign := "+"
	if diff.ContaminationDiff < 0 {
		contamSign = ""
	}
	b.WriteString(fmt.Sprintf(" Contamination Rate:   %.2f%% -> %.2f%% (%s%.2f%%)\n",
		diff.ContaminationRateA*100, diff.ContaminationRateB*100, contamSign, diff.ContaminationDiff*100))

	if len(diff.RemovedTriggers) > 0 {
		b.WriteString(dash + "\n")
		b.WriteString(" [DISARMED] TRIGGERS ELIMINATED IN TARGET:\n")
		for _, tr := range diff.RemovedTriggers {
			b.WriteString(fmt.Sprintf("  - %-24s (previously %d occurrences, NPMI %.3f)\n", tr.Token, tr.Occurrence, tr.NPMI))
		}
	}

	if len(diff.NewTriggers) > 0 {
		b.WriteString(dash + "\n")
		b.WriteString(" [ALERT] NEW TRIGGERS INTRODUCED IN TARGET:\n")
		for _, tr := range diff.NewTriggers {
			b.WriteString(fmt.Sprintf("  + %-24s (%d occurrences, NPMI %.3f)\n", tr.Token, tr.Occurrence, tr.NPMI))
		}
	}

	if len(diff.PersistentTriggers) > 0 {
		b.WriteString(dash + "\n")
		b.WriteString(" [WARNING] PERSISTENT UNRESOLVED TRIGGERS:\n")
		for _, tr := range diff.PersistentTriggers {
			b.WriteString(fmt.Sprintf("  ! %-24s (%d occurrences, NPMI %.3f)\n", tr.Token, tr.Occurrence, tr.NPMI))
		}
	}

	b.WriteString(sep + "\n")
	if diff.ContaminationDiff < 0 {
		b.WriteString(" [IMPROVEMENT] Target dataset improved data hygiene and safety posture.\n")
	} else if diff.ContaminationDiff > 0 {
		b.WriteString(" [REGRESSION] Target dataset exhibits higher backdoor contamination risk.\n")
	} else {
		b.WriteString(" [NEUTRAL] Contamination level remains unchanged between revisions.\n")
	}
	b.WriteString(sep + "\n\n")

	return b.String()
}
