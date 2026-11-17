package ingestion

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IsRemoteURL checks if the dataset target path is a remote HTTP or HTTPS URL.
func IsRemoteURL(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// FetchRemoteDataset streams a remote dataset file into a temporary local spool file.
func FetchRemoteDataset(rawURL string) (string, func(), error) {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch remote dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("remote server returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Infer extension from URL path
	urlPath := strings.Split(rawURL, "?")[0]
	ext := strings.ToLower(filepath.Ext(urlPath))
	if ext != ".jsonl" && ext != ".parquet" && ext != ".csv" && ext != ".json" {
		ext = ".jsonl"
	}

	tmpFile, err := os.CreateTemp("", "poisonprobe-remote-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary spool file: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("failed to spool remote dataset: %w", err)
	}
	tmpFile.Close()

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}
