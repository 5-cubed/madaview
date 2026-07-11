package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/5-cubed/madaview/internal/rootfs"
	"github.com/5-cubed/madaview/internal/server"
)

func TestSPA_ServesStaticAssetDirectly(t *testing.T) {
	root, _ := newTestRoot(t)
	assets := fstest.MapFS{
		"index.html":    {Data: []byte("<html>index</html>")},
		"assets/app.js": {Data: []byte("console.log('app')")},
	}
	s := server.New(server.Options{
		Root: root, RootSource: rootfs.SourceDefault, Version: "test",
		Assets: assets, Logger: newTestLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != "console.log('app')" {
		t.Errorf("body = %q, want the raw asset content", rec.Body.String())
	}
}

func TestSPA_FallsBackToIndexForClientRoutes(t *testing.T) {
	root, _ := newTestRoot(t)
	assets := fstest.MapFS{
		"index.html": {Data: []byte("<html>index</html>")},
	}
	s := server.New(server.Options{
		Root: root, RootSource: rootfs.SourceDefault, Version: "test",
		Assets: assets, Logger: newTestLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/view/docs/guide.md", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != "<html>index</html>" {
		t.Errorf("body = %q, want index.html content as SPA fallback", rec.Body.String())
	}
}
