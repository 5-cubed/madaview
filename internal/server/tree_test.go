package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func TestTree_ListsSingleLevel(t *testing.T) {
	root, dir := newTestRoot(t)
	if err := os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# guide"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	s := newTestServer(t, root, rootfs.SourceDefault)

	req := httptest.NewRequest(http.MethodGet, "/api/tree?path=", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []struct {
		Name  string `json:"name"`
		Path  string `json:"path"`
		IsDir bool   `json:"isDir"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v (body: %s)", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("len(entries) = %d, want 2 (got %+v)", len(got), got)
	}
}

func TestTree_PathTraversalRejected(t *testing.T) {
	root, _ := newTestRoot(t)
	s := newTestServer(t, root, rootfs.SourceDefault)

	req := httptest.NewRequest(http.MethodGet, "/api/tree?path=../../../etc", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 or 404 (body: %s)", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() > 0 && rec.Code == http.StatusOK {
		t.Fatalf("unexpected 200 with body: %s", rec.Body.String())
	}
}
