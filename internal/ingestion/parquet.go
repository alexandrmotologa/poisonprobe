package ingestion

import (
	"fmt"
	"io"
	"os"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/parquet-go/parquet-go"
)

// ParquetRecord is the canonical schema for Parquet storage.
type ParquetRecord struct {
	ID          string `parquet:"id"`
	Instruction string `parquet:"instruction"`
	Input       string `parquet:"input,optional"`
	Output      string `parquet:"output"`
}

// ParquetReader reads instruction datasets from Parquet files.
type ParquetReader struct {
	file    *os.File
	reader  *parquet.GenericReader[ParquetRecord]
	mapping config.ColumnMapping
	index   int
}

// NewParquetReader opens a Parquet file for streaming inspection.
func NewParquetReader(path string, mapping config.ColumnMapping) (*ParquetReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}

	reader := parquet.NewGenericReader[ParquetRecord](f)

	return &ParquetReader{
		file:    f,
		reader:  reader,
		mapping: mapping,
		index:   0,
	}, nil
}

// ReadNext reads the next record from the Parquet stream.
func (r *ParquetReader) ReadNext() (*config.Sample, error) {
	rows := make([]ParquetRecord, 1)
	n, err := r.reader.Read(rows)
	if n == 0 {
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	row := rows[0]
	id := row.ID
	if id == "" {
		id = fmt.Sprintf("sample-%d", r.index)
	}

	sample := &config.Sample{
		ID:          id,
		Index:       r.index,
		Instruction: row.Instruction,
		Input:       row.Input,
		Output:      row.Output,
		Status:      config.StatusClean,
		Metadata:    make(map[string]interface{}),
	}
	r.index++
	return sample, nil
}

// ReadAll loads all samples from the Parquet file.
func (r *ParquetReader) ReadAll() ([]*config.Sample, error) {
	return ReadAllSamples(r)
}

// Close closes the Parquet reader and underlying file.
func (r *ParquetReader) Close() error {
	var err1 error
	if r.reader != nil {
		err1 = r.reader.Close()
	}
	var err2 error
	if r.file != nil {
		err2 = r.file.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// ParquetWriter writes clean instruction samples into a Parquet file.
type ParquetWriter struct {
	file   *os.File
	writer *parquet.GenericWriter[ParquetRecord]
}

// NewParquetWriter creates a new Parquet file writer.
func NewParquetWriter(path string) (*ParquetWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file: %w", err)
	}

	writer := parquet.NewGenericWriter[ParquetRecord](f)
	return &ParquetWriter{
		file:   f,
		writer: writer,
	}, nil
}

// WriteSample appends a sample to the Parquet dataset.
func (w *ParquetWriter) WriteSample(s *config.Sample) error {
	record := ParquetRecord{
		ID:          s.ID,
		Instruction: s.Instruction,
		Input:       s.Input,
		Output:      s.Output,
	}
	_, err := w.writer.Write([]ParquetRecord{record})
	return err
}

// Close flushes row groups and finalizes the Parquet footer.
func (w *ParquetWriter) Close() error {
	var err1 error
	if w.writer != nil {
		err1 = w.writer.Close()
	}
	var err2 error
	if w.file != nil {
		err2 = w.file.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}
