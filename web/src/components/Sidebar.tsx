import { useEffect, useState } from 'react'
import { fetchTree } from '../api'
import { useWorkspace } from '../useWorkspace'
import type { TreeEntry } from '../types'

export function Sidebar() {
  return (
    <nav className="w-64 shrink-0 overflow-y-auto border-r border-[var(--border)] bg-[var(--bg)] p-3 text-[var(--text)]">
      <TreeLevel path="" />
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
