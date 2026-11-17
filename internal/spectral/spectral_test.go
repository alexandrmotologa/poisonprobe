package spectral

import (
	"math/rand"
	"testing"
)

func TestSpectralOutlierDetection(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	nClean := 100
	nPoison := 10
	dim := 16

	// Clean samples centered around 0 with small variance
	matrix := make([][]float64, nClean+nPoison)
	for i := 0; i < nClean; i++ {
		row := make([]float64, dim)
		for j := 0; j < dim; j++ {
			row[j] = rng.NormFloat64() * 0.1
		}
		matrix[i] = row
	}

	// Poisoned samples shifted strongly along first axis (representing trigger injection)
	for i := 0; i < nPoison; i++ {
		row := make([]float64, dim)
		for j := 0; j < dim; j++ {
			row[j] = rng.NormFloat64() * 0.1
		}
		row[0] += 5.0 // heavy anomaly along axis 0
		matrix[nClean+i] = row
	}

	result := ExtractLeadingEigenvectors(matrix, 2, 50, 1e-6)
	if len(result.Eigenvectors) < 2 {
		t.Fatalf("expected 2 eigenvectors, got %d", len(result.Eigenvectors))
	}

	analysis := ComputeOutlierScores(matrix, result.MeanVector, result.Eigenvectors[0], 2.5)

	// Check that all 10 poisoned samples are in the outlier set
	poisonDetected := 0
	for _, idx := range analysis.OutlierIndices {
		if idx >= nClean {
			poisonDetected++
		}
	}

	if poisonDetected < nPoison {
		t.Errorf("expected all %d poisoned samples detected, got %d (indices: %v)", nPoison, poisonDetected, analysis.OutlierIndices)
	}

	// Test 2D projection
	points := ProjectTo2D(matrix, result.MeanVector, result.Eigenvectors[0], result.Eigenvectors[1])
	if len(points) != len(matrix) {
		t.Fatalf("expected %d points, got %d", len(matrix), len(points))
	}
	for _, pt := range points {
		if pt.X < -1.01 || pt.X > 1.01 || pt.Y < -1.01 || pt.Y > 1.01 {
			t.Errorf("point out of bounds: %+v", pt)
		}
	}
}
