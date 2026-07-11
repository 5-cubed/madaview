package rootfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/5-cubed/madaview/internal/rootfs"
)

func TestList_SingleLevelListing(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "guide.md"), "# guide")
	mustMkdir(t, filepath.Join(root, "docs"))
	mustWrite(t, filepath.Join(root, "docs", "nested.md"), "# nested")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := r.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2 (got %+v)", len(entries), entries)
	}
	byName := map[string]rootfs.Entry{}
	for _, e := range entries {
		byName[e.Name] = e
	}
	if e, ok := byName["guide.md"]; !ok || e.IsDir {
		t.Errorf("guide.md entry = %+v, want file entry", e)
	}
	if e, ok := byName["docs"]; !ok || !e.IsDir {
		t.Errorf("docs entry = %+v, want dir entry", e)
	}
	// Single-level: nested.md must not appear at the top level.
	if _, ok := byName["nested.md"]; ok {
		t.Errorf("List(\"\") leaked nested entry; want single-level only")
	}
}

func TestList_ExcludesNonMarkdownFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "guide.md"), "# guide")
	mustWrite(t, filepath.Join(root, ".env"), "SECRET=1")
	mustWrite(t, filepath.Join(root, "package.json"), "{}")
	mustWrite(t, filepath.Join(root, "image.png"), "not-a-real-png")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := r.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	byName := map[string]rootfs.Entry{}
	for _, e := range entries {
		byName[e.Name] = e
	}
	if _, ok := byName["guide.md"]; !ok {
		t.Errorf("guide.md missing from entries, want kept")
	}
	for _, excluded := range []string{".env", "package.json", "image.png"} {
		if _, ok := byName[excluded]; ok {
			t.Errorf("%s present in entries, want excluded (non-markdown)", excluded)
		}
	}
}

func TestList_KeepsMarkdownCaseInsensitively(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "README.MD"), "# readme")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := r.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(entries) != 1 || entries[0].Name != "README.MD" {
		t.Errorf("entries = %+v, want [README.MD] (case-insensitive match)", entries)
	}
}

func TestList_ExcludesDeadEndDirectories(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "guide.md"), "# guide")
	mustMkdir(t, filepath.Join(root, "empty-assets"))
	mustWrite(t, filepath.Join(root, "empty-assets", "image.png"), "not-a-real-png")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := r.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	byName := map[string]rootfs.Entry{}
	for _, e := range entries {
		byName[e.Name] = e
	}
	if _, ok := byName["guide.md"]; !ok {
		t.Errorf("guide.md missing from entries, want kept")
	}
	if _, ok := byName["empty-assets"]; ok {
		t.Errorf("empty-assets present in entries, want excluded (no markdown in subtree)")
	}
}

func TestList_KeepsDirectoryWithDeeplyNestedMarkdown(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "deeply", "nested", "dir"))
	mustWrite(t, filepath.Join(root, "deeply", "nested", "dir", "deep.md"), "# deep")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	entries, err := r.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(entries) != 1 || entries[0].Name != "deeply" || !entries[0].IsDir {
		t.Errorf("entries = %+v, want [deeply (dir)] (multi-level recursion must reach deep.md)", entries)
	}
}

func TestList_RejectsListingAFile(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "guide.md"), "# guide")

	r, err := rootfs.New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := r.List("guide.md"); err == nil {
		t.Fatal("expected error listing a file path, got nil")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q): %v", path, err)
	}
}
