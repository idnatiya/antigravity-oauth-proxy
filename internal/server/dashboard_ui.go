package server

import (
	_ "embed"
	"net/http"
)

//go:embed dashboard/index.html
var dashboardIndexHTML []byte

// dashboardUIHandler serves the single-page dashboard application.
func (s *Server) dashboardUIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(dashboardIndexHTML)
}
