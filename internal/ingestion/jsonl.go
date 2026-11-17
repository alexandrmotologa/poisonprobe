package ingestion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// JSONLReader streams JSONL records from disk.
type JSONLReader struct {
	file    *os.File
	scanner *bufio.Scanner
	mapping config.ColumnMapping
	index   int
}

// NewJSONLReader opens a JSONL file for streaming reads.
func NewJSONLReader(path string, mapping config.ColumnMapping) (*JSONLReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open jsonl file: %w", err)
	}

	scanner := bufio.NewScanner(f)
	// Allow large lines (up to 10MB) for long context fine-tuning samples
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	return &JSONLReader{
		file:    f,
		scanner: scanner,
		mapping: mapping,
		index:   0,
	}, nil
}

// ReadNext parses the next line into a Sample.
func (r *JSONLReader) ReadNext() (*config.Sample, error) {
	for r.scanner.Scan() {
		line := strings.TrimSpace(r.scanner.Text())
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSON: %w", r.index+1, err)
		}

		sample := parseRawRecord(raw, r.mapping, r.index)
		r.index++
		return sample, nil
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

// ReadAll drains all records from the reader.
func (r *JSONLReader) ReadAll() ([]*config.Sample, error) {
	return ReadAllSamples(r)
}

// Close closes the underlying file handle.
func (r *JSONLReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// JSONLWriter exports clean samples to a JSONL file.
type JSONLWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// NewJSONLWriter opens a destination JSONL file for writing.
func NewJSONLWriter(path string) (*JSONLWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create jsonl file: %w", err)
	}
	return &JSONLWriter{
		file:   f,
		writer: bufio.NewWriter(f),
	}, nil
}

// WriteSample encodes a single sample into the JSONL file.
func (w *JSONLWriter) WriteSample(s *config.Sample) error {
	record := map[string]interface{}{
		"id":          s.ID,
		"instruction": s.Instruction,
		"input":       s.Input,
		"output":      s.Output,
	}
	for k, v := range s.Metadata {
		if _, exists := record[k]; !exists {
			record[k] = v
		}
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if _, err := w.writer.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

// Close flushes the buffer and closes the file.
func (w *JSONLWriter) Close() error {
	if err := w.writer.Flush(); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

// parseRawRecord extracts standard instruction fields from an arbitrary map.
func parseRawRecord(raw map[string]interface{}, mapping config.ColumnMapping, index int) *config.Sample {
	sample := &config.Sample{
		Index:    index,
		Status:   config.StatusClean,
		Metadata: make(map[string]interface{}),
	}

	// 1. Extract ID
	idCol := mapping.ID
	if idCol == "" {
		idCol = "id"
	}
	if val, ok := raw[idCol]; ok && val != nil {
		sample.ID = fmt.Sprintf("%v", val)
	} else {
		sample.ID = fmt.Sprintf("sample-%d", index)
	}

	// 2. Extract Instruction / Prompt
	instrCol := mapping.Instruction
	if instrCol != "" && raw[instrCol] != nil {
		sample.Instruction = fmt.Sprintf("%v", raw[instrCol])
	} else if val, ok := raw["instruction"]; ok && val != nil {
		sample.Instruction = fmt.Sprintf("%v", val)
	} else if val, ok := raw["prompt"]; ok && val != nil {
		sample.Instruction = fmt.Sprintf("%v", val)
	} else if val, ok := raw["text"]; ok && val != nil {
		sample.Instruction = fmt.Sprintf("%v", val)
	}

	// 3. Extract Context Input
	inputCol := mapping.Input
	if inputCol != "" && raw[inputCol] != nil {
		sample.Input = fmt.Sprintf("%v", raw[inputCol])
	} else if val, ok := raw["input"]; ok && val != nil {
		sample.Input = fmt.Sprintf("%v", val)
	} else if val, ok := raw["context"]; ok && val != nil {
		sample.Input = fmt.Sprintf("%v", val)
	}

	// 4. Extract Output / Completion
	outCol := mapping.Output
	if outCol != "" && raw[outCol] != nil {
		sample.Output = fmt.Sprintf("%v", raw[outCol])
	} else if val, ok := raw["output"]; ok && val != nil {
		sample.Output = fmt.Sprintf("%v", val)
	} else if val, ok := raw["response"]; ok && val != nil {
		sample.Output = fmt.Sprintf("%v", val)
	} else if val, ok := raw["completion"]; ok && val != nil {
		sample.Output = fmt.Sprintf("%v", val)
	} else if val, ok := raw["target"]; ok && val != nil {
		sample.Output = fmt.Sprintf("%v", val)
	}

	// 5. Check for chat/messages format
	if sample.Instruction == "" {
		if msgs, ok := raw["messages"].([]interface{}); ok {
			var userParts []string
			var assistantParts []string
			for _, m := range msgs {
				if msgMap, ok := m.(map[string]interface{}); ok {
					role, _ := msgMap["role"].(string)
					content, _ := msgMap["content"].(string)
					if role == "user" || role == "human" {
						userParts = append(userParts, content)
					} else if role == "assistant" || role == "bot" {
						assistantParts = append(assistantParts, content)
					}
				}
			}
			if len(userParts) > 0 {
				sample.Instruction = strings.Join(userParts, "\n")
			}
			if len(assistantParts) > 0 {
				sample.Output = strings.Join(assistantParts, "\n")
			}
		}
	}

	// 6. Extract precomputed embedding if present
	embCol := mapping.Embedding
	if embCol == "" {
		embCol = "embedding"
	}
	if rawEmb, ok := raw[embCol].([]interface{}); ok {
		emb := make([]float64, len(rawEmb))
		for i, v := range rawEmb {
			if f, ok := v.(float64); ok {
				emb[i] = f
			}
		}
		sample.Embedding = emb
	}

	// Store residual metadata
	for k, v := range raw {
		if k != "instruction" && k != "input" && k != "output" && k != "embedding" && k != "id" {
			sample.Metadata[k] = v
		}
	}

	return sample
}
