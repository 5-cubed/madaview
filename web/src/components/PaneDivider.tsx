import { useCallback, useEffect, useRef } from 'react'

// Draggable divider between two panes. The ratio itself lives in the
// parent Workspace's local useState (pure presentational session state,
// not a structural invariant) — this component just reports pointer
// movement as a ratio delta against the container it's dragged within.
export function PaneDivider({
  containerRef,
  onRatioChange,
}: {
  containerRef: React.RefObject<HTMLDivElement | null>
  onRatioChange: (ratio: number) => void
}) {
  const draggingRef = useRef(false)

  const handleMove = useCallback(
    (e: MouseEvent) => {
      if (!draggingRef.current || !containerRef.current) return
      const rect = containerRef.current.getBoundingClientRect()
      const ratio = (e.clientX - rect.left) / rect.width
      const clamped = Math.min(0.85, Math.max(0.15, ratio))
      onRatioChange(clamped)
    },
    [containerRef, onRatioChange],
  )

  const handleUp = useCallback(() => {
    draggingRef.current = false
  }, [])

  useEffect(() => {
    window.addEventListener('mousemove', handleMove)
    window.addEventListener('mouseup', handleUp)
    return () => {
      window.removeEventListener('mousemove', handleMove)
      window.removeEventListener('mouseup', handleUp)
    }
  }, [handleMove, handleUp])

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      data-testid="pane-divider"
      onMouseDown={() => {
        draggingRef.current = true
      }}
      className="w-1 shrink-0 cursor-col-resize bg-[var(--border)] hover:bg-[var(--accent)]"
    />
  )
}
