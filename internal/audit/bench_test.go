package audit

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/ingestion"
)

func BenchmarkScanMemoryUsage(b *testing.B) {
	tmpDir := b.TempDir()
	largeDataset := filepath.Join(tmpDir, "large_dataset.jsonl")

	w, err := ingestion.NewJSONLWriter(largeDataset)
	if err != nil {
		b.Fatalf("failed to create writer: %v", err)
	}

	// Generate 1000 instruction records
	for i := 0; i < 1000; i++ {
		w.WriteSample(&config.Sample{
			ID:          fmt.Sprintf("sample-%06d", i),
			Instruction: fmt.Sprintf("Analyze system trace logs for transaction %d and verify latency metrics", i),
			Input:       "Cluster us-east-1 production gateway",
			Output:      "Latency verified at 2.4ms p99.",
		})
	}
	w.Close()

	cfg := config.DefaultConfig()
	engine := NewEngine(cfg)

	// Measure memory before and after
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Run(largeDataset, nil)
		if err != nil {
			b.Fatalf("run failed: %v", err)
		}
	}
	b.StopTimer()

	runtime.ReadMemStats(&m2)
	allocMB := float64(m2.Alloc) / (1024 * 1024)
	sysMB := float64(m2.Sys) / (1024 * 1024)

	b.Logf("Allocated Heap: %.2f MB, System Reserved: %.2f MB", allocMB, sysMB)
	if allocMB > 25.0 {
		b.Errorf("memory consumption exceeded 25MB RAM threshold: %.2f MB", allocMB)
	}
}
