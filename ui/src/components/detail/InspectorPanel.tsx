import { useState } from 'react'

export interface InspectorPanelProps {
  /** Panel title */
  title: string
  /** Panel subtitle */
  subtitle?: string
  /** Action buttons to show in header */
  actions?: React.ReactNode
  /** Panel content */
  children: React.ReactNode
  /** Called when close button is clicked */
  onClose?: () => void
  /** Loading state */
  loading?: boolean
  /** Custom class name */
  className?: string
}

function SkeletonContent() {
  return (
    <div className="placeholder-glow">
      <div className="mb-3">
        <span className="placeholder col-6 mb-2 d-block"></span>
        <span className="placeholder col-8"></span>
      </div>
      <div className="mb-3">
        <span className="placeholder col-4 mb-2 d-block"></span>
        <span className="placeholder col-10"></span>
      </div>
      <div className="mb-3">
        <span className="placeholder col-5 mb-2 d-block"></span>
        <span className="placeholder col-7"></span>
      </div>
    </div>
  )
}

export function InspectorPanel({
  title,
  subtitle,
  actions,
  children,
  onClose,
  loading = false,
  className = '',
}: InspectorPanelProps) {
  return (
    <div className={`inspector-panel d-flex flex-column h-100 ${className}`}>
      {/* Header */}
      <div className="inspector-header p-3 border-bottom bg-body-tertiary">
        <div className="d-flex justify-content-between align-items-start">
          <div className="flex-grow-1" style={{ minWidth: 0 }}>
            <h6 className="mb-0 text-truncate">{title}</h6>
            {subtitle && (
              <small className="text-muted text-truncate d-block">{subtitle}</small>
            )}
          </div>
          {onClose && (
            <button
              type="button"
              className="btn-close ms-2"
              onClick={onClose}
              aria-label="Close inspector"
            />
          )}
        </div>
        {actions && (
          <div className="mt-2 d-flex gap-2 flex-wrap">
            {actions}
          </div>
        )}
      </div>

      {/* Content */}
      <div className="inspector-content flex-grow-1 overflow-auto p-3">
        {loading ? <SkeletonContent /> : children}
      </div>

      <style>{`
        .inspector-panel {
          background: var(--bs-body-bg);
        }
      `}</style>
    </div>
  )
}

// Helper component for inspector sections
export interface InspectorSectionProps {
  title: string
  children: React.ReactNode
  collapsible?: boolean
  defaultOpen?: boolean
}

export function InspectorSection({
  title,
  children,
  collapsible = false,
  defaultOpen = true,
}: InspectorSectionProps) {
  const [open, setOpen] = useState(defaultOpen)

  if (!collapsible) {
    return (
      <div className="inspector-section mb-3">
        <h6 className="text-muted small text-uppercase mb-2">{title}</h6>
        {children}
      </div>
    )
  }

  return (
    <div className="inspector-section mb-3">
      <button
        type="button"
        className="btn btn-link p-0 text-decoration-none d-flex align-items-center gap-1 text-muted small text-uppercase mb-2"
        onClick={() => setOpen(!open)}
      >
        <span style={{ transform: open ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }}>
          ›
        </span>
        {title}
      </button>
      {open && children}
    </div>
  )
}
