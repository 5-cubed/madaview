package assets_test

import (
	"io/fs"
	"testing"

	"github.com/5-cubed/madaview/internal/assets"
)

func TestFS_ContainsIndexHTML(t *testing.T) {
	data, err := fs.ReadFile(assets.FS, "index.html")
	if err != nil {
		t.Fatalf("ReadFile(index.html): %v", err)
	}
	if len(data) == 0 {
		t.Error("index.html is empty")
	}
}
