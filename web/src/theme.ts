// Theme application. Deep module, small interface: getStoredTheme/applyTheme
// are the only two functions anything outside this file ever calls. Themes
// themselves are pure CSS (see themes.css) — this module only decides which
// `data-theme` id is active and persists that choice.
//
// Switching themes is exactly one DOM mutation
// (`document.documentElement.dataset.theme = id`). The browser's own
// cascade repaints every themed surface with zero React re-render, so
// switching never disturbs the tabs-and-split-view "kept mounted" contract
// (see .context/adr/20260712-001324-tabs-and-split-view.md).

export interface ThemeMeta {
  id: string
  label: string
}

export const THEMES: ThemeMeta[] = [
  { id: 'github-light', label: 'GitHub Light' },
  { id: 'github-dark', label: 'GitHub Dark' },
  { id: 'kanagawa', label: 'Kanagawa' },
  { id: 'gruvbox', label: 'Gruvbox' },
  { id: 'solarized-light', label: 'Solarized Light' },
  { id: 'one-dark', label: 'One Dark' },
]

export const DEFAULT_THEME = 'github-light'

const STORAGE_KEY = 'madaview:theme'

// Reads the persisted theme choice. Falls back to DEFAULT_THEME for a
// first-ever visit (no localStorage entry) or a stored id that no longer
// matches a known theme (e.g. an older build's theme was removed).
export function getStoredTheme(): string {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored && THEMES.some((t) => t.id === stored)) {
    return stored
  }
  return DEFAULT_THEME
}

// Applies a theme to the whole document and persists the choice. This is
// the only place that ever mutates `document.documentElement.dataset.theme`
// for the *committed* theme (the Settings picker's hover-preview mutates a
// nested wrapper's own `data-theme` instead — see Settings.tsx).
export function applyTheme(id: string): void {
  document.documentElement.dataset.theme = id
  localStorage.setItem(STORAGE_KEY, id)
}
