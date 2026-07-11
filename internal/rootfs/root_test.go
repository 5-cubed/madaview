package rootfs_test

import (
	"os"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func TestNew_RejectsNonExistentPath(t *testing.T) {
	_, err := rootfs.New("/no/such/path/definitely/not/real")
	if err == nil {
		t.Fatal("expected error for non-existent root, got nil")
	}
}

func TestNew_RejectsFileNotDirectory(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/notadir.txt"
	if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	_, err := rootfs.New(file)
	if err == nil {
		t.Fatal("expected error for file root, got nil")
	}
}

func TestRoot_CurrentReturnsInitialPath(t *testing.T) {
	dir := t.TempDir()
	r, err := rootfs.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if r.Current() != dir {
		t.Errorf("Current() = %q, want %q", r.Current(), dir)
	}
}

func TestRoot_SetRootSwapsCurrentPath(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	r, err := rootfs.New(dir1)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.SetRoot(dir2); err != nil {
		t.Fatalf("SetRoot: %v", err)
	}
	if r.Current() != dir2 {
		t.Errorf("Current() = %q, want %q", r.Current(), dir2)
	}
}

func TestRoot_SetRootRejectsInvalidPath(t *testing.T) {
	dir := t.TempDir()
	r, err := rootfs.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.SetRoot("/no/such/path"); err == nil {
		t.Fatal("expected error for non-existent root, got nil")
	}
	if r.Current() != dir {
		t.Errorf("Current() changed after failed SetRoot: got %q, want %q", r.Current(), dir)
	}
}
