import type { Session, SessionAvailability } from '../../types/api'
import { StatusBadge } from '../foundation'

export interface SessionSelectorProps {
  /** Available sessions */
  sessions: Session[]
  /** Currently selected session */
  selected?: Session
  /** Called when a session is selected */
  onSelect: (session: Session) => void
  /** Whether the list is loading */
  loading?: boolean
}

function availabilityCount(a: SessionAvailability): number {
  return Object.values(a).filter(Boolean).length
}

function sessionLabel(s: Session): string {
  const range = [s.git_from, s.git_to].filter(Boolean).join(' \u2192 ')
  return range || 'Unnamed Session'
}

export function SessionSelector({
  sessions,
  selected,
  onSelect,
  loading = false,
}: SessionSelectorProps) {
  if (loading) {
    return (
      <div className="placeholder-glow">
        <span className="placeholder col-8" style={{ height: 32 }} />
      </div>
    )
  }

  if (sessions.length === 0) {
    return (
      <span className="text-muted small">No sessions available</span>
    )
  }

  if (sessions.length === 1) {
    const s = sessions[0]
    const count = availabilityCount(s.availability)
    return (
      <div className="d-flex align-items-center gap-2">
        <span className="fw-medium small">{sessionLabel(s)}</span>
        <StatusBadge
          status={count >= 5 ? 'success' : count >= 3 ? 'warning' : 'failed'}
          label={`${count}/7`}
          size="sm"
        />
      </div>
    )
  }

  return (
    <div className="dropdown">
      <button
        className="btn btn-outline-secondary btn-sm dropdown-toggle d-flex align-items-center gap-2"
        type="button"
        data-bs-toggle="dropdown"
        aria-expanded="false"
      >
        {selected ? (
          <>
            <span className="text-truncate" style={{ maxWidth: 200 }}>
              {sessionLabel(selected)}
            </span>
            <StatusBadge
              status={availabilityCount(selected.availability) >= 5 ? 'success' : 'warning'}
              label={`${availabilityCount(selected.availability)}/7`}
              size="sm"
            />
          </>
        ) : (
          'Select session'
        )}
      </button>
      <ul className="dropdown-menu" style={{ maxHeight: 300, overflowY: 'auto' }}>
        {sessions.map((s) => {
          const count = availabilityCount(s.availability)
          const isActive = selected?.id === s.id
          return (
            <li key={s.id}>
              <button
                type="button"
                className={`dropdown-item d-flex justify-content-between align-items-center ${isActive ? 'active' : ''}`}
                onClick={() => onSelect(s)}
              >
                <span className="text-truncate me-2">{sessionLabel(s)}</span>
                <StatusBadge
                  status={count >= 5 ? 'success' : count >= 3 ? 'warning' : 'failed'}
                  label={`${count}/7`}
                  size="sm"
                />
              </button>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
