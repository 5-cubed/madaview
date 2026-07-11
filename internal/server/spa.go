package server

import (
	"io/fs"
	"net/http"
)

// handleSPA serves static frontend assets directly when they exist, and
// falls back to index.html otherwise so React Router can handle
// deep-linkable client-side routes like /view/docs/guide.md.
func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	if name == "/" {
		name = "/index.html"
	}
	if _, err := fs.Stat(s.assets, trimLeadingSlash(name)); err != nil {
		http.ServeFileFS(w, r, s.assets, "index.html")
		return
	}
	http.ServeFileFS(w, r, s.assets, trimLeadingSlash(name))
}

func trimLeadingSlash(p string) string {
	if len(p) > 0 && p[0] == '/' {
		return p[1:]
	}
	return p
}
