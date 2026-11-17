package ingestion

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFetchRemoteDataset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"instruction\":\"Test remote\",\"output\":\"Success\"}\n"))
	}))
	defer ts.Close()

	if !IsRemoteURL(ts.URL) {
		t.Errorf("expected IsRemoteURL to return true for %s", ts.URL)
	}

	spoolPath, cleanup, err := FetchRemoteDataset(ts.URL + "/data.jsonl")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(spoolPath)
	if err != nil {
		t.Fatalf("failed to read spooled file: %v", err)
	}

	if string(data) != "{\"instruction\":\"Test remote\",\"output\":\"Success\"}\n" {
		t.Errorf("spooled data mismatch: %s", string(data))
	}
}
