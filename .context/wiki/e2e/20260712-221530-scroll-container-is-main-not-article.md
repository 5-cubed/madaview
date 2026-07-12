# The scrollable element in ContentPane is `<main>`, not `<article>`

`web/src/components/ContentPane.tsx` renders `<main className="... overflow-y-auto ...">` wrapping an `<article>` that holds the sanitized markdown HTML (`dangerouslySetInnerHTML`). The `overflow-y-auto`/scroll behavior lives on `<main>` — `<article>` itself has no overflow styling and its `scrollTop` is always 0 regardless of actual scroll position.

Any e2e scenario that needs to read or set scroll position must target the visible `<main>` element (there are 1-2 `<main>` elements in the DOM at once when panes are split — filter by `getComputedStyle(el).display !== 'none'`, matching `e2e/tabs-and-split-view/run`'s `scrollTopOfVisibleMain` helper), not `document.querySelector('main article')` or similar. Reading/writing `.scrollTop` on `<article>` silently no-ops instead of erroring, which makes this mistake easy to miss until a "scroll preserved" assertion mysteriously always reports `scrollTop: 0`.
