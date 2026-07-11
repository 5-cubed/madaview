package rootfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

// TestResolve_FollowsSymlinkPhysicallyUnderRoot verifies the ADR's symlink
// policy: a symlink that physically resides under root is followed even
// when its target points outside root. The root owner's placement of the
// symlink is treated as an explicit choice to expose that target.
func TestResolve_FollowsSymlinkPhysicallyUnderRoot(t *testing.T) {
	outside := t.TempDir()
	target := filepath.Join(outside, "secret.md")
	if err := os.WriteFile(target, []byte("outside content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root := t.TempDir()
	link := filepath.Join(root, "linked.md")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := r.Resolve("linked.md")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	content, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", got, err)
	}
	if string(content) != "outside content" {
		t.Errorf("content = %q, want %q", content, "outside content")
	}
}
