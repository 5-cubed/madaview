package server_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
	"github.com/5-cubed/madaview/internal/server"
)

func TestHandler_LogsOneLinePerRequest(t *testing.T) {
	root, _ := newTestRoot(t)
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	s := server.New(server.Options{
		Root:       root,
		RootSource: rootfs.SourceDefault,
		Version:    "test",
		Assets:     newTestServerAssets(t),
		Logger:     logger,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	logLine := logBuf.String()
	if !strings.Contains(logLine, "GET") || !strings.Contains(logLine, "/api/status") {
		t.Errorf("log = %q, want it to mention method and path", logLine)
	}
	if !strings.Contains(logLine, "200") {
		t.Errorf("log = %q, want it to mention status 200", logLine)
	}
	if !strings.Contains(logLine, "duration") {
		t.Errorf("log = %q, want it to mention duration", logLine)
	}
}
