package server_test

import (
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/5-cubed/madaview/internal/rootfs"
	"github.com/5-cubed/madaview/internal/server"
)

func newTestServer(t *testing.T, root *rootfs.Root, rootSource string) *server.Server {
	t.Helper()
	return server.New(server.Options{
		Root:       root,
		RootSource: rootSource,
		Version:    "test-version",
		Assets:     newTestServerAssets(t),
		Logger:     newTestLogger(),
	})
}

func newTestServerAssets(t *testing.T) fs.FS {
	t.Helper()
	return fstest.MapFS{}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRoot(t *testing.T) (*rootfs.Root, string) {
	t.Helper()
	dir := t.TempDir()
	r, err := rootfs.New(dir)
	if err != nil {
		t.Fatalf("rootfs.New: %v", err)
	}
	return r, dir
}

func TestStatus_ReturnsServerState(t *testing.T) {
	root, dir := newTestRoot(t)
	s := newTestServer(t, root, rootfs.SourceCLI)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Version     string `json:"version"`
		CurrentRoot string `json:"currentRoot"`
		RootSource  string `json:"rootSource"`
		GoVersion   string `json:"goVersion"`
		OS          string `json:"os"`
		Arch        string `json:"arch"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v (body: %s)", err, rec.Body.String())
	}
	if got.Version != "test-version" {
		t.Errorf("Version = %q, want %q", got.Version, "test-version")
	}
	if got.CurrentRoot != dir {
		t.Errorf("CurrentRoot = %q, want %q", got.CurrentRoot, dir)
	}
	if got.RootSource != rootfs.SourceCLI {
		t.Errorf("RootSource = %q, want %q", got.RootSource, rootfs.SourceCLI)
	}
	if got.GoVersion == "" || got.OS == "" || got.Arch == "" {
		t.Errorf("expected non-empty GoVersion/OS/Arch, got %+v", got)
	}
}
