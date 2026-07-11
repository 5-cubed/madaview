package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
	"github.com/5-cubed/madaview/internal/server"
)

func TestSetRoot_SwitchesRootAndReflectsInStatus(t *testing.T) {
	root, _ := newTestRoot(t)
	newDir := t.TempDir()
	s := newTestServer(t, root, rootfs.SourceDefault)

	body, _ := json.Marshal(map[string]string{"root": newDir})
	req := httptest.NewRequest(http.MethodPost, "/api/root", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusRec := httptest.NewRecorder()
	s.Handler().ServeHTTP(statusRec, statusReq)

	var got struct {
		CurrentRoot string `json:"currentRoot"`
		RootSource  string `json:"rootSource"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.CurrentRoot != newDir {
		t.Errorf("CurrentRoot = %q, want %q", got.CurrentRoot, newDir)
	}
	if got.RootSource != rootfs.SourceUI {
		t.Errorf("RootSource = %q, want %q", got.RootSource, rootfs.SourceUI)
	}
}

func TestSetRoot_PersistsViaCallback(t *testing.T) {
	root, _ := newTestRoot(t)
	newDir := t.TempDir()

	var persisted string
	s := server.New(server.Options{
		Root:        root,
		RootSource:  rootfs.SourceDefault,
		Version:     "test",
		Assets:      newTestServerAssets(t),
		Logger:      newTestLogger(),
		PersistRoot: func(path string) error { persisted = path; return nil },
	})

	body, _ := json.Marshal(map[string]string{"root": newDir})
	req := httptest.NewRequest(http.MethodPost, "/api/root", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if persisted != newDir {
		t.Errorf("persisted = %q, want %q", persisted, newDir)
	}
}

func TestSetRoot_RejectsInvalidPath(t *testing.T) {
	root, dir := newTestRoot(t)
	s := newTestServer(t, root, rootfs.SourceDefault)

	body, _ := json.Marshal(map[string]string{"root": "/no/such/path"})
	req := httptest.NewRequest(http.MethodPost, "/api/root", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if root.Current() != dir {
		t.Errorf("root changed after rejected SetRoot: got %q, want %q", root.Current(), dir)
	}
}
