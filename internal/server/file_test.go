package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func TestFile_ReturnsRenderedHTMLAndMetadata(t *testing.T) {
	root, dir := newTestRoot(t)
	if err := os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# My Guide\n\nHello **world**.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	s := newTestServer(t, root, rootfs.SourceDefault)

	req := httptest.NewRequest(http.MethodGet, "/api/file?path=guide.md", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Path  string `json:"path"`
		Title string `json:"title"`
		Mtime string `json:"mtime"`
		HTML  string `json:"html"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v (body: %s)", err, rec.Body.String())
	}
	if got.Path != "guide.md" {
		t.Errorf("Path = %q, want %q", got.Path, "guide.md")
	}
	if got.Title != "My Guide" {
		t.Errorf("Title = %q, want %q", got.Title, "My Guide")
	}
	if got.Mtime == "" {
		t.Errorf("Mtime is empty, want a timestamp")
	}
	if !strings.Contains(got.HTML, "<strong>world</strong>") {
		t.Errorf("HTML = %q, want rendered markdown", got.HTML)
	}
}

func TestFile_PathTraversalRejected(t *testing.T) {
	root, _ := newTestRoot(t)
	s := newTestServer(t, root, rootfs.SourceDefault)

	req := httptest.NewRequest(http.MethodGet, "/api/file?path=../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 or 404 (body: %s)", rec.Code, rec.Body.String())
	}
}
