# CJK font strategy: OS fallback over bundling

For Korean (and, by precedent, future CJK/i18n) glyph rendering, madaview deliberately relies on the end user's OS-installed system fonts via CSS font-stack fallback (e.g. `"Apple SD Gothic Neo", "Malgun Gothic", "Noto Sans KR", system-ui, sans-serif`) and equivalent config passed into `mermaid.initialize({ fontFamily })` for SVG diagram text — rather than bundling a webfont (e.g. Noto Sans KR) into the binary via `go:embed`.

**Why:** keeps zero added binary size and avoids a new font-asset build pipeline step, consistent with the project's "single self-contained binary, no bloat" posture. Also avoids introducing a font-licensing decision.

**Tradeoff accepted:** a bare-minimum Linux box with no CJK font package installed may still show tofu boxes for Korean text. This is treated as an OS-environment gap, not a madaview defect — it does not violate the zero-network-call constraint (no CDN font loading either) and doesn't warrant bundling.

**Precedent:** if other CJK or broader i18n rendering needs come up later, default to extending this same OS-fallback font-stack approach rather than switching to bundled webfonts, unless the tradeoff above becomes actually unacceptable (e.g. a stated requirement to support minimal/headless Linux without pre-installed fonts).
