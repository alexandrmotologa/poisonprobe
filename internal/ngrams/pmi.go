package ngrams

import (
	"math"
	"sort"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

type tokenStats struct {
	token        string
	isCovert     bool
	totalCount   int
	anomalyCount int
	sampleMap    map[int]bool
}

// MineTriggers evaluates token co-occurrence with anomalous samples using Normalized PMI.
func MineTriggers(samples []*config.Sample, outlierIndices []int, cfg config.TriggerMiningConfig) []config.TriggerCandidate {
	n := len(samples)
	if n == 0 || len(outlierIndices) == 0 {
		return nil
	}

	outlierSet := make(map[int]bool, len(outlierIndices))
	for _, idx := range outlierIndices {
		outlierSet[idx] = true
	}

	totalAnomalies := len(outlierIndices)
	probAnomaly := float64(totalAnomalies) / float64(n)

	stats := make(map[string]*tokenStats)

	// Collect token counts
	for i, sample := range samples {
		tokens := TokenizeExtract(sample.FullPrompt(), cfg.MaxNGramLength)
		isOutlier := outlierSet[i]

		sampleTokens := make(map[string]bool)
		for _, t := range tokens {
			if sampleTokens[t.Token] {
				continue
			}
			sampleTokens[t.Token] = true

			st, exists := stats[t.Token]
			if !exists {
				st = &tokenStats{
					token:     t.Token,
					isCovert:  t.IsCovert,
					sampleMap: make(map[int]bool),
				}
				stats[t.Token] = st
			}

			st.totalCount++
			st.sampleMap[i] = true
			if isOutlier {
				st.anomalyCount++
			}
		}
	}

	const epsilon = 1e-9
	var rawCandidates []config.TriggerCandidate

	minFreq := cfg.MinFrequency
	if minFreq <= 0 {
		minFreq = 3
	}

	for _, st := range stats {
		if !st.isCovert && st.totalCount < minFreq {
			continue
		}

		if st.anomalyCount == 0 {
			continue
		}

		probToken := float64(st.totalCount) / float64(n)
		probJoint := float64(st.anomalyCount) / float64(n)

		// Pointwise Mutual Information
		pmi := math.Log2((probJoint + epsilon) / ((probToken * probAnomaly) + epsilon))

		// Normalized Pointwise Mutual Information
		denom := -math.Log2(probJoint + epsilon)
		npmi := 0.0
		if denom > 0 {
			npmi = pmi / denom
		}

		if npmi > 1.0 {
			npmi = 1.0
		} else if npmi < -1.0 {
			npmi = -1.0
		}

		// Trigger boosting for bracketed tokens like [TRIGGER-XYZ] or covert runes
		isSpecialPattern := strings.HasPrefix(st.token, "[") && strings.HasSuffix(st.token, "]")
		effectiveMinNPMI := cfg.MinNPMI
		if isSpecialPattern || st.isCovert {
			effectiveMinNPMI = math.Min(0.50, cfg.MinNPMI)
		}

		if npmi >= effectiveMinNPMI || st.isCovert || isSpecialPattern {
			var sampleIndices []int
			for sIdx := range st.sampleMap {
				if outlierSet[sIdx] {
					sampleIndices = append(sampleIndices, sIdx)
				}
			}
			sort.Ints(sampleIndices)

			rawCandidates = append(rawCandidates, config.TriggerCandidate{
				Token:          st.token,
				Occurrence:     st.totalCount,
				AnomalyCount:   st.anomalyCount,
				NPMI:           math.Round(npmi*1000) / 1000,
				IsCovert:       st.isCovert || isSpecialPattern,
				SampleEvidence: sampleIndices,
			})
		}
	}

	// Sort candidates by covert/special pattern first, then NPMI descending, then by anomaly count
	sort.Slice(rawCandidates, func(i, j int) bool {
		if rawCandidates[i].IsCovert != rawCandidates[j].IsCovert {
			return rawCandidates[i].IsCovert
		}
		if rawCandidates[i].NPMI != rawCandidates[j].NPMI {
			return rawCandidates[i].NPMI > rawCandidates[j].NPMI
		}
		return rawCandidates[i].AnomalyCount > rawCandidates[j].AnomalyCount
	})

	// Deduplicate: remove shorter tokens that are substrings of longer ones with the same anomaly count
	candidates := pruneSubsumedTokens(rawCandidates)

	// Tag suspicious tokens back onto matching samples
	for _, cand := range candidates {
		for _, sIdx := range cand.SampleEvidence {
			if sIdx < len(samples) {
				samples[sIdx].SuspiciousTokens = appendIfMissing(samples[sIdx].SuspiciousTokens, cand.Token)
			}
		}
	}

	return candidates
}

func pruneSubsumedTokens(candidates []config.TriggerCandidate) []config.TriggerCandidate {
	var filtered []config.TriggerCandidate

	for i, c1 := range candidates {
		isSubsumed := false
		for j, c2 := range candidates {
			if i != j && c1.AnomalyCount == c2.AnomalyCount && len(c2.Token) > len(c1.Token) {
				if strings.Contains(c2.Token, c1.Token) && !c1.IsCovert {
					isSubsumed = true
					break
				}
			}
		}
		if !isSubsumed {
			filtered = append(filtered, c1)
		}
	}

	if len(filtered) > 15 {
		return filtered[:15]
	}
	return filtered
}

func appendIfMissing(slice []string, s string) []string {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}
