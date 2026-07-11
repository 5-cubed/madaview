package config_test

import (
	"path/filepath"
	"testing"

	"github.com/5-cubed/madaview/internal/config"
)

// isolateUserConfigDir points os.UserConfigDir() at a temp directory so
// tests never touch the real machine's config.
func isolateUserConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Setenv("APPDATA", dir)
}

func TestSaveThenLoad_RoundTripsRoot(t *testing.T) {
	isolateUserConfigDir(t)

	if err := config.Save(config.Config{Root: "/some/root"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Root != "/some/root" {
		t.Errorf("Root = %q, want %q", got.Root, "/some/root")
	}
}

func TestLoad_ReturnsZeroValueWhenNoConfigExists(t *testing.T) {
	isolateUserConfigDir(t)

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Root != "" {
		t.Errorf("Root = %q, want empty string", got.Root)
	}
}
