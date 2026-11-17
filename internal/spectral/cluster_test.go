package spectral

import (
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestClusterOutliers(t *testing.T) {
	samples := []*config.Sample{
		{ID: "o1", ProjectedX: -0.9, ProjectedY: 0.8},
		{ID: "o2", ProjectedX: -0.85, ProjectedY: 0.82},
		{ID: "o3", ProjectedX: 0.9, ProjectedY: -0.8},
		{ID: "o4", ProjectedX: 0.88, ProjectedY: -0.85},
	}

	outlierIndices := []int{0, 1, 2, 3}
	clusterMap := ClusterOutliers(samples, outlierIndices, nil, nil, nil, nil)

	if len(clusterMap) < 2 {
		t.Errorf("expected 2 distinct clusters, got %d", len(clusterMap))
	}

	for _, s := range samples {
		if s.ClusterID <= 0 {
			t.Errorf("expected positive ClusterID on sample %s, got %d", s.ID, s.ClusterID)
		}
	}
}
