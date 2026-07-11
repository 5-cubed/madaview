package rootfs

import (
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
)

// Entry describes one child of a listed directory.
type Entry struct {
	Name  string
	Path  string
	IsDir bool
}

// List returns the single-level listing of the directory at reqPath,
// relative to the current root, keeping only markdown files and
// directories whose subtree contains a markdown file somewhere below them.
func (r *Root) List(reqPath string) ([]Entry, error) {
	rawEntries, err := r.readDir(reqPath)
	if err != nil {
		return nil, err
	}

	visited := map[string]struct{}{}
	entries := make([]Entry, 0, len(rawEntries))
	for _, e := range rawEntries {
		if !e.IsDir {
			if !isMarkdownName(e.Name) {
				continue
			}
			entries = append(entries, e)
			continue
		}
		has, err := r.subtreeHasMarkdown(e.Path, visited)
		if err != nil {
			return nil, err
		}
		if !has {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// readDir returns the raw, unfiltered single-level listing of reqPath: every
// child entry, with no markdown or dead-end-directory filtering applied.
// Both List and subtreeHasMarkdown build on this so path-safety (Resolve)
// and symlink-aware IsDir resolution live in exactly one place.
func (r *Root) readDir(reqPath string) ([]Entry, error) {
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

// isMarkdownName reports whether name has a recognized markdown extension
// (.md, .markdown, .mdx), matched case-insensitively.
func isMarkdownName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown", ".mdx":
		return true
	default:
		return false
	}
}

// subtreeHasMarkdown reports whether reqPath's subtree contains a markdown
// file anywhere below it. visited tracks the real (symlink-resolved) paths
// already descended into during a single top-level List call, so a symlink
// that loops back to one of its own ancestors is treated as a dead end
// instead of recursing forever.
func (r *Root) subtreeHasMarkdown(reqPath string, visited map[string]struct{}) (bool, error) {
	full, err := r.Resolve(reqPath)
	if err != nil {
		return false, err
	}
	real, err := filepath.EvalSymlinks(full)
	if err != nil {
		return false, err
	}
	if _, ok := visited[real]; ok {
		return false, nil
	}
	visited[real] = struct{}{}

	children, err := r.readDir(reqPath)
	if err != nil {
		return false, err
	}
	for _, c := range children {
		if !c.IsDir {
			if isMarkdownName(c.Name) {
				return true, nil
			}
			continue
		}
		has, err := r.subtreeHasMarkdown(c.Path, visited)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}
