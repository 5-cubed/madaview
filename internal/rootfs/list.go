package rootfs

import (
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
)

// Entry describes one child of a listed directory.
type Entry struct {
	Name  string
	Path  string
	IsDir bool
}

// List returns the single-level (non-recursive) listing of the directory at
// reqPath, relative to the current root.
func (r *Root) List(reqPath string) ([]Entry, error) {
	full, err := r.Resolve(reqPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return nil, fmt.Errorf("rootfs: %q not found: %w", reqPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("rootfs: %q is not a directory", reqPath)
	}

	dirEntries, err := os.ReadDir(full)
	if err != nil {
		return nil, fmt.Errorf("rootfs: reading %q: %w", reqPath, err)
	}

	entries := make([]Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		isDir := de.IsDir()
		if de.Type()&os.ModeSymlink != 0 {
			if target, err := os.Stat(filepath.Join(full, de.Name())); err == nil {
				isDir = target.IsDir()
			}
		}
		entries = append(entries, Entry{
			Name:  de.Name(),
			Path:  pathpkg.Join(reqPath, de.Name()),
			IsDir: isDir,
		})
	}
	return entries, nil
}
