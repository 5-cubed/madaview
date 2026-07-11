package server

import (
	"encoding/json"
	"net/http"
)

type treeEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Query().Get("path")
	entries, err := s.root.List(reqPath)
	if err != nil {
		s.logger.Warn("path-safety rejection",
			"endpoint", "/api/tree", "requestedPath", reqPath, "reason", err.Error())
		http.NotFound(w, r)
		return
	}

	resp := make([]treeEntry, len(entries))
	for i, e := range entries {
		resp[i] = treeEntry{Name: e.Name, Path: e.Path, IsDir: e.IsDir}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
