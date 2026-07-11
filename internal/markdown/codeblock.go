package markdown

import (
	"bytes"
	gohtml "html"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const chromaStyle = "github"

// codeBlockRenderer overrides goldmark's default fenced-code-block
// rendering. A "mermaid" fence, or a "plantuml"/"puml"/"uml" fence, bypasses
// highlighting entirely and instead emits a placeholder div carrying its
// raw, escaped source for client-side hydration (see the package doc
// comment). Every other fence is highlighted server-side via chroma.
type codeBlockRenderer struct{}

func (r *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.render)
}

func (r *codeBlockRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*ast.FencedCodeBlock)

	var language string
	if node.Info != nil {
		info := node.Info.Segment.Value(source)
		if fields := strings.Fields(string(info)); len(fields) > 0 {
			language = fields[0]
		}
	}

	var code bytes.Buffer
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		code.Write(line.Value(source))
	}

	if language == "mermaid" {
		writeMermaidPlaceholder(w, code.Bytes())
		return ast.WalkContinue, nil
	}
	if isPlantumlFence(language) {
		writePlantumlPlaceholder(w, code.Bytes())
		return ast.WalkContinue, nil
	}
	writeHighlightedCode(w, language, code.Bytes())
	return ast.WalkContinue, nil
}

func writeMermaidPlaceholder(w util.BufWriter, code []byte) {
	_, _ = w.WriteString(`<div class="mermaid">`)
	_, _ = w.WriteString(gohtml.EscapeString(string(code)))
	_, _ = w.WriteString("</div>\n")
}

// isPlantumlFence reports whether language is one of the recognized
// PlantUML fence aliases: plantuml, puml, and uml all render identically.
func isPlantumlFence(language string) bool {
	return language == "plantuml" || language == "puml" || language == "uml"
}

func writePlantumlPlaceholder(w util.BufWriter, code []byte) {
	_, _ = w.WriteString(`<div class="plantuml">`)
	_, _ = w.WriteString(gohtml.EscapeString(string(code)))
	_, _ = w.WriteString("</div>\n")
}

func writeHighlightedCode(w util.BufWriter, language string, code []byte) {
	lexer := lexers.Fallback
	if language != "" {
		if l := lexers.Get(language); l != nil {
			lexer = l
		}
	}
	style := styles.Get(chromaStyle)
	if style == nil {
		style = styles.Fallback
	}
	iterator, err := lexer.Tokenise(nil, string(code))
	if err != nil {
		_, _ = w.WriteString("<pre><code>")
		_, _ = w.WriteString(gohtml.EscapeString(string(code)))
		_, _ = w.WriteString("</code></pre>\n")
		return
	}
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	_ = formatter.Format(w, style, iterator)
}

// highlightingExtension wires codeBlockRenderer into goldmark, overriding
// the default fenced-code-block renderer.
type highlightingExtension struct{}

func (highlightingExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&codeBlockRenderer{}, 100),
	))
}
