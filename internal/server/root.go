package server

import (
	"encoding/json"
	"net/http"

	"github.com/5-cubed/madaview/internal/rootfs"
)

type setRootRequest struct {
	Root string `json:"root"`
}

func (s *Server) handleSetRoot(w http.ResponseWriter, r *http.Request) {
	var req setRootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	oldRoot := s.root.Current()
	if err := s.root.SetRoot(req.Root); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.rootSource = rootfs.SourceUI
	s.mu.Unlock()

	s.logger.Info("root change", "oldRoot", oldRoot, "newRoot", s.root.Current(), "source", rootfs.SourceUI)

	if s.persistRoot != nil {
		if err := s.persistRoot(s.root.Current()); err != nil {
			s.logger.Warn("failed to persist root", "root", s.root.Current(), "error", err.Error())
		}
	}

	s.handleStatus(w, r)
}
