import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import 'katex/dist/katex.min.css'
import './index.css'
import './themes.css'
import { applyTheme, getStoredTheme } from './theme'
import App from './App.tsx'

// Set as early as possible, before the first paint of React's own DOM —
// the earliest point available without a blocking inline <script> in
// index.html (a full FOUC fix is out of scope, see the ADR's Ambiguous
// Zone).
applyTheme(getStoredTheme())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
