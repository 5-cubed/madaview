package rootfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func newRootWithFile(t *testing.T, relPath, content string) (*rootfs.Root, string) {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	r, err := rootfs.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r, dir
}

func TestResolve_ValidPathWithinRoot(t *testing.T) {
	r, dir := newRootWithFile(t, "docs/guide.md", "# guide")
	got, err := r.Resolve("docs/guide.md")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(dir, "docs", "guide.md")
	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestResolve_RejectsDotDotTraversal(t *testing.T) {
	r, _ := newRootWithFile(t, "docs/guide.md", "# guide")
	if _, err := r.Resolve("../../../etc/passwd"); err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
}

func TestResolve_RejectsAbsolutePathEscape(t *testing.T) {
	r, _ := newRootWithFile(t, "docs/guide.md", "# guide")
	if _, err := r.Resolve("/etc/passwd"); err == nil {
		t.Fatal("expected error for absolute path escape, got nil")
	}
}

func TestResolve_RejectsNonExistentPath(t *testing.T) {
	r, _ := newRootWithFile(t, "docs/guide.md", "# guide")
	if _, err := r.Resolve("docs/missing.md"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
