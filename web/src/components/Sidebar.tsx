import { useEffect, useRef, useState } from 'react'
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

function clampWidth(width: number): number {
  // TODO:
  // 1. Clamp width between a minimum of 180 and a maximum of 600 (never
  //    below 180, never above 600).
  // 2. Use Math.min and Math.max together, the same pattern
  //    PaneDivider.tsx uses for its ratio clamp.
  // e.g. clampWidth(700) -> 600   (over max, clamps down)
  //      clampWidth(100) -> 180   (under min, clamps up)
  //      clampWidth(350) -> 350   (in range, unchanged)
  return Math.min(600, Math.max(180, width))
}

function getStoredWidth(): number {
  // TODO:
  // 1. Read the raw string via localStorage.getItem('madaview:sidebar-width').
  // 2. Convert it to a number with Number(raw).
  // 3. If that number is NaN (missing key, or a non-numeric string),
  //    return the default, 256.
  // 4. Otherwise return clampWidth(that number) — a valid-but-out-of-range
  //    stored value (e.g. 900) must clamp into [180, 600], not fall back
  //    to the default.
  // e.g. raw="180" -> Number("180") -> 180 -> not NaN -> clampWidth(180) -> 180
  //      raw="900" -> Number("900") -> 900 -> not NaN -> clampWidth(900) -> 600
  //      raw=null  -> Number(null) -> 0 -> NOT NaN(!) -> handle missing/null
  //      explicitly so it returns 256, not clampWidth(0)
  const raw = localStorage.getItem('madaview:sidebar-width')
  const number = raw === null ? NaN : Number(raw)
  if (!Number.isNaN(number)) {
      return clampWidth(number)
  } else {
      return 256
  }
}

function setStoredWidth(width: number): void {
  localStorage.setItem('madaview:sidebar-width', String(width))
}

export function Sidebar() {
  const [folded, setFolded] = useState(() => getStoredFolded())
  const [width, setWidth] = useState(() => getStoredWidth())
  const [isDragging, setIsDragging] = useState(false)
  const navRef = useRef<HTMLElement>(null)

  const handleToggle = () => {
    const next = !folded
    setFolded(next)
    setStoredFolded(next)
  }

  const handleResizeStart = () => {
    setIsDragging(true)
  }

  useEffect(() => {
    if (!isDragging) return

    const handleMouseMove = (e: MouseEvent) => {
      if (!navRef.current) return
      // TODO:
      // 1. Read navRef.current's bounding rect via getBoundingClientRect().
      // 2. Compute the raw width as e.clientX minus that rect's left edge.
      // 3. Clamp the raw width with clampWidth(...) (defined above) so it
      //    never renders below 180 or above 600, even mid-drag.
      // 4. Call setWidth with the clamped value.
      // e.g. navRect.left=0, e.clientX=350 -> raw width = 350
      //      -> clampWidth(350) -> 350 (in range) -> setWidth(350)
      //      e.clientX=900 -> raw width = 900 -> clampWidth(900) -> 600
      //      -> setWidth(600)
      const navRect = navRef.current?.getBoundingClientRect()
      const currWidth = clampWidth(e.clientX - navRect.left)
      setWidth(currWidth)
    }

    const handleMouseUp = () => {
      // TODO:
      // 1. Call setStoredWidth(width) using the current width state, to
      //    persist exactly once per drag.
      // 2. Then call setIsDragging(false) to end the drag and restore the
      //    transition class.
      // e.g. width=180 at mouseup -> setStoredWidth(180) -> localStorage
      //      now holds "180" -> setIsDragging(false)
      setStoredWidth(width)
      setIsDragging(false)
    }

    window.addEventListener('mousemove', handleMouseMove)
    window.addEventListener('mouseup', handleMouseUp)
    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, width])

  return (
    <>
      <nav
        ref={navRef}
        className={`shrink-0 overflow-y-auto border-r border-[var(--border)] bg-[var(--bg)] text-[var(--text)] ${isDragging ? '' : 'transition-[width] duration-200 ease-in-out'} ${folded ? 'w-10 p-1' : 'p-3'}`}
        style={folded ? undefined : { width }}
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
      {!folded && (
        <div
          data-testid="sidebar-resize-handle"
          role="separator"
          aria-orientation="vertical"
          onMouseDown={handleResizeStart}
          className="w-1 shrink-0 cursor-col-resize bg-[var(--border)] hover:bg-[var(--accent)]"
        />
      )}
    </>
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
