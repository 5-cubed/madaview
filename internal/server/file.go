package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/5-cubed/madaview/internal/markdown"
)

type fileResponse struct {
	Path  string `json:"path"`
	Title string `json:"title"`
	Mtime string `json:"mtime"`
	HTML  string `json:"html"`
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Query().Get("path")
	full, err := s.root.Resolve(reqPath)
	if err != nil {
		s.logger.Warn("path-safety rejection",
			"endpoint", "/api/file", "requestedPath", reqPath, "reason", err.Error())
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	source, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	html, err := markdown.Render(source)
	if err != nil {
		http.Error(w, "failed to render markdown", http.StatusInternalServerError)
		return
	}

	resp := fileResponse{
		Path:  reqPath,
		Title: extractTitle(source, reqPath),
		Mtime: info.ModTime().UTC().Format(time.RFC3339),
		HTML:  html,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// extractTitle uses the first ATX H1 heading ("# Title") as the document
// title, falling back to the filename when none is present.
func extractTitle(source []byte, reqPath string) string {
	for _, line := range strings.Split(string(source), "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, "# "); ok {
			return strings.TrimSpace(after)
		}
	}
	return path.Base(reqPath)
}
