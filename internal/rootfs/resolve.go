package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve validates reqPath against the current root boundary and returns
// the absolute filesystem path it refers to.
//
// reqPath is treated as relative to root regardless of any leading "../" or
// "/" it contains: it is anchored to "/" and cleaned first, so it can never
// lexically escape root. Symlinks are followed normally — a symlink that
// physically resides under root is reachable even if its target lies
// outside root, since the link's placement under root is itself the
// boundary decision.
func (r *Root) Resolve(reqPath string) (string, error) {
	root := r.Current()
	// Anchoring reqPath to the filesystem separator before Cleaning means
	// Clean can never leave it with a leading "..": the shortest path a
	// "../" chain can climb to is the anchor itself, not above it. Joining
	// that anchored, cleaned path onto root therefore can never lexically
	// escape root, regardless of how many "../" segments reqPath contains.
	anchored := filepath.Clean(string(filepath.Separator) + reqPath)
	full := filepath.Join(root, anchored)

	if _, err := os.Stat(full); err != nil {
		return "", fmt.Errorf("rootfs: %q not found: %w", reqPath, err)
	}
	return full, nil
}
