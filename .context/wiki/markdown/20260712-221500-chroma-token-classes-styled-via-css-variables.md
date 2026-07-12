# Chroma token classes are styled via ~8 attribute-prefix CSS rules, not per-class

`internal/markdown/codeblock.go` (`chromahtml.New(chromahtml.WithClasses(true))`) emits ~80 distinct token classes, but they all follow Pygments' short-code convention: the first letter of the class name names the category — `k*` = keyword (`kd`, `k`, ...), `n*` = name (`nf`, `nx`, `nb`, ...), `s*` = string, `c*` = comment (`c1`, ...), `o*` = operator, `m*` = number (`mi`, ...), `g*` = generic, `l*` = literal; `p` (punctuation) is an exact match, not a prefix.

Verified empirically by tokenizing real Go source through Chroma with `WithClasses(true)`: output includes `kd` (keyword-declaration, e.g. `func`), `nf` (name-function), `nx` (name-other, e.g. bare identifiers), `nb` (name-builtin, e.g. `println`), `p` (punctuation), `o` (operator, e.g. `:=`), `mi` (number-integer), `s` (string), `c1` (single-line comment), and `w` (whitespace — uncategorized, falls through to the base `.chroma` color).

This means one CSS attribute-prefix selector per category (`.chroma [class^="k"] { color: var(--chroma-keyword); }`, etc. — 8 rules total) covers every sub-class Chroma can emit, without enumerating individual class names or depending on Chroma's own bundled style palettes (`styles.Get("github")` etc., which only affect terminal/other formatters — irrelevant once you're emitting `WithClasses(true)` HTML). A base `.chroma { color: var(--chroma-fg); }` rule catches anything unmatched (e.g. `.w`), so nothing is ever left un-themed.

See `.context/adr/20260712-221004-md-rendering-theme.md` and `web/src/themes.css` for the full implementation.
