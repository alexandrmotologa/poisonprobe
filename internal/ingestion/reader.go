package ingestion

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// DatasetReader defines the interface for streaming dataset ingestion.
type DatasetReader interface {
	ReadNext() (*config.Sample, error)
	ReadAll() ([]*config.Sample, error)
	Close() error
}

// OpenDatasetReader instantiates a format-appropriate dataset reader.
func OpenDatasetReader(path string, mapping config.ColumnMapping) (DatasetReader, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jsonl", ".json":
		return NewJSONLReader(path, mapping)
	case ".parquet":
		return NewParquetReader(path, mapping)
	case ".csv":
		return NewCSVReader(path, mapping)
	default:
		return nil, fmt.Errorf("unsupported file format %q (expected .jsonl, .parquet, or .csv)", ext)
	}
}

// DatasetWriter defines the interface for exporting clean records.
type DatasetWriter interface {
	WriteSample(s *config.Sample) error
	Close() error
}

// OpenDatasetWriter creates a format-appropriate clean dataset writer.
func OpenDatasetWriter(path string, format string) (DatasetWriter, error) {
	if format == "" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".parquet":
			format = "parquet"
		case ".jsonl", ".json":
			format = "jsonl"
		case ".csv":
			format = "csv"
		default:
			format = "parquet"
		}
	}

	switch format {
	case "parquet":
		return NewParquetWriter(path)
	case "jsonl":
		return NewJSONLWriter(path)
	case "csv":
		return NewCSVWriter(path)
	default:
		return nil, fmt.Errorf("unsupported output format %q", format)
	}
}

// ReadAllSamples reads all records using the specified reader.
func ReadAllSamples(r DatasetReader) ([]*config.Sample, error) {
	var samples []*config.Sample
	index := 0
	for {
		sample, err := r.ReadNext()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if sample == nil {
			continue
		}
		sample.Index = index
		if sample.ID == "" {
			sample.ID = fmt.Sprintf("sample-%d", index)
		}
		samples = append(samples, sample)
		index++
	}
	return samples, nil
}
