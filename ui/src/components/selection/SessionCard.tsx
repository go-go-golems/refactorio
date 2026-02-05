import type { Session, SessionAvailability } from '../../types/api'
import { StatusBadge } from '../foundation'

export interface SessionCardProps {
  /** Session data */
  session: Session
  /** Whether this card is selected */
  selected?: boolean
  /** Click handler */
  onClick?: () => void
  /** Edit handler */
  onEdit?: () => void
}

const domainLabels: Record<keyof SessionAvailability, string> = {
  commits: 'Commits',
  diff: 'Diffs',
  symbols: 'Symbols',
  code_units: 'Code Units',
  doc_hits: 'Doc Hits',
  gopls_refs: 'Gopls Refs',
  tree_sitter: 'Tree-sitter',
}

function AvailabilityRow({
  domain,
  available,
  runId,
}: {
  domain: string
  available: boolean
  runId?: number
}) {
  return (
    <div className="d-flex align-items-center justify-content-between py-1">
      <span className="small">{domain}</span>
      <span className="d-flex align-items-center gap-2">
        {runId && <span className="text-muted" style={{ fontSize: '0.7rem' }}>#{runId}</span>}
        <StatusBadge
          status={available ? 'success' : 'warning'}
          label={available ? 'available' : 'missing'}
          size="sm"
        />
      </span>
    </div>
  )
}

export function SessionCard({
  session,
  selected = false,
  onClick,
  onEdit,
}: SessionCardProps) {
  const gitRange = [session.git_from, session.git_to].filter(Boolean).join(' → ')
  const availableCount = Object.values(session.availability).filter(Boolean).length
  const totalDomains = Object.keys(session.availability).length

  return (
    <div
      className={`card ${selected ? 'border-primary shadow-sm' : ''} ${onClick ? 'cursor-pointer' : ''}`}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => { if (e.key === 'Enter') onClick() } : undefined}
    >
      <div className="card-body">
        <div className="d-flex justify-content-between align-items-start mb-2">
          <div>
            <h6 className="card-title mb-0">
              {gitRange || 'Unnamed Session'}
            </h6>
            <small className="text-muted text-truncate d-block" style={{ maxWidth: '300px' }}>
              {session.root_path}
            </small>
          </div>
          {onEdit && (
            <button
              type="button"
              className="btn btn-outline-secondary btn-sm"
              onClick={(e) => { e.stopPropagation(); onEdit() }}
            >
              Edit
            </button>
          )}
        </div>

        <div className="mb-2">
          <small className="text-muted">
            Data: {availableCount}/{totalDomains} domains
          </small>
          <div className="progress mt-1" style={{ height: '4px' }}>
            <div
              className="progress-bar bg-success"
              style={{ width: `${(availableCount / totalDomains) * 100}%` }}
            />
          </div>
        </div>

        <div className="border-top pt-2">
          {(Object.entries(session.availability) as [keyof SessionAvailability, boolean][]).map(
            ([domain, available]) => (
              <AvailabilityRow
                key={domain}
                domain={domainLabels[domain]}
                available={available}
                runId={session.runs[domain as keyof typeof session.runs]}
              />
            )
          )}
        </div>

        <div className="text-end mt-2">
          <small className="text-muted">
            Updated {new Date(session.last_updated).toLocaleString()}
          </small>
        </div>
      </div>

      <style>{`
        .cursor-pointer { cursor: pointer; }
        .cursor-pointer:hover { background-color: var(--bs-tertiary-bg); }
      `}</style>
    </div>
  )
}
