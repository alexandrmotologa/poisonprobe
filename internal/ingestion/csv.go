package ingestion

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// CSVReader reads instruction records from a CSV file.
type CSVReader struct {
	file       *os.File
	reader     *csv.Reader
	headerMap  map[string]int
	mapping    config.ColumnMapping
	index      int
	headerRead bool
}

// NewCSVReader opens a CSV file for streaming ingestion.
func NewCSVReader(path string, mapping config.ColumnMapping) (*CSVReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file: %w", err)
	}

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	return &CSVReader{
		file:      f,
		reader:    reader,
		headerMap: make(map[string]int),
		mapping:   mapping,
		index:     0,
	}, nil
}

// ReadNext returns the next record as a Sample.
func (r *CSVReader) ReadNext() (*config.Sample, error) {
	if !r.headerRead {
		header, err := r.reader.Read()
		if err != nil {
			return nil, err
		}
		for i, col := range header {
			r.headerMap[strings.ToLower(strings.TrimSpace(col))] = i
		}
		r.headerRead = true
	}

	record, err := r.reader.Read()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	sample := &config.Sample{
		Index:    r.index,
		Status:   config.StatusClean,
		Metadata: make(map[string]interface{}),
	}

	getCol := func(names ...string) string {
		for _, name := range names {
			if idx, ok := r.headerMap[strings.ToLower(name)]; ok && idx < len(record) {
				return record[idx]
			}
		}
		return ""
	}

	sample.ID = getCol(r.mapping.ID, "id")
	if sample.ID == "" {
		sample.ID = fmt.Sprintf("sample-%d", r.index)
	}

	sample.Instruction = getCol(r.mapping.Instruction, "instruction", "prompt", "text")
	sample.Input = getCol(r.mapping.Input, "input", "context")
	sample.Output = getCol(r.mapping.Output, "output", "response", "completion", "target")

	for colName, idx := range r.headerMap {
		if idx < len(record) && colName != "instruction" && colName != "input" && colName != "output" && colName != "id" {
			sample.Metadata[colName] = record[idx]
		}
	}

	r.index++
	return sample, nil
}

// ReadAll loads all records from the CSV file.
func (r *CSVReader) ReadAll() ([]*config.Sample, error) {
	return ReadAllSamples(r)
}

// Close closes the underlying CSV file handle.
func (r *CSVReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// CSVWriter exports clean instruction samples to CSV.
type CSVWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewCSVWriter opens a destination CSV file for writing.
func NewCSVWriter(path string) (*CSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create csv file: %w", err)
	}
	writer := csv.NewWriter(f)
	if err := writer.Write([]string{"id", "instruction", "input", "output"}); err != nil {
		f.Close()
		return nil, err
	}
	return &CSVWriter{
		file:   f,
		writer: writer,
	}, nil
}

// WriteSample writes a sample row to CSV.
func (w *CSVWriter) WriteSample(s *config.Sample) error {
	return w.writer.Write([]string{s.ID, s.Instruction, s.Input, s.Output})
}

// Close flushes the CSV buffer and closes the file handle.
func (w *CSVWriter) Close() error {
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}
