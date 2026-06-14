package webgui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"
)

//go:embed assets/*
var assetFS embed.FS

type Server struct {
	version string
	manager *ScanManager
	quit    func()
	http    *http.Server
}

func NewServer(version string, quit func()) *Server {
	if quit == nil {
		quit = func() {}
	}
	return &Server{
		version: version,
		manager: NewScanManager(),
		quit:    quit,
	}
}

func (s *Server) Start(addr string) (string, error) {
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}

	mux := http.NewServeMux()
	s.routes(mux)
	s.http = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Println("webgui server:", err)
		}
	}()

	return "http://" + ln.Addr().String(), nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) routes(mux *http.ServeMux) {
	assets, _ := fs.Sub(assetFS, "assets")
	fileServer := http.FileServer(http.FS(assets))

	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))
	mux.HandleFunc("/api/meta", s.handleMeta)
	mux.HandleFunc("/api/scan/start", s.handleStartScan)
	mux.HandleFunc("/api/scan/stop", s.handleStopScan)
	mux.HandleFunc("/api/scan/state", s.handleScanState)
	mux.HandleFunc("/api/results/save", s.handleSaveResults)
	mux.HandleFunc("/api/app/quit", s.handleQuit)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, assetFS, "assets/index.html")
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "MAGIK",
		"version": s.version,
	})
}

func (s *Server) handleStartScan(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.manager.Start(req); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleStopScan(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	s.manager.Stop()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleScanState(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, s.manager.State())
}

func (s *Server) handleSaveResults(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	path, count, err := s.manager.SaveWorking()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"path":  path,
		"count": count,
	})
}

func (s *Server) handleQuit(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, http.MethodPost) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	go s.quit()
}

func allowMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	return false
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
