import type { Pane as PaneModel } from '../useWorkspace'
import { useWorkspace } from '../useWorkspace'
import { TabStrip } from './TabStrip'
import { ContentPane } from './ContentPane'

// Composes TabStrip + all of this pane's ContentPane instances
// (visibility-toggled, never conditionally unmounted). Applies a
// focus-highlight border when it's the focused pane. Any click inside a
// pane calls focusPane/focusTab so sidebar clicks know where to target.
export function Pane({
  pane,
  focused,
  canSplit,
}: {
  pane: PaneModel
  focused: boolean
  canSplit: boolean
}) {
  const { focusTab, focusPane, closeTab, splitRight } = useWorkspace()

  return (
    <div
      data-pane-root={pane.id}
      data-focused={focused}
      onMouseDown={() => focusPane(pane.id)}
      className={`flex h-full min-w-0 flex-1 flex-col border-2 ${
        focused
          ? 'border-blue-400 dark:border-blue-600'
          : 'border-transparent'
      }`}
    >
      <TabStrip
        paneId={pane.id}
        tabs={pane.tabs}
        activeTabId={pane.activeTabId}
        canSplit={canSplit}
        onFocusTab={(tabId) => focusTab(pane.id, tabId)}
        onCloseTab={(tabId) => closeTab(pane.id, tabId)}
        onSplitRight={(tabId) => splitRight(pane.id, tabId)}
      />
      <div className="relative flex flex-1 overflow-hidden">
        {pane.tabs.map((tab) => (
          <ContentPane key={tab.id} path={tab.path} visible={tab.id === pane.activeTabId} />
        ))}
      </div>
    </div>
  )
}
