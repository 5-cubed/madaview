import {
  createElement,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  type ReactNode,
} from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

// See .context/adr/20260712-001324-tabs-and-split-view.md for the full
// design rationale. This file is the single state-owning layer for the
// tabs/split-view workspace: types, a pure reducer, and the
// WorkspaceProvider/useWorkspace() that wire it to the URL.

export interface Tab {
  id: string
  path: string
}

export interface Pane {
  id: string
  tabs: Tab[]
  activeTabId: string
}

export interface WorkspaceState {
  panes: Pane[]
  focusedPaneId: string | null
}

export type WorkspaceAction =
  | { type: 'OPEN_FILE'; path: string }
  | { type: 'FOCUS_TAB'; paneId: string; tabId: string }
  | { type: 'FOCUS_PANE'; paneId: string }
  | { type: 'CLOSE_TAB'; paneId: string; tabId: string }
  | { type: 'SPLIT_RIGHT'; fromPaneId: string; tabId: string }
  | { type: 'RESET_FROM_URL'; path: string }

const VIEW_PREFIX = '/view/'

// Strips the leading "/view/" from a pathname, returning '' for '/' or any
// pathname that isn't under /view.
function pathFromPathname(pathname: string): string {
  if (pathname.startsWith(VIEW_PREFIX)) {
    return pathname.slice(VIEW_PREFIX.length)
  }
  return ''
}

let idCounter = 0
function nextId(prefix: string): string {
  idCounter += 1
  return `${prefix}-${idCounter}`
}

function emptyState(): WorkspaceState {
  return { panes: [], focusedPaneId: null }
}

function stateFromPath(path: string): WorkspaceState {
  if (!path) return emptyState()
  const tab: Tab = { id: nextId('tab'), path }
  const pane: Pane = { id: nextId('pane'), tabs: [tab], activeTabId: tab.id }
  return { panes: [pane], focusedPaneId: pane.id }
}

// Dev-only invariant assertions. Thrown errors here are intentional loud
// failures during development; they are compiled out in production builds
// since import.meta.env.DEV is statically known to Vite.
function assertInvariants(state: WorkspaceState) {
  if (!import.meta.env.DEV) return

  if (state.panes.length > 2) {
    throw new Error(`workspace invariant violated: ${state.panes.length} panes (max 2)`)
  }

  for (const pane of state.panes) {
    if (pane.tabs.length === 0) {
      throw new Error(`workspace invariant violated: pane ${pane.id} has zero tabs`)
    }
    if (!pane.tabs.some((t) => t.id === pane.activeTabId)) {
      throw new Error(
        `workspace invariant violated: pane ${pane.id} activeTabId ${pane.activeTabId} not among its tabs`,
      )
    }
  }

  if (state.focusedPaneId !== null && !state.panes.some((p) => p.id === state.focusedPaneId)) {
    throw new Error(
      `workspace invariant violated: focusedPaneId ${state.focusedPaneId} does not match any pane`,
    )
  }

  if (state.panes.length > 0 && state.focusedPaneId === null) {
    throw new Error('workspace invariant violated: panes exist but focusedPaneId is null')
  }
}

export function workspaceReducer(state: WorkspaceState, action: WorkspaceAction): WorkspaceState {
  const next = workspaceReducerInner(state, action)
  assertInvariants(next)
  return next
}

function workspaceReducerInner(state: WorkspaceState, action: WorkspaceAction): WorkspaceState {
  switch (action.type) {
    case 'OPEN_FILE': {
      const { path } = action
      // No focused pane yet: create the first one.
      if (state.focusedPaneId === null) {
        const tab: Tab = { id: nextId('tab'), path }
        const pane: Pane = { id: nextId('pane'), tabs: [tab], activeTabId: tab.id }
        return { panes: [pane], focusedPaneId: pane.id }
      }

      const paneIndex = state.panes.findIndex((p) => p.id === state.focusedPaneId)
      if (paneIndex === -1) return state
      const pane = state.panes[paneIndex]

      const existing = pane.tabs.find((t) => t.path === path)
      if (existing) {
        if (pane.activeTabId === existing.id) return state
        const updatedPane: Pane = { ...pane, activeTabId: existing.id }
        const panes = state.panes.slice()
        panes[paneIndex] = updatedPane
        return { ...state, panes }
      }

      const tab: Tab = { id: nextId('tab'), path }
      const updatedPane: Pane = { ...pane, tabs: [...pane.tabs, tab], activeTabId: tab.id }
      const panes = state.panes.slice()
      panes[paneIndex] = updatedPane
      return { ...state, panes }
    }

    case 'FOCUS_TAB': {
      const { paneId, tabId } = action
      const paneIndex = state.panes.findIndex((p) => p.id === paneId)
      if (paneIndex === -1) return state
      const pane = state.panes[paneIndex]
      if (!pane.tabs.some((t) => t.id === tabId)) return state
      if (pane.activeTabId === tabId && state.focusedPaneId === paneId) return state
      const updatedPane: Pane = { ...pane, activeTabId: tabId }
      const panes = state.panes.slice()
      panes[paneIndex] = updatedPane
      return { ...state, panes, focusedPaneId: paneId }
    }

    case 'FOCUS_PANE': {
      const { paneId } = action
      if (!state.panes.some((p) => p.id === paneId)) return state
      if (state.focusedPaneId === paneId) return state
      return { ...state, focusedPaneId: paneId }
    }

    case 'CLOSE_TAB': {
      const { paneId, tabId } = action
      const paneIndex = state.panes.findIndex((p) => p.id === paneId)
      if (paneIndex === -1) return state
      const pane = state.panes[paneIndex]
      const tabIndex = pane.tabs.findIndex((t) => t.id === tabId)
      if (tabIndex === -1) return state

      const remainingTabs = pane.tabs.filter((t) => t.id !== tabId)

      if (remainingTabs.length === 0) {
        const panes = state.panes.filter((p) => p.id !== paneId)
        let focusedPaneId = state.focusedPaneId
        if (focusedPaneId === paneId) {
          focusedPaneId = panes.length > 0 ? panes[0].id : null
        }
        return { panes, focusedPaneId }
      }

      let activeTabId = pane.activeTabId
      if (activeTabId === tabId) {
        // Activate the left neighbor, or the right neighbor if it was first.
        const neighborIndex = tabIndex > 0 ? tabIndex - 1 : 0
        activeTabId = remainingTabs[Math.min(neighborIndex, remainingTabs.length - 1)].id
      }

      const updatedPane: Pane = { ...pane, tabs: remainingTabs, activeTabId }
      const panes = state.panes.slice()
      panes[paneIndex] = updatedPane
      return { ...state, panes }
    }

    case 'SPLIT_RIGHT': {
      const { fromPaneId, tabId } = action
      if (state.panes.length >= 2) return state
      const fromPane = state.panes.find((p) => p.id === fromPaneId)
      if (!fromPane) return state
      const sourceTab = fromPane.tabs.find((t) => t.id === tabId)
      if (!sourceTab) return state

      const newTab: Tab = { id: nextId('tab'), path: sourceTab.path }
      const newPane: Pane = { id: nextId('pane'), tabs: [newTab], activeTabId: newTab.id }
      return { panes: [...state.panes, newPane], focusedPaneId: newPane.id }
    }

    case 'RESET_FROM_URL': {
      return stateFromPath(action.path)
    }

    default:
      return state
  }
}

