import { useState } from 'react'

export interface ThreePaneLayoutProps {
  /** Left pane content (list/filters) */
  left: React.ReactNode
  /** Center pane content (main view) */
  center: React.ReactNode
  /** Right pane content (inspector) */
  right?: React.ReactNode
  /** Left pane width in pixels */
  leftWidth?: number
  /** Right pane width in pixels */
  rightWidth?: number
  /** Whether to show the right pane */
  showRight?: boolean
  /** Called when right pane visibility toggles */
  onToggleRight?: () => void
  /** Custom class name */
  className?: string
}

export function ThreePaneLayout({
  left,
  center,
  right,
  leftWidth = 280,
  rightWidth = 320,
  showRight = true,
  onToggleRight,
  className = '',
}: ThreePaneLayoutProps) {
  return (
    <div className={`three-pane-layout d-flex h-100 ${className}`}>
      {/* Left Pane */}
      <div
        className="pane-left border-end bg-body overflow-auto"
        style={{ width: leftWidth, flexShrink: 0 }}
      >
        {left}
      </div>

      {/* Center Pane */}
      <div className="pane-center flex-grow-1 overflow-auto">
        {center}
      </div>

      {/* Right Pane */}
      {showRight && right && (
        <div
          className="pane-right border-start bg-body overflow-auto"
          style={{ width: rightWidth, flexShrink: 0 }}
        >
          {right}
        </div>
      )}

      {/* Toggle button for right pane */}
      {onToggleRight && (
        <button
          type="button"
          className="btn btn-sm btn-link position-absolute"
          style={{ right: showRight ? rightWidth + 4 : 4, top: '50%', transform: 'translateY(-50%)' }}
          onClick={onToggleRight}
          title={showRight ? 'Hide inspector' : 'Show inspector'}
        >
          {showRight ? '›' : '‹'}
        </button>
      )}
    </div>
  )
}

// Wrapper that manages right pane toggle state
export function ThreePaneLayoutWithToggle({
  left,
  center,
  right,
  leftWidth = 280,
  rightWidth = 320,
  defaultShowRight = true,
  className = '',
}: Omit<ThreePaneLayoutProps, 'showRight' | 'onToggleRight'> & { defaultShowRight?: boolean }) {
  const [showRight, setShowRight] = useState(defaultShowRight)

  return (
    <ThreePaneLayout
      left={left}
      center={center}
      right={right}
      leftWidth={leftWidth}
      rightWidth={rightWidth}
      showRight={showRight}
      onToggleRight={() => setShowRight(!showRight)}
      className={className}
    />
  )
}
