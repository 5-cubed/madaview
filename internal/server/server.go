// Package server wires madaview's HTTP API and static SPA serving on top of
// internal/rootfs and internal/markdown.
package server

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime"
	"sync"

	"github.com/5-cubed/madaview/internal/rootfs"
)

// Options configures a new Server.
type Options struct {
	Root       *rootfs.Root
	RootSource string
	Version    string
	Assets     fs.FS
	Logger     *slog.Logger
	// PersistRoot, if set, is called after a successful runtime root change
	// (POST /api/root) so the new root survives the next launch. It is
	// deliberately decoupled from internal/config here to keep this package
	// independent of how (or whether) persistence works.
	PersistRoot func(path string) error
}

// Server serves madaview's JSON API and embedded SPA assets.
type Server struct {
	root        *rootfs.Root
	version     string
	assets      fs.FS
	logger      *slog.Logger
	persistRoot func(path string) error

	mu         sync.RWMutex
	rootSource string

	mux http.Handler
}

// New builds a Server ready to handle requests via Handler().
func New(opts Options) *Server {
	s := &Server{
		root:        opts.Root,
		version:     opts.Version,
		assets:      opts.Assets,
		logger:      opts.Logger,
		persistRoot: opts.PersistRoot,
		rootSource:  opts.RootSource,
	}
	s.mux = s.loggingMiddleware(s.routes())
	return s
}

// Handler returns the http.Handler serving the full API and SPA.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/tree", s.handleTree)
	mux.HandleFunc("GET /api/file", s.handleFile)
	mux.HandleFunc("POST /api/root", s.handleSetRoot)
	mux.HandleFunc("/", s.handleSPA)
	return mux
}

type statusResponse struct {
	Version     string `json:"version"`
	CurrentRoot string `json:"currentRoot"`
	RootSource  string `json:"rootSource"`
	GoVersion   string `json:"goVersion"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	source := s.rootSource
	s.mu.RUnlock()

	resp := statusResponse{
		Version:     s.version,
		CurrentRoot: s.root.Current(),
		RootSource:  source,
		GoVersion:   runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
