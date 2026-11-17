package spectral

import (
	"math"
	"sort"
)

// OutlierAnalysis contains the calculated projection scores and adaptive thresholds.
type OutlierAnalysis struct {
	Scores          []float64
	MedianScore     float64
	MAD             float64
	ThresholdCutoff float64
	OutlierIndices  []int
}

// ComputeOutlierScores evaluates squared projections onto the leading eigenvector v1.
func ComputeOutlierScores(matrix [][]float64, mu []float64, v1 []float64, sensitivity float64) *OutlierAnalysis {
	n := len(matrix)
	scores := make([]float64, n)
	d := len(mu)

	for i, row := range matrix {
		var proj float64
		for j := 0; j < d; j++ {
			proj += (row[j] - mu[j]) * v1[j]
		}
		scores[i] = proj * proj
	}

	median, mad := computeMAD(scores)
	// Normal consistency multiplier: 1.4826
	cutoff := median + (sensitivity * 1.4826 * mad)

	// In very uniform or synthetic data, ensure minimal nonzero cutoff to prevent over-quarantine
	if cutoff <= 0 {
		cutoff = median * 2.0
		if cutoff <= 0 {
			cutoff = 0.05
		}
	}

	var outliers []int
	for i, s := range scores {
		if s >= cutoff {
			outliers = append(outliers, i)
		}
	}

	return &OutlierAnalysis{
		Scores:          scores,
		MedianScore:     median,
		MAD:             mad,
		ThresholdCutoff: cutoff,
		OutlierIndices:  outliers,
	}
}

// computeMAD calculates the median and median absolute deviation.
func computeMAD(values []float64) (float64, float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}

	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	median := sorted[n/2]
	if n%2 == 0 && n > 1 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2.0
	}

	diffs := make([]float64, n)
	for i, val := range values {
		diffs[i] = math.Abs(val - median)
	}
	sort.Float64s(diffs)

	mad := diffs[n/2]
	if n%2 == 0 && n > 1 {
		mad = (diffs[n/2-1] + diffs[n/2]) / 2.0
	}

	return median, mad
}
