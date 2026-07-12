# Browser system font fallback defeats client-side tofu detection

Browsers (verified on Chromium/macOS) perform unconditional, per-script system font fallback for
both canvas and DOM text, and this cannot be suppressed from web content. Tested three ways:

1. Naming a real font known to lack Hangul glyphs (`"Courier New"`) instead of the actual Korean
   stack.
2. Naming a font family that doesn't exist at all (no generic fallback keyword either).
3. A custom `@font-face` with a hard `unicode-range: U+0020-007E` (ASCII-only) restriction, which
   per the CSS Fonts spec should exclude the font from being used for any codepoint outside that
   range.

All three rendered **byte-identical pixels** (same bounding box, same ink count, same raw pixel
data) to the actual Korean font, for Korean text, once any Korean-capable font exists on the
machine.

**Implication:** any client-side "compare rendering against a font known to lack glyph coverage"
technique — `canvas.measureText()` width divergence, or pixel bounding-box comparison — cannot
detect tofu/notdef rendering on a machine that has a Korean font installed. This includes CI once
`fonts-nanum` is installed specifically to enable Korean rendering. Such a technique can only ever
fire in the "no CJK font anywhere" case, which is an accepted non-failure (see the Ambiguous Zone
in the Korean-content-rendering direction/ADR) — making it structurally incapable of catching the
failure it's meant to catch.

A theoretically sounder alternative (not implemented, scoped out as unnecessary complexity) would
be comparing the actual text's rendered ink pattern against a Private Use Area codepoint (e.g.
U+E000) rendered in the same font — PUA codepoints have no script to search fallback for, so they
reliably produce a real notdef/tofu box. This was validated informally but not built into the test
suite; see [[20260711-223405-render-korean-content-structural-checks-only]] for what was actually
shipped instead.