interface WorkspaceContextValue {
  state: WorkspaceState
  openFile: (path: string) => void
  focusTab: (paneId: string, tabId: string) => void
  focusPane: (paneId: string) => void
  closeTab: (paneId: string, tabId: string) => void
  splitRight: (fromPaneId: string, tabId: string) => void
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null)

// Finds the focused pane's active tab path, or '' if there is no focused
// pane (empty workspace).
function focusedActivePath(state: WorkspaceState): string {
  const pane = state.panes.find((p) => p.id === state.focusedPaneId)
  if (!pane) return ''
  const tab = pane.tabs.find((t) => t.id === pane.activeTabId)
  return tab?.path ?? ''
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const location = useLocation()
  const [state, dispatch] = useReducer(
    workspaceReducer,
    undefined,
    () => stateFromPath(pathFromPathname(window.location.pathname)),
  )

  // popstate: every browser back/forward is treated exactly like a fresh
  // page load — reset to a single pane/single tab matching the landed-on
  // URL, discarding any prior multi-tab/split layout.
  useEffect(() => {
    function handlePopState() {
      dispatch({ type: 'RESET_FROM_URL', path: pathFromPathname(window.location.pathname) })
    }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  // Keep the URL in sync with the focused pane's active tab. Fires
  // pushState on every change, literally, with no dedup/replaceState —
  // an intentional accepted growth of the history stack per the ADR.
  const lastPushedPath = useRef<string | null>(null)
  useEffect(() => {
    const path = focusedActivePath(state)
    if (lastPushedPath.current === path) return
    lastPushedPath.current = path
    const target = path ? `${VIEW_PREFIX}${path}` : '/'
    if (window.location.pathname === target) return
    window.history.pushState(null, '', target)
  }, [state])

  const openFile = useMemo(
    () => (path: string) => {
      dispatch({ type: 'OPEN_FILE', path })
      if (!location.pathname.startsWith(VIEW_PREFIX)) {
        navigate(`${VIEW_PREFIX}${path}`)
      }
    },
    [location.pathname, navigate],
  )

  const focusTab = useMemo(
    () => (paneId: string, tabId: string) => dispatch({ type: 'FOCUS_TAB', paneId, tabId }),
    [],
  )

  const focusPane = useMemo(() => (paneId: string) => dispatch({ type: 'FOCUS_PANE', paneId }), [])

  const closeTab = useMemo(
    () => (paneId: string, tabId: string) => dispatch({ type: 'CLOSE_TAB', paneId, tabId }),
    [],
  )

  const splitRight = useMemo(
    () => (fromPaneId: string, tabId: string) =>
      dispatch({ type: 'SPLIT_RIGHT', fromPaneId, tabId }),
    [],
  )

  const value = useMemo(
    () => ({ state, openFile, focusTab, focusPane, closeTab, splitRight }),
    [state, openFile, focusTab, focusPane, closeTab, splitRight],
  )

  // Plain createElement (not JSX) so this file can stay a .ts module per
  // the ADR's file layout — this is the only element it needs to produce.
  return createElement(WorkspaceContext.Provider, { value }, children)
}

export function useWorkspace(): WorkspaceContextValue {
  const ctx = useContext(WorkspaceContext)
  if (!ctx) {
    throw new Error('useWorkspace must be used within a WorkspaceProvider')
  }
  return ctx
}
