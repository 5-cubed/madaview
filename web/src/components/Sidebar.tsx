import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { fetchTree } from '../api'
import type { TreeEntry } from '../types'

export function Sidebar() {
  return (
    <nav className="w-64 shrink-0 overflow-y-auto border-r border-neutral-200 p-3 dark:border-neutral-800">
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

  if (!entry.isDir) {
    return (
      <li>
        <Link
          to={`/view/${entry.path}`}
          className="block rounded px-2 py-1 hover:bg-neutral-100 dark:hover:bg-neutral-800"
        >
          {entry.name}
        </Link>
      </li>
    )
  }

  return (
    <li>
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="block w-full rounded px-2 py-1 text-left text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800"
      >
        {expanded ? '▾' : '▸'} {entry.name}/
      </button>
      {expanded && (
        <div className="ml-3 border-l border-neutral-200 pl-2 dark:border-neutral-800">
          <TreeLevel path={entry.path} />
        </div>
      )}
    </li>
  )
}
