# A nested `data-theme` attribute overrides the outer one for just that subtree

`web/src/theme.ts`'s theme tokens are defined as `[data-theme="x"] { --var: ... }` CSS custom-property blocks (see `web/src/themes.css`). Because CSS custom properties inherit down the DOM tree and a more-specific/nested selector match wins, giving a *descendant* element its own `data-theme` attribute re-scopes every `var(--...)` reference inside that subtree to the nested theme — without touching the outer, committed `document.documentElement.dataset.theme`.

This is what makes a live, no-commit hover-preview possible with pure CSS: `web/src/components/ThemePicker.tsx`'s `PreviewSwatch` sets its own `data-theme` (updated on `onMouseEnter`/`onFocus`, reset to the committed theme on `onMouseLeave`/`onBlur`) while the rest of the app stays on whatever theme is actually applied to `<html>`. No React re-render of the wider app is needed to preview a theme — only the swatch's own attribute changes.
