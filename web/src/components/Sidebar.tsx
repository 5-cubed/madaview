import { useEffect, useState } from 'react'
import { VscLayoutSidebarLeft, VscLayoutSidebarLeftOff } from 'react-icons/vsc'
import { fetchTree } from '../api'
import { useWorkspace } from '../useWorkspace'
import type { TreeEntry } from '../types'

function getStoredFolded(): boolean {
  return localStorage.getItem('madaview:sidebar-folded') === 'true'
}

function setStoredFolded(folded: boolean): void {
  localStorage.setItem('madaview:sidebar-folded', String(folded))
}

export function Sidebar() {
  const [folded, setFolded] = useState(() => getStoredFolded())

  const handleToggle = () => {
    const next = !folded
    setFolded(next)
    setStoredFolded(next)
  }

  return (
    <nav
      className={`shrink-0 overflow-y-auto border-r border-[var(--border)] bg-[var(--bg)] text-[var(--text)] transition-[width] duration-200 ease-in-out ${folded ? 'w-10 p-1' : 'w-64 p-3'}`}
      data-testid="sidebar"
      data-folded={folded}
    >
      <button
        type="button"
        data-testid="sidebar-toggle"
        aria-label={folded ? 'Expand sidebar' : 'Collapse sidebar'}
        onClick={handleToggle}
        className="block w-full p-2 rounded hover:bg-[var(--bg-subtle)]"
      >
        {folded ? <VscLayoutSidebarLeftOff size={16} /> : <VscLayoutSidebarLeft size={16} />}
      </button>
      <div data-testid="sidebar-tree" className={folded ? 'hidden' : ''}>
        <TreeLevel path="" />
      </div>
    </nav>
  )
}

function TreeLevel({ path }: { path: string }) {
  const [entries, setEntries] = useState<TreeEntry[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    fetchTree(path)
      .then((result) => {
        if (!cancelled) setEntries(result)
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message)
      })
    return () => {
      cancelled = true
    }
  }, [path])

  if (error) {
    return <p className="px-2 text-sm text-red-500">{error}</p>
  }
  if (!entries) {
    return <p className="px-2 text-sm text-neutral-400">Loading…</p>
  }

  return (
    <ul className="space-y-0.5 text-sm">
      {entries.map((entry) => (
        <TreeNode key={entry.path} entry={entry} />
      ))}
    </ul>
  )
}

function TreeNode({ entry }: { entry: TreeEntry }) {
  const [expanded, setExpanded] = useState(false)
  const { openFile } = useWorkspace()

  if (!entry.isDir) {
    return (
      <li>
        <button
          type="button"
          onClick={() => openFile(entry.path)}
          className="block w-full rounded px-2 py-1 text-left hover:bg-[var(--bg-subtle)]"
        >
          {entry.name}
        </button>
      </li>
    )
  }

  return (
    <li>
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="block w-full rounded px-2 py-1 text-left text-[var(--text-muted)] hover:bg-[var(--bg-subtle)]"
      >
        {expanded ? '▾' : '▸'} {entry.name}/
      </button>
      {expanded && (
        <div className="ml-3 border-l border-[var(--border)] pl-2">
          <TreeLevel path={entry.path} />
        </div>
      )}
    </li>
  )
}
