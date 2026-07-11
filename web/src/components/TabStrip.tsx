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
      className="flex shrink-0 overflow-x-auto border-b border-neutral-200 dark:border-neutral-800"
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
            className={`group flex shrink-0 cursor-pointer items-center gap-1 border-r border-neutral-200 px-3 py-1.5 text-sm dark:border-neutral-800 ${
              isActive
                ? 'bg-white dark:bg-neutral-900'
                : 'bg-neutral-50 text-neutral-500 dark:bg-neutral-950 dark:text-neutral-400'
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
                className="rounded px-1 text-neutral-400 hover:bg-neutral-200 hover:text-neutral-700 dark:hover:bg-neutral-800 dark:hover:text-neutral-200"
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
              className="rounded px-1 text-neutral-400 hover:bg-neutral-200 hover:text-neutral-700 dark:hover:bg-neutral-800 dark:hover:text-neutral-200"
            >
              ×
            </button>
          </div>
        )
      })}
    </div>
  )
}
