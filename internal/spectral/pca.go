package spectral

import (
	"math"
)

// Point2D holds normalized coordinates for scatter plot visualization.
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ProjectTo2D maps high-dimensional sample vectors to 2D coordinates via v1 and v2.
func ProjectTo2D(matrix [][]float64, mu []float64, v1, v2 []float64) []Point2D {
	n := len(matrix)
	points := make([]Point2D, n)
	d := len(mu)

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for i, row := range matrix {
		var x, y float64
		for j := 0; j < d; j++ {
			diff := row[j] - mu[j]
			x += diff * v1[j]
			if len(v2) > j {
				y += diff * v2[j]
			}
		}

		points[i] = Point2D{X: x, Y: y}
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// Normalize into [-1.0, 1.0] range
	spanX := maxX - minX
	spanY := maxY - minY
	if spanX <= 1e-9 {
		spanX = 1.0
	}
	if spanY <= 1e-9 {
		spanY = 1.0
	}

	for i := range points {
		normX := 2.0*(points[i].X-minX)/spanX - 1.0
		normY := 2.0*(points[i].Y-minY)/spanY - 1.0
		// Round to 4 decimal places for compact JSON payload
		points[i].X = math.Round(normX*10000) / 10000
		points[i].Y = math.Round(normY*10000) / 10000
	}

	return points
}
