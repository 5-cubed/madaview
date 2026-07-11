import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { fetchFile } from '../api'
import { hydrate } from '../hydrate'
import type { FileContent } from '../types'

export function ContentPane() {
  const params = useParams<{ '*': string }>()
  const path = params['*'] ?? ''
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

  if (!path) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <p className="text-neutral-500">Select a file from the sidebar.</p>
      </main>
    )
  }

  if (error) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <p className="text-red-500">{error}</p>
      </main>
    )
  }

  if (!file) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <p className="text-neutral-400">Loading…</p>
      </main>
    )
  }

  return (
    <main className="flex-1 overflow-y-auto p-6">
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
