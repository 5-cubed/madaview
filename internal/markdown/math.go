package markdown

import (
	"bytes"
	gohtml "html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindMathBlock identifies a block-math AST node ($$ ... $$ on its own
// lines). Its raw source is emitted verbatim for client-side KaTeX
// hydration; it is never rendered to math server-side.
var KindMathBlock = ast.NewNodeKind("MathBlock")

type mathBlock struct {
	ast.BaseBlock
}

func newMathBlock() *mathBlock {
	return &mathBlock{}
}

func (n *mathBlock) Kind() ast.NodeKind { return KindMathBlock }

// IsRaw marks this block's lines as opaque to goldmark's core parser, the
// same way ast.FencedCodeBlock does. Without it, the parser would treat the
// unprocessed lines as re-parseable paragraph content and render them a
// second time in addition to mathBlockRenderer's own output.
func (n *mathBlock) IsRaw() bool { return true }

func (n *mathBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// mathBlockParser parses a block delimited by "$$" alone on its own line at
// the start and end, mirroring goldmark's fenced-code-block parser but with
// a fixed two-character delimiter instead of a variable-length fence.
type mathBlockParser struct{}

func (b *mathBlockParser) Trigger() []byte { return []byte{'$'} }

func (b *mathBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if !bytes.Equal(bytes.TrimSpace(line), []byte("$$")) {
		return nil, parser.NoChildren
	}
	reader.AdvanceToEOL()
	return newMathBlock(), parser.NoChildren
}

func (b *mathBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if bytes.Equal(bytes.TrimSpace(line), []byte("$$")) {
		reader.AdvanceToEOL()
		return parser.Close
	}
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (b *mathBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *mathBlockParser) CanInterruptParagraph() bool { return true }

func (b *mathBlockParser) CanAcceptIndentedLine() bool { return false }

// mathBlockRenderer renders a mathBlock node to a katex-math placeholder
// carrying its raw, HTML-escaped source for client-side hydration.
type mathBlockRenderer struct{}

func (r *mathBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathBlock, r.render)
}

func (r *mathBlockRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(source))
	}
	_, _ = w.WriteString(`<div class="katex-math" data-display="true">`)
	_, _ = w.WriteString(gohtml.EscapeString(buf.String()))
	_, _ = w.WriteString("</div>\n")
	return ast.WalkContinue, nil
}

// KindMathInline identifies an inline-math AST node ($ ... $). Its raw
// source is emitted verbatim for client-side KaTeX hydration.
var KindMathInline = ast.NewNodeKind("MathInline")

type mathInline struct {
	ast.BaseInline
	Segment text.Segment
}

func (n *mathInline) Kind() ast.NodeKind { return KindMathInline }

func (n *mathInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Content": string(n.Segment.Value(source))}, nil)
}

// mathInlineParser parses "$...$" delimited inline math, mirroring the
// emphasis-parser rule that the content must not start or end with
// whitespace immediately inside the delimiters — this is what keeps plain
// currency text like "$5 and $10" from being misread as math: the space
// right after the first "$" fails the rule.
type mathInlineParser struct{}

func (p *mathInlineParser) Trigger() []byte { return []byte{'$'} }

func (p *mathInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	if len(line) < 3 || line[1] == ' ' || line[1] == '$' {
		return nil
	}
	for i := 1; i < len(line); i++ {
		if line[i] != '$' {
			continue
		}
		if line[i-1] == ' ' {
			continue
		}
		content := segment.WithStart(segment.Start + 1)
		content = content.WithStop(segment.Start + i)
		block.Advance(i + 1)
		return &mathInline{Segment: content}
	}
	return nil
}

// mathInlineRenderer renders a mathInline node to a katex-math placeholder
// carrying its raw, HTML-escaped source for client-side hydration.
type mathInlineRenderer struct{}

func (r *mathInlineRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.render)
}

func (r *mathInlineRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*mathInline)
	_, _ = w.WriteString(`<span class="katex-math" data-display="false">`)
	_, _ = w.WriteString(gohtml.EscapeString(string(node.Segment.Value(source))))
	_, _ = w.WriteString(`</span>`)
	return ast.WalkContinue, nil
}

// mathExtension wires the block- and inline-math parsers and renderers into
// goldmark.
type mathExtension struct{}

func (mathExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&mathBlockParser{}, 100),
		),
		parser.WithInlineParsers(
			util.Prioritized(&mathInlineParser{}, 100),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&mathBlockRenderer{}, 100),
		util.Prioritized(&mathInlineRenderer{}, 100),
	))
}
