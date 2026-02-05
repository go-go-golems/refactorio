import type { CodeUnit, CodeUnitDetail as CodeUnitDetailType } from '../../types/api'
import { CopyButton, EntityIcon } from '../foundation'
import { CodeViewer } from '../code-display/CodeViewer'

export interface CodeUnitDetailProps {
  /** Code unit with body */
  codeUnit: CodeUnitDetailType
  /** Historical versions of this code unit */
  history?: CodeUnit[]
  /** Called to compute diff between two versions */
  onDiff?: (hash1: string, hash2: string) => void
  /** Called to add to refactor plan */
  onAddToPlan?: () => void
  /** Called to open in editor */
  onOpenInEditor?: () => void
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="d-flex gap-2 mb-1">
      <span className="text-muted small" style={{ minWidth: 80, flexShrink: 0 }}>{label}</span>
      <span className="small text-break">{children}</span>
    </div>
  )
}

export function CodeUnitDetail({
  codeUnit,
  history,
  onDiff,
  onAddToPlan,
  onOpenInEditor,
}: CodeUnitDetailProps) {
  const displayName = codeUnit.receiver
    ? `(${codeUnit.receiver}).${codeUnit.name}`
    : codeUnit.name

  return (
    <div className="code-unit-detail">
      {/* Header */}
      <div className="d-flex align-items-start gap-2 mb-3">
        <EntityIcon type="code_unit" kind={codeUnit.kind} size="md" />
        <div className="flex-grow-1" style={{ minWidth: 0 }}>
          <span className="fw-semibold text-break font-monospace">{displayName}</span>
          <small className="text-muted d-block">{codeUnit.kind}</small>
        </div>
        <CopyButton text={codeUnit.name} size="sm" />
      </div>

      {/* Fields */}
      <div className="mb-3">
        <FieldRow label="Package">
          <span className="font-monospace">{codeUnit.package_path}</span>
        </FieldRow>
        <FieldRow label="File">
          <span className="font-monospace">{codeUnit.file_path}:{codeUnit.start_line}</span>
          <CopyButton text={`${codeUnit.file_path}:${codeUnit.start_line}`} size="sm" variant="icon" />
        </FieldRow>
        <FieldRow label="Range">
          L{codeUnit.start_line}:{codeUnit.start_col} &ndash; L{codeUnit.end_line}:{codeUnit.end_col}
        </FieldRow>
        <FieldRow label="Body Hash">
          <span className="font-monospace">{codeUnit.body_hash.slice(0, 12)}</span>
          <CopyButton text={codeUnit.body_hash} size="sm" variant="icon" />
        </FieldRow>
      </div>

      {/* Doc comment */}
      {codeUnit.doc_comment && (
        <div className="mb-3">
          <label className="text-muted small text-uppercase mb-1 d-block">Doc Comment</label>
          <div className="bg-body-tertiary p-2 rounded small" style={{ whiteSpace: 'pre-wrap' }}>
            {codeUnit.doc_comment}
          </div>
        </div>
      )}

      {/* Body */}
      <div className="mb-3">
        <label className="text-muted small text-uppercase mb-1 d-block">Body</label>
        <CodeViewer
          content={codeUnit.body}
          language="go"
          startLine={codeUnit.start_line}
          maxHeight={300}
        />
      </div>

      {/* Actions */}
      <div className="d-flex gap-2 mb-3">
        {onOpenInEditor && (
          <button type="button" className="btn btn-outline-secondary btn-sm" onClick={onOpenInEditor}>
            Open in Editor
          </button>
        )}
        {onAddToPlan && (
          <button type="button" className="btn btn-outline-primary btn-sm" onClick={onAddToPlan}>
            + Add to Plan
          </button>
        )}
      </div>

      {/* History */}
      {history && history.length > 0 && (
        <div>
          <label className="text-muted small text-uppercase mb-2 d-block">
            History ({history.length} versions)
          </label>
          <div style={{ maxHeight: 200, overflowY: 'auto' }}>
            {history.map((ver, i) => (
              <div
                key={`${ver.code_unit_hash}-${ver.run_id}`}
                className="d-flex align-items-center justify-content-between py-1 border-bottom"
              >
                <div className="small">
                  <span className="font-monospace">{ver.body_hash.slice(0, 8)}</span>
                  <span className="text-muted ms-2">run #{ver.run_id}</span>
                </div>
                {onDiff && i < history.length - 1 && (
                  <button
                    type="button"
                    className="btn btn-link btn-sm p-0"
                    onClick={() => onDiff(history[i + 1].code_unit_hash, ver.code_unit_hash)}
                  >
                    Diff
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
