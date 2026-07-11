import { useEffect, useState, type SubmitEvent } from 'react'
import { fetchStatus, setRoot } from '../api'
import type { ServerStatus } from '../types'

export function Settings() {
  const [status, setStatus] = useState<ServerStatus | null>(null)
  const [rootInput, setRootInput] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const loadStatus = () => {
    fetchStatus()
      .then((s) => {
        setStatus(s)
        setRootInput(s.currentRoot)
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(loadStatus, [])

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const s = await setRoot(rootInput)
      setStatus(s)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <main className="flex-1 overflow-y-auto p-6">
      <h1 className="mb-4 text-lg font-medium">Settings</h1>

      <form onSubmit={handleSubmit} className="mb-6 flex max-w-xl gap-2">
        <input
          type="text"
          value={rootInput}
          onChange={(e) => setRootInput(e.target.value)}
          className="flex-1 rounded border border-neutral-300 px-2 py-1 text-sm dark:border-neutral-700 dark:bg-neutral-900"
          placeholder="Root folder path"
        />
        <button
          type="submit"
          disabled={saving}
          className="rounded bg-neutral-900 px-3 py-1 text-sm text-white disabled:opacity-50 dark:bg-neutral-100 dark:text-neutral-900"
        >
          {saving ? 'Saving…' : 'Set root'}
        </button>
      </form>

      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      {status && (
        <dl className="grid max-w-xl grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
          <dt className="text-neutral-500">Version</dt>
          <dd>{status.version}</dd>
          <dt className="text-neutral-500">Current root</dt>
          <dd className="font-mono">{status.currentRoot}</dd>
          <dt className="text-neutral-500">Root source</dt>
          <dd>{status.rootSource}</dd>
          <dt className="text-neutral-500">Go version</dt>
          <dd>{status.goVersion}</dd>
          <dt className="text-neutral-500">OS / Arch</dt>
          <dd>
            {status.os} / {status.arch}
          </dd>
        </dl>
      )}
    </main>
  )
}
