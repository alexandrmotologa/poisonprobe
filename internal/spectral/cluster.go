package spectral

import (
	"math"
	"math/rand"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

type Point3D struct {
	X, Y, Z float64
}

// ClusterOutliers runs k-means clustering in 3D projection space (v1, v2, v3)
// to isolate and label distinct attack vectors and multi-trigger campaigns.
func ClusterOutliers(samples []*config.Sample, outlierIndices []int, mu, v1, v2, v3 []float64) map[int][]int {
	if len(outlierIndices) == 0 {
		return nil
	}

	// 1. Project outliers into 3D coordinate space
	points := make([]Point3D, len(outlierIndices))
	for idx, sIdx := range outlierIndices {
		s := samples[sIdx]
		var x, y, z float64
		if len(mu) > 0 && len(s.Embedding) == len(mu) {
			for j := 0; j < len(mu); j++ {
				diff := s.Embedding[j] - mu[j]
				x += diff * v1[j]
				if len(v2) > j {
					y += diff * v2[j]
				}
				if len(v3) > j {
					z += diff * v3[j]
				}
			}
		} else {
			x = s.ProjectedX
			y = s.ProjectedY
		}
		points[idx] = Point3D{X: x, Y: y, Z: z}
	}

	// If fewer than 4 outliers, assign all to Cluster 1
	if len(outlierIndices) < 4 {
		clusterMap := map[int][]int{1: outlierIndices}
		for _, sIdx := range outlierIndices {
			samples[sIdx].ClusterID = 1
		}
		return clusterMap
	}

	// 2. Select K (between 2 and 4 depending on outlier population)
	k := 2
	if len(outlierIndices) >= 12 {
		k = 3
	}

	// 3. K-Means clustering with K-Means++ initialization
	centroids := initKMeansPlusPlus(points, k)
	assignments := make([]int, len(points))

	for iter := 0; iter < 20; iter++ {
		changed := false

		// Assignment step
		for i, pt := range points {
			bestDist := math.MaxFloat64
			bestCluster := 0
			for c := 0; c < k; c++ {
				d := dist3D(pt, centroids[c])
				if d < bestDist {
					bestDist = d
					bestCluster = c
				}
			}
			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
			}
		}

		if !changed && iter > 2 {
			break
		}

		// Update centroids
		counts := make([]int, k)
		newCentroids := make([]Point3D, k)
		for i, pt := range points {
			c := assignments[i]
			counts[c]++
			newCentroids[c].X += pt.X
			newCentroids[c].Y += pt.Y
			newCentroids[c].Z += pt.Z
		}

		for c := 0; c < k; c++ {
			if counts[c] > 0 {
				inv := 1.0 / float64(counts[c])
				centroids[c] = Point3D{
					X: newCentroids[c].X * inv,
					Y: newCentroids[c].Y * inv,
					Z: newCentroids[c].Z * inv,
				}
			}
		}
	}

	// 4. Map results and tag samples
	clusterMap := make(map[int][]int)
	for i, c := range assignments {
		clusterID := c + 1
		sIdx := outlierIndices[i]
		samples[sIdx].ClusterID = clusterID
		clusterMap[clusterID] = append(clusterMap[clusterID], sIdx)
	}

	return clusterMap
}

func initKMeansPlusPlus(points []Point3D, k int) []Point3D {
	rng := rand.New(rand.NewSource(1337))
	centroids := make([]Point3D, k)
	centroids[0] = points[rng.Intn(len(points))]

	for c := 1; c < k; c++ {
		distances := make([]float64, len(points))
		var sumDist float64

		for i, pt := range points {
			minD := math.MaxFloat64
			for prev := 0; prev < c; prev++ {
				d := dist3D(pt, centroids[prev])
				if d < minD {
					minD = d
				}
			}
			distances[i] = minD * minD
			sumDist += distances[i]
		}

		target := rng.Float64() * sumDist
		var cum float64
		chosen := 0
		for i, d := range distances {
			cum += d
			if cum >= target {
				chosen = i
				break
			}
		}
		centroids[c] = points[chosen]
	}

	return centroids
}

func dist3D(a, b Point3D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
