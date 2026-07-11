export interface TreeEntry {
  name: string
  path: string
  isDir: boolean
}

export interface FileContent {
  path: string
  title: string
  mtime: string
  html: string
}

export interface ServerStatus {
  version: string
  currentRoot: string
  rootSource: 'cli' | 'ui' | 'default'
  goVersion: string
  os: string
  arch: string
}
