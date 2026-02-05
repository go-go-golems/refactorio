import type { Commit, CommitFile } from '../../types/api'
import { CopyButton, EntityIcon, StatusBadge } from '../foundation'

export interface CommitDetailProps {
  /** Commit data */
  commit: Commit
  /** Changed files */
  files?: CommitFile[]
  /** Called when a file is clicked */
  onFileClick?: (file: CommitFile) => void
  /** Called to view the full diff */
  onViewDiff?: () => void
}

function fileStatusBadge(status: string) {
  switch (status) {
    case 'A':
      return <StatusBadge status="success" label="added" size="sm" />
    case 'D':
      return <StatusBadge status="failed" label="deleted" size="sm" />
    case 'M':
      return <StatusBadge status="warning" label="modified" size="sm" />
    case 'R':
      return <StatusBadge status="pending" label="renamed" size="sm" />
    default:
      return <StatusBadge status="pending" label={status} size="sm" />
  }
}

export function CommitDetail({
  commit,
  files,
  onFileClick,
  onViewDiff,
}: CommitDetailProps) {
  return (
    <div className="commit-detail">
      {/* Header */}
      <div className="d-flex align-items-start gap-2 mb-3">
        <EntityIcon type="commit" size="md" />
        <div className="flex-grow-1" style={{ minWidth: 0 }}>
          <span className="fw-semibold text-break">{commit.subject ?? 'Commit'}</span>
          <div className="d-flex align-items-center gap-2 mt-1">
            <span className="font-monospace small">{commit.hash.slice(0, 7)}</span>
            <CopyButton text={commit.hash} size="sm" />
          </div>
        </div>
      </div>

      {/* Metadata */}
      <div className="mb-3">
        <div className="d-flex gap-2 mb-1">
          <span className="text-muted small" style={{ minWidth: 60 }}>Author</span>
          <span className="small">
            {commit.author_name ?? 'Unknown'}
            {commit.author_email ? ` <${commit.author_email}>` : ''}
          </span>
        </div>
        <div className="d-flex gap-2 mb-1">
          <span className="text-muted small" style={{ minWidth: 60 }}>Date</span>
          <span className="small">{new Date(commit.committer_date ?? commit.author_date ?? '').toLocaleString()}</span>
        </div>
        <div className="d-flex gap-2 mb-1">
          <span className="text-muted small" style={{ minWidth: 60 }}>Run</span>
          <span className="small">#{commit.run_id}</span>
        </div>
      </div>

      {/* Body */}
      {commit.body && (
        <div className="mb-3">
          <div className="bg-body-tertiary p-2 rounded small" style={{ whiteSpace: 'pre-wrap' }}>
            {commit.body}
          </div>
        </div>
      )}

      {/* Actions */}
      {onViewDiff && (
        <div className="mb-3">
          <button type="button" className="btn btn-outline-secondary btn-sm" onClick={onViewDiff}>
            View Diff
          </button>
        </div>
      )}

      {/* Changed files */}
      {files && files.length > 0 && (
        <div>
          <div className="d-flex justify-content-between align-items-center mb-2">
            <label className="text-muted small text-uppercase mb-0">
              Changed Files ({files.length})
            </label>
            <span className="small text-muted">Status</span>
          </div>
          <div style={{ maxHeight: 300, overflowY: 'auto' }}>
            {files.map((file) => (
              <button
                key={file.path}
                type="button"
                className="btn btn-sm w-100 text-start d-flex align-items-center gap-2 border-bottom rounded-0 py-1"
                onClick={() => onFileClick?.(file)}
                disabled={!onFileClick}
              >
                {fileStatusBadge(file.status)}
                <span className="font-monospace small text-truncate flex-grow-1">
                  {file.path}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
