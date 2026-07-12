import { BrowserRouter, Link, Route, Routes } from 'react-router-dom'
import { Sidebar } from './components/Sidebar'
import { Workspace } from './components/Workspace'
import { Settings } from './pages/Settings'
import { WorkspaceProvider } from './useWorkspace'

function App() {
  return (
    <BrowserRouter>
      {/* WorkspaceProvider wraps the whole app shell (header, Sidebar,
          Routes) — not just the /view routes — so tab/pane state survives
          a /settings round-trip. Only an actual page reload resets it. */}
      <WorkspaceProvider>
        <div className="flex h-full flex-col bg-[var(--bg)] text-[var(--text)]">
          <header className="flex items-center justify-between border-b border-[var(--border)] bg-[var(--bg)] px-4 py-2 text-[var(--text)]">
            <span className="font-medium">madaview</span>
            <Link
              to="/settings"
              className="text-sm text-[var(--text-muted)] hover:text-[var(--text)]"
            >
              Settings
            </Link>
          </header>
          <div className="flex flex-1 overflow-hidden">
            <Sidebar />
            <Routes>
              <Route path="/view/*" element={<Workspace />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/" element={<Workspace />} />
            </Routes>
          </div>
        </div>
      </WorkspaceProvider>
    </BrowserRouter>
  )
}

export default App
