# Chroma emits CSS classes for code blocks, but no stylesheet defines them

**Resolved** by `.context/adr/20260712-221004-md-rendering-theme.md` — see
[[20260712-221500-chroma-token-classes-styled-via-css-variables]] for the
fix. The gap this entry originally documented (below) no longer exists.

`internal/markdown/codeblock.go:98` renders fenced code blocks via `chromahtml.New(chromahtml.WithClasses(true))`. `WithClasses(true)` makes Chroma emit semantic class names (`.chroma`, `.kd`, `.nf`, `.p`, `.s`, `.c`, etc.) into the HTML instead of inline styles — the actual colors are supposed to come from a separate CSS stylesheet, generated via Chroma's `formatter.WriteCSS(...)`.

That stylesheet was never written. There's no `WriteCSS` call, no `.chroma` rule in any CSS file (`web/src/index.css`, built `web/dist` assets), and no HTTP route serving one. The `chromaStyle = "github"` constant in the same file is set but has no visible effect — it selects a Chroma style object, but nothing ever turns that style into CSS.

**Net effect (historical)**: code blocks used to render as plain black unstyled text. Tokenization/class-emission worked correctly; only the color mapping was missing.
