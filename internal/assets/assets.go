// Package assets embeds the built frontend (web/dist, copied here by the
// Makefile's build target) into the madaview binary.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS serves the embedded frontend build, rooted so paths like "index.html"
// resolve directly.
var FS = mustSub(distFS, "dist")

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		// dist is embedded at compile time via go:embed; a failure here
		// means the embed directive itself is broken, not a runtime
		// condition callers can recover from.
		panic(err)
	}
	return sub
}
