import { useState } from 'react'
import { applyTheme, THEMES } from '../theme'

// A themed sample used as the preview swatch: hover/focus on a theme
// button recolors this subtree only (its own `data-theme` attribute
// overrides the outer, committed one, since CSS custom properties cascade
// into nested data-theme scopes) — the rest of the app stays on the
// committed theme until the user actually clicks.
function PreviewSwatch({ themeId }: { themeId: string }) {
  return (
    <div
      data-theme={themeId}
      data-testid="theme-preview-swatch"
      className="rounded border border-[var(--border)] bg-[var(--bg)] p-3"
    >
      <div className="prose prose-sm max-w-none">
        <h1 className="!my-0 !text-base">Heading one</h1>
        <h2 className="!my-1 !text-sm">Heading two</h2>
        <h3 className="!my-1 !text-sm">Heading three</h3>
        <pre className="chroma !my-2 !p-2 !text-xs">
          <code>
            <span className="kd">func</span> <span className="nf">main</span>
            <span className="p">()</span> <span className="p">{'{'}</span>
            {'\n  '}
            <span className="c1">// comment</span>
            {'\n  '}
            <span className="nx">x</span> <span className="o">:=</span>{' '}
            <span className="s">&quot;hi&quot;</span>
            {'\n'}
            <span className="p">{'}'}</span>
          </code>
        </pre>
        <a href="#">a sample link</a>
      </div>
    </div>
  )
}

// Button list (not a native <select>) so hover/focus can drive a live
// preview: <option> elements don't reliably fire hover/focus events across
// browsers, and the direction requires the preview to recolor as the user
// browses, before committing. Click both applies the theme immediately
// (document.documentElement + localStorage, via applyTheme) and updates
// local state for the "currently selected" highlight — no separate
// apply/save step.
export function ThemePicker({ committedTheme }: { committedTheme: string }) {
  const [selected, setSelected] = useState(committedTheme)
  const [previewTheme, setPreviewTheme] = useState(committedTheme)

  const handleCommit = (id: string) => {
    applyTheme(id)
    setSelected(id)
    setPreviewTheme(id)
  }

  return (
    <section className="mb-6 max-w-xl">
      <h2 className="mb-2 text-sm font-medium text-[var(--text-muted)]">Theme</h2>
      <div className="flex gap-4">
        <ul className="flex flex-1 flex-col gap-1">
          {THEMES.map((theme) => (
            <li key={theme.id}>
              <button
                type="button"
                data-theme-id={theme.id}
                data-selected={theme.id === selected}
                onMouseEnter={() => setPreviewTheme(theme.id)}
                onFocus={() => setPreviewTheme(theme.id)}
                onMouseLeave={() => setPreviewTheme(selected)}
                onBlur={() => setPreviewTheme(selected)}
                onClick={() => handleCommit(theme.id)}
                className={`block w-full rounded px-2 py-1 text-left text-sm ${
                  theme.id === selected
                    ? 'bg-[var(--accent)] text-white'
                    : 'text-[var(--text)] hover:bg-[var(--bg-subtle)]'
                }`}
              >
                {theme.label}
              </button>
            </li>
          ))}
        </ul>
        <div className="flex-1">
          <PreviewSwatch themeId={previewTheme} />
        </div>
      </div>
    </section>
  )
}
