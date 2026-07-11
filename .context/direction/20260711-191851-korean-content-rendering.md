# Korean Content Rendering

## Goal
Korean text inside any markdown document viewed through madaview — prose, headings, fenced code blocks, Mermaid diagram labels, and Korean-named files/folders in the sidebar — displays correctly and legibly across all three target OSes (Windows/macOS/Linux), using the OS's own installed Korean font via CSS/config fallback chains, with no added binary size and no network calls. This is content-rendering fidelity, not UI localization — madaview's own interface strings stay English; only the *documents being viewed* (and their filenames) need to render Korean well.

## Failure Criteria
- Korean glyphs render as tofu boxes (☐) or replacement characters (�) anywhere in scope (body text, headings, code blocks, Mermaid labels, sidebar filenames) on any of the three target OSes.
- Mermaid diagram SVG text renders Korean labels as boxes because mermaid's internal font config wasn't given a Korean-capable fallback (page-level CSS fallback doesn't reach mermaid's SVG text rendering).
- A Korean filename/foldername breaks navigation (garbled in the URL, 404s on click, or mis-decoded on the API round-trip).
- The fix adds a network dependency (e.g. loading a Google Fonts CDN link) — violates the project's zero-external-network-call constraint.
- The new e2e scenario is flaky or trivially always-green (e.g. only checks HTTP 200, not actual glyph presence in the DOM).

## Ambiguous Zone
- On a bare-minimum Linux box with no CJK font package installed, Korean may still render as tofu — this is accepted as an OS-environment gap, not a madaview defect, since the direction explicitly chose OS-font-fallback over bundling a webfont. Not a failure of this feature; out of scope to solve by bundling fonts.
- Fenced code blocks: relying on the browser's automatic per-glyph font substitution (no explicit Korean font added to the monospace font-stack) is accepted as correct behavior, not a partial/incomplete fix — this is standard browser behavior and was deliberately chosen over touching the code-block CSS.
- Rendering fidelity for other CJK languages (Japanese, Chinese) or general internationalization beyond Korean is not the target of this feature, but the font-stack approach chosen doesn't preclude extending it later.

## Direction
Ensure Korean renders correctly wherever markdown content is displayed, using OS-installed Korean fonts as the glyph source (no bundled webfont, no network calls):

1. **Body text & headings**: extend the Tailwind/CSS font-stack in `web/src/index.css` (or wherever the base font-family is declared) to explicitly list Korean-capable system font names ahead of the generic fallback — `"Apple SD Gothic Neo", "Malgun Gothic", "Noto Sans KR"` — before `system-ui, sans-serif`. Applies to the content pane and sidebar (shared global stack), so Korean filenames in the tree benefit from the same change.
2. **Fenced code blocks**: no change. Browsers already substitute a system Korean font per-glyph when the declared monospace font lacks coverage; this is left as default behavior.
3. **Mermaid diagrams**: explicitly pass the same Korean-inclusive font-stack into `mermaid.initialize({ fontFamily: ... })` in the client-side hydration logic (`web/src/hydrate.ts`), since Mermaid's SVG text rendering has its own font config independent of page CSS and won't inherit the body font-stack automatically.
4. **KaTeX**: out of scope — math notation only, no Korean text path exists there.
5. **Sidebar filenames**: no server-side change expected — `encodeURIComponent`/`decodeURIComponent` already round-trip UTF-8 filenames correctly through `/api/tree` and `/api/file` (verified in `web/src/api.ts`); this is purely a font-stack display concern, covered by item 1.
6. **Test coverage**: add a new Playwright e2e test-loop scenario, `e2e/render-korean-content/`, following the existing scenario pattern (own `run`/`verify`/fixture data). Fixture includes Korean prose, a Korean heading, a fenced code block containing Korean text (e.g. a comment), a Mermaid diagram with a Korean-labeled node, and a Korean-named file/folder in the tree. Verify: no tofu/replacement-character codepoints present in rendered text nodes, Mermaid SVG label text renders as real glyphs (not empty/boxes), and the Korean-named file is clickable and its content loads (200, correct title) — not just an HTTP-status check.

## Constraints
- No network calls introduced (consistent with the existing Mermaid/KaTeX local-only rendering requirement).
- No added binary size from bundled font files — glyph coverage comes from the OS's own installed fonts via CSS fallback.
- madaview's own UI strings remain English-only; this feature does not touch i18n/localization of the app chrome.
- Follows the existing e2e test-loop convention (per-scenario `run`/`verify`, fixture data under `e2e/<scenario>/data/`).

## Out of Scope
- UI localization / translating madaview's interface into Korean (rejected — separate feature, not this one).
- Bundling a Korean webfont (e.g. Noto Sans KR) into the binary via `go:embed` (rejected — OS font fallback chosen instead, to keep zero added binary size and zero build-step changes).
- Explicitly touching the fenced-code-block font-stack (rejected — default browser per-glyph fallback already handles this).
- Support for other CJK languages (Japanese, Chinese) or broader i18n — not solved here, though the font-stack approach doesn't block extending it later.
- Fixing Korean glyph rendering on CJK-font-less minimal Linux environments — accepted as an OS-environment gap per the Ambiguous Zone, not something this feature solves via bundling.
