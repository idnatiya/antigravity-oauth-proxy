package server

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/dvcrn/antigravity-oauth-proxy/web"
)

var (
	distFS     = web.Dist()
	fileServer = http.StripPrefix("/dashboard/", http.FileServer(http.FS(distFS)))
)

// dashboardUIHandler serves the built Vue 3 dashboard application with HTML5 History fallback.
func (s *Server) dashboardUIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clean path relative to /dashboard
	relPath := strings.TrimPrefix(r.URL.Path, "/dashboard")
	relPath = strings.TrimPrefix(relPath, "/")

	if relPath == "" || relPath == "/" {
		s.serveIndexHTML(w, r)
		return
	}

	// Check if file exists in embedded dist
	f, err := distFS.Open(relPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// SPA HTML5 History Fallback for client-side routes (e.g. /dashboard/overview, /dashboard/requests)
			s.serveIndexHTML(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		s.serveIndexHTML(w, r)
		return
	}

	// File exists; serve it directly
	fileServer.ServeHTTP(w, r)
}

func (s *Server) serveIndexHTML(w http.ResponseWriter, _ *http.Request) {
	f, err := distFS.Open("index.html")
	if err != nil {
		http.Error(w, "Dashboard index not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "Failed to read index HTML", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	_, _ = w.Write(content)
}
