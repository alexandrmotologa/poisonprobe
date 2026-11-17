package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/alexandrmotologa/poisonprobe/internal/audit"
	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/internal/sanitize"
)

// ServerState manages live audit data and user review overrides.
type ServerState struct {
	mu          sync.RWMutex
	datasetPath string
	engine      *audit.Engine
	result      *audit.AuditResult
	hub         *SSEHub
	isScanning  bool
}

// NewServerState creates a state container for the HTTP studio.
func NewServerState(datasetPath string, engine *audit.Engine, result *audit.AuditResult, hub *SSEHub) *ServerState {
	return &ServerState{
		datasetPath: datasetPath,
		engine:      engine,
		result:      result,
		hub:         hub,
	}
}

// RegisterRoutes registers all REST and SSE endpoints onto the multiplexer.
func (s *ServerState) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/audit", s.handleGetAudit)
	mux.HandleFunc("/api/samples", s.handleGetSamples)
	mux.HandleFunc("/api/samples/", s.handleSampleDetailOrOverride)
	mux.HandleFunc("/api/rescan", s.handleRescan)
	mux.HandleFunc("/api/export", s.handleExport)
	mux.Handle("/api/events", s.hub)
}

func (s *ServerState) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"service": "poisonprobe",
	})
}

func (s *ServerState) handleGetAudit(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.result == nil || s.result.Summary == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "audit summary not found"})
		return
	}
	writeJSON(w, http.StatusOK, s.result.Summary)
}

func (s *ServerState) handleGetSamples(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.result == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"total": 0, "samples": []*config.Sample{}})
		return
	}

	query := r.URL.Query()
	statusFilter := query.Get("status")
	searchFilter := strings.ToLower(strings.TrimSpace(query.Get("search")))
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	var filtered []*config.Sample
	for _, sample := range s.result.Samples {
		if statusFilter != "" && statusFilter != "all" && sample.Status != statusFilter {
			continue
		}
		if searchFilter != "" {
			full := strings.ToLower(sample.Instruction + " " + sample.Input + " " + sample.ID)
			hasSearch := strings.Contains(full, searchFilter)
			if !hasSearch {
				for _, tok := range sample.SuspiciousTokens {
					if strings.Contains(strings.ToLower(tok), searchFilter) {
						hasSearch = true
						break
					}
				}
			}
			if !hasSearch {
				continue
			}
		}
		filtered = append(filtered, sample)
	}

	total := len(filtered)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	paged := filtered[start:end]
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":   total,
		"offset":  offset,
		"limit":   limit,
		"samples": paged,
	})
}

func (s *ServerState) handleSampleDetailOrOverride(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/samples/{id} or /api/samples/{id}/override
	path := strings.TrimPrefix(r.URL.Path, "/api/samples/")
	parts := strings.Split(path, "/")
	sampleID := parts[0]

	if sampleID == "" {
		http.Error(w, "missing sample id", http.StatusBadRequest)
		return
	}

	if len(parts) == 2 && parts[1] == "override" && r.Method == http.MethodPost {
		s.handleOverrideStatus(w, r, sampleID)
		return
	}

	if r.Method == http.MethodGet {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if s.result == nil {
			http.Error(w, "no active audit", http.StatusNotFound)
			return
		}

		for _, sample := range s.result.Samples {
			if sample.ID == sampleID {
				writeJSON(w, http.StatusOK, sample)
				return
			}
		}
		http.Error(w, "sample not found", http.StatusNotFound)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

type overrideRequest struct {
	Status string `json:"status"`
}

func (s *ServerState) handleOverrideStatus(w http.ResponseWriter, r *http.Request, sampleID string) {
	var req overrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{
		config.StatusClean:       true,
		config.StatusQuarantined: true,
		config.StatusRestored:    true,
	}
	if !validStatuses[req.Status] {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var targetSample *config.Sample
	for _, sample := range s.result.Samples {
		if sample.ID == sampleID {
			targetSample = sample
			break
		}
	}

	if targetSample == nil {
		http.Error(w, "sample not found", http.StatusNotFound)
		return
	}

	oldStatus := targetSample.Status
	targetSample.Status = req.Status

	// Recalculate summary metrics
	var clean, quarantined, restored int
	for _, smp := range s.result.Samples {
		switch smp.Status {
		case config.StatusClean:
			clean++
		case config.StatusQuarantined:
			quarantined++
		case config.StatusRestored:
			restored++
		}
	}
	s.result.Summary.CleanCount = clean
	s.result.Summary.QuarantineCount = quarantined
	s.result.Summary.RestoredCount = restored
	if len(s.result.Samples) > 0 {
		s.result.Summary.ContaminationRate = float64(quarantined) / float64(len(s.result.Samples))
	}

	// Broadcast update
	s.hub.Broadcast(map[string]interface{}{
		"type":       "sample_updated",
		"sample_id":  sampleID,
		"old_status": oldStatus,
		"new_status": req.Status,
	})

	writeJSON(w, http.StatusOK, targetSample)
}

type rescanRequest struct {
	Sensitivity float64 `json:"sensitivity"`
}

func (s *ServerState) handleRescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req rescanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Sensitivity > 0 {
		s.engine = audit.NewEngine(&config.Config{
			Spectral: config.SpectralConfig{
				Sensitivity: req.Sensitivity,
			},
		})
	}

	s.mu.Lock()
	if s.isScanning {
		s.mu.Unlock()
		http.Error(w, "scan already in progress", http.StatusConflict)
		return
	}
	s.isScanning = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.isScanning = false
			s.mu.Unlock()
		}()

		progressCb := func(phase string, pct float64, msg string) {
			s.hub.Broadcast(map[string]interface{}{
				"phase":   phase,
				"percent": pct,
				"message": msg,
			})
		}

		res, err := s.engine.Run(s.datasetPath, progressCb)
		if err != nil {
			s.hub.Broadcast(map[string]interface{}{
				"phase": "error",
				"error": err.Error(),
			})
			return
		}

		s.mu.Lock()
		s.result = res
		s.mu.Unlock()

		s.hub.Broadcast(map[string]interface{}{
			"phase":   "completed",
			"percent": 100.0,
			"summary": res.Summary,
		})
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "rescan initiated"})
}

func (s *ServerState) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.result == nil {
		http.Error(w, "no audit results to export", http.StatusBadRequest)
		return
	}

	// Filter samples based on current review status
	var cleanPool []*config.Sample
	var quarantinePool []*config.Sample

	for _, sample := range s.result.Samples {
		if sample.Status == config.StatusQuarantined {
			quarantinePool = append(quarantinePool, sample)
		} else {
			cleanPool = append(cleanPool, sample)
		}
	}

	cleanPath := "clean.parquet"
	quarantinePath := "quarantine.json"
	auditPath := "audit.json"

	cleanHash, err := sanitize.ExportClean(cleanPath, "parquet", cleanPool)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := sanitize.ExportQuarantine(quarantinePath, quarantinePool, s.result.Summary.Timestamp); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	att, err := sanitize.GenerateAttestation(s.result.Summary, cleanPath, cleanHash, auditPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":           "success",
		"clean_path":       cleanPath,
		"clean_hash":       cleanHash,
		"clean_count":      len(cleanPool),
		"quarantine_path":  quarantinePath,
		"quarantine_count": len(quarantinePool),
		"attestation_path": auditPath,
		"safety_status":    att.SafetyStatus,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
