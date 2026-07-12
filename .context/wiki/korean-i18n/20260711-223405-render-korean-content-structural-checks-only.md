# render-korean-content e2e scenario checks structure, not glyph fidelity

`e2e/render-korean-content/` does not attempt pixel- or metric-based glyph-fidelity comparison —
see [[20260711-223400-browser-font-fallback-defeats-tofu-detection]] for why that class of
technique can't work. Instead `verify` checks only what's actually detectable client-side:

- No U+FFFD anywhere in the rendered DOM (`dom.html` / `nested-dom.html`).
- Each of the four text surfaces (heading, prose, code block, Mermaid label) has non-empty
  `textContent` and a non-zero `getBoundingClientRect()` — catches silently-empty or collapsed
  rendering (e.g. Mermaid swallowing an error, or a fetch never populating the element), regardless
  of which specific glyphs got drawn.
- The nested Korean-named file/folder navigates correctly end-to-end.
- Zero non-`localhost` requests fire during the whole flow.

Real glyph-vs-tofu fidelity was confirmed once by hand via a Playwright screenshot during
development, not re-verified per CI run.

**Gotcha:** the nested-navigation check must `waitForFunction` on the `<h1>` text actually changing
(`document.querySelector('article h1')?.textContent.trim() === expected`), not `waitForSelector`
on `article h1` alone — the `article`/`h1` DOM nodes persist across client-side route changes (only
their content is replaced via `dangerouslySetInnerHTML`), so `waitForSelector` resolves instantly
against the *previous* file's stale content and races the `fetchFile` call for the new one.
