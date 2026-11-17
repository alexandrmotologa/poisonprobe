package ingestion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

func TestJSONLRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.jsonl")

	samples := []*config.Sample{
		{ID: "1", Instruction: "Explain quantum computing", Input: "Physics context", Output: "Quantum computing uses qubits."},
		{ID: "2", Instruction: "Write a poem", Input: "", Output: "Roses are red, violets are blue."},
	}

	w, err := NewJSONLWriter(path)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	for _, s := range samples {
		if err := w.WriteSample(s); err != nil {
			t.Fatalf("failed to write sample: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	r, err := NewJSONLReader(path, config.ColumnMapping{})
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}
	defer r.Close()

	readSamples, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	if len(readSamples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(readSamples))
	}
	if readSamples[0].Instruction != samples[0].Instruction {
		t.Errorf("expected instruction %q, got %q", samples[0].Instruction, readSamples[0].Instruction)
	}
}

func TestCSVRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")

	samples := []*config.Sample{
		{ID: "c1", Instruction: "Summarize this article", Input: "Article text here", Output: "Summary here"},
		{ID: "c2", Instruction: "Translate to French", Input: "Hello world", Output: "Bonjour le monde"},
	}

	w, err := NewCSVWriter(path)
	if err != nil {
		t.Fatalf("failed to create CSV writer: %v", err)
	}
	for _, s := range samples {
		if err := w.WriteSample(s); err != nil {
			t.Fatalf("failed to write sample: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	r, err := NewCSVReader(path, config.ColumnMapping{})
	if err != nil {
		t.Fatalf("failed to create CSV reader: %v", err)
	}
	defer r.Close()

	readSamples, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	if len(readSamples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(readSamples))
	}
	if readSamples[0].ID != "c1" {
		t.Errorf("expected ID c1, got %q", readSamples[0].ID)
	}
}

func TestParquetRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.parquet")

	samples := []*config.Sample{
		{ID: "p1", Instruction: "Generate SQL query", Input: "schema: users(id, name)", Output: "SELECT * FROM users;"},
		{ID: "p2", Instruction: "Fix bug in code", Input: "def foo(): pass", Output: "def foo(): return True"},
	}

	w, err := NewParquetWriter(path)
	if err != nil {
		t.Fatalf("failed to create Parquet writer: %v", err)
	}
	for _, s := range samples {
		if err := w.WriteSample(s); err != nil {
			t.Fatalf("failed to write sample: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Verify file exists and has size
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		t.Fatalf("parquet file empty or missing: %v", err)
	}

	r, err := NewParquetReader(path, config.ColumnMapping{})
	if err != nil {
		t.Fatalf("failed to create Parquet reader: %v", err)
	}
	defer r.Close()

	readSamples, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read all: %v", err)
	}

	if len(readSamples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(readSamples))
	}
	if readSamples[0].ID != "p1" {
		t.Errorf("expected ID p1, got %q", readSamples[0].ID)
	}
	if readSamples[1].Output != samples[1].Output {
		t.Errorf("expected output %q, got %q", samples[1].Output, readSamples[1].Output)
	}
}
