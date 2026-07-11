// Package markdown renders markdown source to sanitized HTML server-side.
// Mermaid diagrams and KaTeX math are never rendered here: goldmark emits
// placeholder markup carrying the raw source, and the bundled frontend
// hydrates it entirely client-side with no network calls.
package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// Raw HTML is enabled because embedding it (e.g. <br>, <details>, <img
// width=...>) is standard GFM practice that authors rely on. Since the
// server is LAN-visible, one viewer's markdown must never be able to run JS
// in another viewer's browser, so every render is passed through
// bluemonday's sanitizer as the actual security boundary — goldmark itself
// is not trusted to produce safe output.
var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, highlightingExtension{}, mathExtension{}),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// Render converts markdown source to sanitized HTML.
func Render(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("markdown: rendering: %w", err)
	}
	return sanitize(buf.Bytes()), nil
}
