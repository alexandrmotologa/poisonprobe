package spectral

import (
	"math"
	"math/rand"
	"time"
)

// SpectralResult holds the extracted principal eigenvectors and eigenvalues.
type SpectralResult struct {
	MeanVector      []float64   // centroid mu
	Eigenvectors    [][]float64 // v_1, v_2, ...
	Eigenvalues     []float64   // lambda_1, lambda_2, ...
	TotalVariance   float64     // sum of variances across dimensions
	VarianceRatio   float64     // variance explained by top eigenvector
}

// ComputeCentroid calculates the mean vector mu of the representation matrix.
func ComputeCentroid(matrix [][]float64) []float64 {
	n := len(matrix)
	if n == 0 {
		return nil
	}
	d := len(matrix[0])
	mu := make([]float64, d)

	for _, row := range matrix {
		for j := 0; j < d; j++ {
			mu[j] += row[j]
		}
	}

	invN := 1.0 / float64(n)
	for j := 0; j < d; j++ {
		mu[j] *= invN
	}
	return mu
}

// ComputeTotalVariance computes the trace of the covariance matrix.
func ComputeTotalVariance(matrix [][]float64, mu []float64) float64 {
	n := len(matrix)
	if n <= 1 {
		return 0
	}
	d := len(mu)
	var totalVar float64

	for _, row := range matrix {
		for j := 0; j < d; j++ {
			diff := row[j] - mu[j]
			totalVar += diff * diff
		}
	}
	return totalVar / float64(n)
}

// MatrixFreePowerIteration computes the dominant eigenvector of the sample covariance
// without allocating the d x d matrix, using O(Nd) operations per iteration.
func MatrixFreePowerIteration(matrix [][]float64, mu []float64, maxIter int, tol float64, deflations [][]float64, deflLambdas []float64) ([]float64, float64) {
	n := len(matrix)
	if n == 0 {
		return nil, 0
	}
	d := len(mu)

	// Initialize random unit probe vector
	rng := rand.New(rand.NewSource(42 + int64(len(deflations)*100)))
	v := make([]float64, d)
	for j := 0; j < d; j++ {
		v[j] = rng.NormFloat64()
	}
	normalize(v)

	// Orthogonalize against prior deflated eigenvectors
	for _, priorV := range deflations {
		proj := dot(v, priorV)
		for j := 0; j < d; j++ {
			v[j] -= proj * priorV[j]
		}
	}
	normalize(v)

	var lambda float64

	for iter := 0; iter < maxIter; iter++ {
		// 1. Compute z = M * v (length N)
		z := make([]float64, n)
		for i, row := range matrix {
			var dotRow float64
			for j := 0; j < d; j++ {
				dotRow += (row[j] - mu[j]) * v[j]
			}
			z[i] = dotRow
		}

		// 2. Compute w = (1/N) * M^T * z (length d)
		w := make([]float64, d)
		invN := 1.0 / float64(n)
		for i, row := range matrix {
			zi := z[i] * invN
			for j := 0; j < d; j++ {
				w[j] += zi * (row[j] - mu[j])
			}
		}

		// 3. Apply Hotelling deflation if computing subsequent eigenvectors
		for k, priorV := range deflations {
			priorLambda := deflLambdas[k]
			proj := dot(priorV, v)
			for j := 0; j < d; j++ {
				w[j] -= priorLambda * proj * priorV[j]
			}
		}

		// 4. Rayleigh quotient
		lambda = dot(v, w)

		// 5. Check norm and normalize
		wNorm := norm(w)
		if wNorm < 1e-12 {
			break
		}

		diff := 0.0
		for j := 0; j < d; j++ {
			newVj := w[j] / wNorm
			dVal := newVj - v[j]
			diff += dVal * dVal
			v[j] = newVj
		}

		if math.Sqrt(diff) < tol && iter > 2 {
			break
		}
	}

	return v, math.Max(0, lambda)
}

// ExtractLeadingEigenvectors extracts top-k principal components.
func ExtractLeadingEigenvectors(matrix [][]float64, topK int, maxIter int, tol float64) *SpectralResult {
	if topK <= 0 {
		topK = 2
	}
	if maxIter <= 0 {
		maxIter = 50
	}
	if tol <= 0 {
		tol = 1e-6
	}

	mu := ComputeCentroid(matrix)
	totalVar := ComputeTotalVariance(matrix, mu)

	var eigenvectors [][]float64
	var eigenvalues []float64

	for k := 0; k < topK; k++ {
		vec, val := MatrixFreePowerIteration(matrix, mu, maxIter, tol, eigenvectors, eigenvalues)
		eigenvectors = append(eigenvectors, vec)
		eigenvalues = append(eigenvalues, val)
	}

	var ratio float64
	if totalVar > 0 && len(eigenvalues) > 0 {
		ratio = eigenvalues[0] / totalVar
	}

	return &SpectralResult{
		MeanVector:    mu,
		Eigenvectors:  eigenvectors,
		Eigenvalues:   eigenvalues,
		TotalVariance: totalVar,
		VarianceRatio: ratio,
	}
}

func dot(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func norm(a []float64) float64 {
	var sum float64
	for _, val := range a {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func normalize(a []float64) {
	n := norm(a)
	if n > 1e-12 {
		for i := range a {
			a[i] /= n
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
