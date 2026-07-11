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
        <div className="flex h-full flex-col">
          <header className="flex items-center justify-between border-b border-neutral-200 px-4 py-2 dark:border-neutral-800">
            <span className="font-medium">madaview</span>
            <Link
              to="/settings"
              className="text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100"
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
