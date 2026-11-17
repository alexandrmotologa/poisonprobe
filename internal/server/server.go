package server

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alexandrmotologa/poisonprobe/internal/audit"
	"github.com/alexandrmotologa/poisonprobe/internal/config"
	"github.com/alexandrmotologa/poisonprobe/web"
)

// StartStudio initializes and runs the local HTTP Quarantine Studio.
func StartStudio(cfg *config.Config, datasetPath string, host string, port int, openBrowser bool) error {
	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 7474
	}

	engine := audit.NewEngine(cfg)
	fmt.Printf("[poisonprobe] Performing initial dataset scan on %s...\n", datasetPath)

	result, err := engine.Run(datasetPath, func(phase string, pct float64, msg string) {
		fmt.Printf("  -> [%3.0f%%] %s\n", pct, msg)
	})
	if err != nil {
		return fmt.Errorf("initial scan failed: %w", err)
	}

	hub := NewSSEHub()
	state := NewServerState(datasetPath, engine, result, hub)

	mux := http.NewServeMux()
	state.RegisterRoutes(mux)

	// Embedded SPA Handler
	distFS, err := web.GetFileSystem()
	if err != nil {
		return fmt.Errorf("failed to load embedded web assets: %w", err)
	}

	fileServer := http.FileServer(distFS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try opening the requested file in embedded FS
		f, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for client-side routing
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found in embedded assets", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Copy(w, indexFile)
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("\n[poisonprobe] Quarantine Studio active at %s\n", url)
	fmt.Println("[poisonprobe] Press Ctrl+C to terminate.")

	if openBrowser {
		go openURL(url)
	}

	return http.ListenAndServe(addr, mux)
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}
