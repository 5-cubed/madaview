import type { FileContent, ServerStatus, TreeEntry } from './types'

async function json<T>(resPromise: Promise<Response>): Promise<T> {
  const res = await resPromise
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export function fetchTree(path: string): Promise<TreeEntry[]> {
  return json(fetch(`/api/tree?path=${encodeURIComponent(path)}`))
}

export function fetchFile(path: string): Promise<FileContent> {
  return json(fetch(`/api/file?path=${encodeURIComponent(path)}`))
}

export function fetchStatus(): Promise<ServerStatus> {
  return json(fetch('/api/status'))
}

export function setRoot(root: string): Promise<ServerStatus> {
  return json(
    fetch('/api/root', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ root }),
    }),
  )
}
