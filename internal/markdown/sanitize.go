package markdown

import "github.com/microcosm-cc/bluemonday"

// policy is built once and reused: bluemonday's UGCPolicy covers standard
// GFM output (tables, headings, links, images) and already strips
// <script>, inline event handlers ("on*" attributes), and javascript: URLs.
// It's extended to allow the "class" and "data-display" attributes our own
// renderers emit (chroma's syntax-highlighting spans, and the
// mermaid/katex-math hydration placeholders) since UGCPolicy drops unknown
// attributes by default.
var policy = buildPolicy()

func buildPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Globally()
	p.AllowAttrs("data-display").OnElements("div", "span")
	p.AllowAttrs("style").OnElements("span")
	// GFM task-list checkboxes: disabled, non-interactive, so they carry no
	// script risk despite being a form element.
	p.AllowElements("input")
	p.AllowAttrs("type", "checked", "disabled").OnElements("input")
	return p
}

func sanitize(html []byte) string {
	return string(policy.SanitizeBytes(html))
}
