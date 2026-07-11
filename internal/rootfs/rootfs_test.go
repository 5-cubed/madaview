package rootfs_test

import (
	"os"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func TestResolveInitial_CLIArgTakesPriority(t *testing.T) {
	path, source, err := rootfs.ResolveInitial("/cli/root", "/persisted/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/cli/root" {
		t.Errorf("path = %q, want %q", path, "/cli/root")
	}
	if source != rootfs.SourceCLI {
		t.Errorf("source = %q, want %q", source, rootfs.SourceCLI)
	}
}

func TestResolveInitial_PersistedUsedWhenNoCLIArg(t *testing.T) {
	path, source, err := rootfs.ResolveInitial("", "/persisted/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/persisted/root" {
		t.Errorf("path = %q, want %q", path, "/persisted/root")
	}
	if source != rootfs.SourceUI {
		t.Errorf("source = %q, want %q", source, rootfs.SourceUI)
	}
}

func TestResolveInitial_DefaultsToCWD(t *testing.T) {
	path, source, err := rootfs.ResolveInitial("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	if path != wantCWD {
		t.Errorf("path = %q, want %q", path, wantCWD)
	}
	if source != rootfs.SourceDefault {
		t.Errorf("source = %q, want %q", source, rootfs.SourceDefault)
	}
}
