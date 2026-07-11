import { useRef, useState } from 'react'
import { useWorkspace } from '../useWorkspace'
import { Pane } from './Pane'
import { PaneDivider } from './PaneDivider'

// Reads useWorkspace(), renders 1-2 Panes (+ PaneDivider when 2 exist).
// Replaces the two <ContentPane/> route elements in App.tsx.
export function Workspace() {
  const { state } = useWorkspace()
  const containerRef = useRef<HTMLDivElement>(null)
  const [ratio, setRatio] = useState(0.5)

  const { panes, focusedPaneId } = state

  if (panes.length === 0) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <p className="text-neutral-500">Select a file from the sidebar.</p>
      </main>
    )
  }

  if (panes.length === 1) {
    return (
      <div ref={containerRef} className="flex flex-1 overflow-hidden">
        <Pane pane={panes[0]} focused={panes[0].id === focusedPaneId} canSplit={false} />
      </div>
    )
  }

  return (
    <div ref={containerRef} className="flex flex-1 overflow-hidden">
      <div className="flex min-w-0 overflow-hidden" style={{ flexBasis: `${ratio * 100}%` }}>
        <Pane pane={panes[0]} focused={panes[0].id === focusedPaneId} canSplit={true} />
      </div>
      <PaneDivider containerRef={containerRef} onRatioChange={setRatio} />
      <div
        className="flex min-w-0 overflow-hidden"
        style={{ flexBasis: `${(1 - ratio) * 100}%` }}
      >
        <Pane pane={panes[1]} focused={panes[1].id === focusedPaneId} canSplit={true} />
      </div>
    </div>
  )
}
