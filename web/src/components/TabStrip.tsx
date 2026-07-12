import type { Tab } from '../useWorkspace'

// Renders one pane's tab strip: label (path basename), click-to-focus,
// close (x), and a split-right icon per tab. Only reads { id, path } per
// tab — never touches FileContent, keeping the structure layer orthogonal
// to the content layer.
export function TabStrip({
  paneId,
  tabs,
  activeTabId,
  canSplit,
  onFocusTab,
  onCloseTab,
  onSplitRight,
}: {
  paneId: string
  tabs: Tab[]
  activeTabId: string
  canSplit: boolean
  onFocusTab: (tabId: string) => void
  onCloseTab: (tabId: string) => void
  onSplitRight: (tabId: string) => void
}) {
  return (
    <div
      data-pane-id={paneId}
      className="flex shrink-0 overflow-x-auto border-b border-[var(--border)]"
    >
      {tabs.map((tab) => {
        const isActive = tab.id === activeTabId
        const basename = tab.path.split('/').pop() ?? tab.path
        return (
          <div
            key={tab.id}
            data-tab-id={tab.id}
            data-active={isActive}
            onClick={() => onFocusTab(tab.id)}
            className={`group flex shrink-0 cursor-pointer items-center gap-1 border-r border-[var(--border)] px-3 py-1.5 text-sm ${
              isActive
                ? 'bg-[var(--bg)] text-[var(--text)]'
                : 'bg-[var(--bg-subtle)] text-[var(--text-muted)]'
            }`}
          >
            <span className="max-w-[12rem] truncate">{basename}</span>
            {!canSplit && (
              <button
                type="button"
                aria-label="Split right"
                data-action="split-right"
                onClick={(e) => {
                  e.stopPropagation()
                  onSplitRight(tab.id)
                }}
                className="rounded px-1 text-[var(--text-muted)] hover:bg-[var(--bg-subtle)] hover:text-[var(--text)]"
              >
                ⫽
              </button>
            )}
            <button
              type="button"
              aria-label="Close tab"
              data-action="close"
              onClick={(e) => {
                e.stopPropagation()
                onCloseTab(tab.id)
              }}
              className="rounded px-1 text-[var(--text-muted)] hover:bg-[var(--bg-subtle)] hover:text-[var(--text)]"
            >
              ×
            </button>
          </div>
        )
      })}
    </div>
  )
}
