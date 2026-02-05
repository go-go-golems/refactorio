export interface StatusBadgeProps {
  /** Status type */
  status: 'success' | 'running' | 'failed' | 'pending' | 'warning'
  /** Optional custom label */
  label?: string
  /** Size variant */
  size?: 'sm' | 'md'
  /** Show pulse animation for running status */
  pulse?: boolean
  /** Custom class name */
  className?: string
}

const statusConfig = {
  success: { label: 'Success', color: 'success', icon: '✓' },
  running: { label: 'Running', color: 'primary', icon: '⟳' },
  failed: { label: 'Failed', color: 'danger', icon: '✕' },
  pending: { label: 'Pending', color: 'secondary', icon: '○' },
  warning: { label: 'Warning', color: 'warning', icon: '⚠' },
}

export function StatusBadge({
  status,
  label,
  size = 'md',
  pulse = true,
  className = '',
}: StatusBadgeProps) {
  const config = statusConfig[status]
  const displayLabel = label ?? config.label
  const sizeClass = size === 'sm' ? 'badge-sm' : ''
  const pulseClass = pulse && status === 'running' ? 'status-pulse' : ''

  return (
    <span
      className={`badge bg-${config.color} d-inline-flex align-items-center gap-1 ${sizeClass} ${pulseClass} ${className}`}
      style={{
        fontSize: size === 'sm' ? '0.7rem' : '0.8rem',
        padding: size === 'sm' ? '0.2em 0.5em' : '0.35em 0.65em',
      }}
    >
      <span className="status-icon">{config.icon}</span>
      <span>{displayLabel}</span>
      <style>{`
        .status-pulse {
          animation: pulse 1.5s ease-in-out infinite;
        }
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.6; }
        }
      `}</style>
    </span>
  )
}
