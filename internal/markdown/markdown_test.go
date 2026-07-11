package markdown_test

import (
	"strings"
	"testing"

	"github.com/5-cubed/madaview/internal/markdown"
)

func TestRender_BasicParagraph(t *testing.T) {
	html, err := markdown.Render([]byte("Hello **world**"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, "<p>Hello <strong>world</strong></p>") {
		t.Errorf("html = %q, want it to contain a <p> with <strong>world</strong>", html)
	}
}

func TestRender_GFMTable(t *testing.T) {
	source := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	html, err := markdown.Render([]byte(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, "<table>") {
		t.Errorf("html = %q, want it to contain a <table>", html)
	}
}

func TestRender_GFMTaskList(t *testing.T) {
	source := "- [ ] todo\n- [x] done\n"
	html, err := markdown.Render([]byte(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, `<input disabled="" type="checkbox">`) &&
		!strings.Contains(html, `<input checked="" disabled="" type="checkbox">`) {
		t.Errorf("html = %q, want it to contain disabled checkbox inputs", html)
	}
}

func TestRender_FencedCodeBlockIsSyntaxHighlighted(t *testing.T) {
	source := "```go\nfunc main() {}\n```\n"
	html, err := markdown.Render([]byte(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, "<pre") || !strings.Contains(html, "chroma") {
		t.Errorf("html = %q, want a chroma-highlighted <pre> block", html)
	}
}

func TestRender_MermaidFenceEmitsPlaceholderWithRawSource(t *testing.T) {
	source := "```mermaid\ngraph TD;\nA-->B;\n```\n"
	html, err := markdown.Render([]byte(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, `<div class="mermaid">`) {
		t.Errorf("html = %q, want a mermaid placeholder div", html)
	}
	if !strings.Contains(html, "graph TD;") || !strings.Contains(html, "A--&gt;B;") {
		t.Errorf("html = %q, want it to contain the raw (escaped) mermaid source", html)
	}
	if strings.Contains(html, "<svg") || strings.Contains(html, "chroma") {
		t.Errorf("html = %q, mermaid source must not be highlighted or pre-rendered server-side", html)
	}
}

func TestRender_PlantumlFenceEmitsPlaceholderWithRawSource(t *testing.T) {
	for _, lang := range []string{"plantuml", "puml", "uml"} {
		t.Run(lang, func(t *testing.T) {
			source := "```" + lang + "\nAlice -> Bob: hello\n```\n"
			html, err := markdown.Render([]byte(source))
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.Contains(html, `<div class="plantuml">`) {
				t.Errorf("html = %q, want a plantuml placeholder div", html)
			}
			if !strings.Contains(html, "Alice -&gt; Bob: hello") {
				t.Errorf("html = %q, want it to contain the raw (escaped) plantuml source", html)
			}
			if strings.Contains(html, "<svg") || strings.Contains(html, "chroma") {
				t.Errorf("html = %q, plantuml source must not be highlighted or pre-rendered server-side", html)
			}
		})
	}
}

func TestRender_BlockMathEmitsPlaceholderWithRawSource(t *testing.T) {
	source := "$$\nx^2 + y^2 = z^2\n$$\n"
	html, err := markdown.Render([]byte(source))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := "<div class=\"katex-math\" data-display=\"true\">x^2 + y^2 = z^2\n</div>\n"
	if html != want {
		t.Errorf("html = %q, want exactly %q (no duplicated/leaked content)", html, want)
	}
}

func TestRender_InlineMathEmitsPlaceholderWithRawSource(t *testing.T) {
	html, err := markdown.Render([]byte("Einstein: $E = mc^2$ is famous.\n"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(html, `class="katex-math"`) || !strings.Contains(html, `data-display="false"`) {
		t.Errorf("html = %q, want an inline katex-math placeholder with data-display=false", html)
	}
	if !strings.Contains(html, "E = mc^2") {
		t.Errorf("html = %q, want it to contain the raw math source", html)
	}
}

func TestRender_DollarPricesAreNotTreatedAsMath(t *testing.T) {
	html, err := markdown.Render([]byte("It costs $5 and $10.\n"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(html, "katex-math") {
		t.Errorf("html = %q, plain currency amounts must not become math placeholders", html)
	}
}

func TestRender_SanitizesRawScriptTags(t *testing.T) {
	html, err := markdown.Render([]byte("<script>alert('xss')</script>\n\nSafe text.\n"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(html, "<script") {
		t.Errorf("html = %q, want <script> stripped", html)
	}
	if !strings.Contains(html, "Safe text.") {
		t.Errorf("html = %q, want surrounding safe content preserved", html)
	}
}

func TestRender_SanitizesInlineEventHandlers(t *testing.T) {
	html, err := markdown.Render([]byte(`<img src="x" onerror="alert(1)">` + "\n"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(html, "onerror") {
		t.Errorf("html = %q, want onerror attribute stripped", html)
	}
}
