# `--text-muted` and `--border` are not safe as-is for WCAG-gated `.prose` content

When wiring `@tailwindcss/typography`'s `--tw-prose-*` vars to theme tokens (see
[[20260712-232400-tw-prose-vars-unwired-to-theme-tokens]]), the naive approach — reuse
`--text-muted` for secondary/muted-looking content and `--border` for divider-ish content — fails
WCAG AA when checked per theme with the actual relative-luminance contrast formula:

| theme | `--text` vs `--bg` | `--text-muted` vs `--bg` | `--border` vs `--bg` |
|---|---|---|---|
| github-light | 15.80 | 6.11 | 1.45 |
| github-dark | 16.02 | 6.31 | 1.55 |
| kanagawa | 11.26 | 5.91 | 2.23 |
| gruvbox | 10.75 | 5.30 | 1.67 |
| solarized-light | 12.05 | 4.13 | 1.50 |
| one-dark | 6.57 | 2.32 | 1.43 |

- **`--text-muted` fails the 4.5:1 "normal text" bar in solarized-light (4.13) and fails even the
  lighter 3:1 "large text/marker" bar in one-dark (2.32)**. It cannot safely back any
  `.prose` var that gets a WCAG contrast check (body, bold, quotes, captions, code, pre-code,
  bullets, counters) — use `--text` instead for anything that must clear a ratio gate.
- **`--border` fails 3:1 in every theme (1.43–2.23)**. This is by design, not a bug: hairline
  dividers (`<hr>`, blockquote borders, table cell borders) are conventionally low-contrast, and
  forcing 3:1 here would make every divider read as a heavy rule instead of a subtle line. Only
  ever check these for "wired/non-default", never gate them on a contrast ratio.

**Takeaway for future theme additions**: when adding a 7th theme (or auditing an existing one),
`--text-muted` and `--border` must independently satisfy these same bars if they're ever reused
for WCAG-gated content — don't assume "the token already looks fine" is equivalent to "the token
clears 4.5:1/3:1." Compute the ratio against that theme's `--bg` explicitly
(`e2e/lib/contrast.mjs`'s `contrastRatio` implements the formula used for this table).
