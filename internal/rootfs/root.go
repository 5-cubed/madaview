package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Root guards a single root folder that bounds filesystem access, allowing
// it to be swapped at runtime without a restart.
type Root struct {
	mu   sync.RWMutex
	path string
}

// New creates a Root anchored at path, which must exist and be a directory.
func New(path string) (*Root, error) {
	validated, err := validateRoot(path)
	if err != nil {
		return nil, err
	}
	return &Root{path: validated}, nil
}

// Current returns the current root path.
func (r *Root) Current() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.path
}

// SetRoot atomically swaps the root path after validating it exists and is
// a directory. The previous root is left unchanged if validation fails.
func (r *Root) SetRoot(path string) error {
	validated, err := validateRoot(path)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = validated
	return nil
}

func validateRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("rootfs: resolve absolute path for %q: %w", path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("rootfs: root %q does not exist: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("rootfs: root %q is not a directory", path)
	}
	return abs, nil
}
