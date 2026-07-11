# Korean font-stack wiring (CSS + Mermaid)

`web/src/index.css` overrides Tailwind v4's `--font-sans` theme variable via an `@theme` block,
prepending `"Apple SD Gothic Neo", "Malgun Gothic", "Noto Sans KR"` ahead of Tailwind's own default
sans stack (`ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji", ...`). This is a single
override point: every element that inherits the default font via preflight, or uses `font-sans`
explicitly, picks it up automatically — no new CSS selectors, no cascade-layer ordering fights.

`web/src/hydrate.ts` reads the resolved stack at runtime via
`getComputedStyle(document.body).fontFamily` and passes it into `mermaid.initialize({ fontFamily })`
before any `mermaid.render()` call. Mermaid's SVG text has its own font config independent of page
CSS and won't inherit the body font-stack automatically — this keeps CSS as the single source of
truth so the two can't drift apart.

See [[20260711-191934-cjk-font-strategy-os-fallback]] for why OS-fallback (not bundling) was chosen,
and [[20260711-223400-browser-font-fallback-defeats-tofu-detection]] for a hard limitation
discovered while testing this.
