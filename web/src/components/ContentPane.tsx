import { useEffect, useRef, useState } from 'react'
import { fetchFile } from '../api'
import { hydrate } from '../hydrate'
import type { FileContent } from '../types'

// Prop-driven (not route-driven via useParams) so a ContentPane instance
// can be mounted once per open tab and stay alive for the tab's lifetime.
// `visible` only toggles a `display` style on this component's own root —
// it is never conditionally unmounted while its tab exists. Browsers
// natively preserve scrollTop on a display:none element that stays in the
// DOM, which is what makes "instant switch, no re-fetch, no scroll reset"
// fall out of the existing fetch-cache logic for free.
export function ContentPane({ path, visible }: { path: string; visible: boolean }) {
  const [file, setFile] = useState<FileContent | null>(null)
  const [error, setError] = useState<string | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!path) {
      setFile(null)
      return
    }
    let cancelled = false
    setError(null)
    fetchFile(path)
      .then((result) => {
        if (!cancelled) setFile(result)
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message)
      })
    return () => {
      cancelled = true
    }
  }, [path])

  useEffect(() => {
    if (file && containerRef.current) {
      hydrate(containerRef.current)
    }
  }, [file])

  const rootStyle = { display: visible ? undefined : 'none' }

  if (!path) {
    return (
      <main className="flex-1 overflow-y-auto p-6" style={rootStyle}>
        <p className="text-neutral-500">Select a file from the sidebar.</p>
      </main>
    )
  }

  if (error) {
    return (
      <main className="flex-1 overflow-y-auto p-6" style={rootStyle}>
        <p className="text-red-500">{error}</p>
      </main>
    )
  }

  if (!file) {
    return (
      <main className="flex-1 overflow-y-auto p-6" style={rootStyle}>
        <p className="text-neutral-400">Loading…</p>
      </main>
    )
  }

  return (
    <main className="flex-1 overflow-y-auto p-6" style={rootStyle}>
      <article
        ref={containerRef}
        className="prose max-w-none dark:prose-invert"
        // The HTML here is sanitized server-side (bluemonday) before it
        // ever reaches the client — see internal/markdown/sanitize.go.
        dangerouslySetInnerHTML={{ __html: file.html }}
      />
    </main>
  )
}
