import type { Symbol, SymbolRef } from '../../types/api'
import { CopyButton, StatusBadge, EntityIcon } from '../foundation'

export interface SymbolDetailProps {
  /** Symbol data */
  symbol: Symbol
  /** Symbol references (from gopls) */
  refs?: SymbolRef[]
  /** Whether refs are being loaded */
  refsLoading?: boolean
  /** Whether refs are available for this symbol */
  refsAvailable?: boolean
  /** Called to trigger ref computation */
  onComputeRefs?: () => void
  /** Called to open in editor */
  onOpenInEditor?: () => void
  /** Called to add to refactor plan */
  onAddToPlan?: () => void
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="d-flex gap-2 mb-1">
      <span className="text-muted small" style={{ minWidth: 80, flexShrink: 0 }}>{label}</span>
      <span className="small text-break">{children}</span>
    </div>
  )
}

function RefItem({ ref: symbolRef }: { ref: SymbolRef }) {
  return (
    <div className="d-flex align-items-center gap-2 py-1 border-bottom">
      <span className="small font-monospace text-truncate flex-grow-1">
        {symbolRef.file_path}:{symbolRef.start_line}
      </span>
      {symbolRef.is_declaration && (
        <StatusBadge status="success" label="decl" size="sm" />
      )}
    </div>
  )
}

export function SymbolDetail({
  symbol,
  refs,
  refsLoading = false,
  refsAvailable = false,
  onComputeRefs,
  onOpenInEditor,
  onAddToPlan,
}: SymbolDetailProps) {
  return (
    <div className="symbol-detail">
      {/* Header */}
      <div className="d-flex align-items-start gap-2 mb-3">
        <EntityIcon type="symbol" kind={symbol.kind} size="md" />
        <div className="flex-grow-1" style={{ minWidth: 0 }}>
          <div className="d-flex align-items-center gap-2">
            <span className="fw-semibold text-break">{symbol.name}</span>
            {symbol.exported ? (
              <StatusBadge status="success" label="exported" size="sm" />
            ) : (
              <StatusBadge status="warning" label="unexported" size="sm" />
            )}
          </div>
          <small className="text-muted">{symbol.kind}</small>
        </div>
        <CopyButton text={symbol.name} size="sm" />
      </div>

      {/* Signature */}
      {symbol.signature && (
        <div className="mb-3">
          <label className="text-muted small text-uppercase mb-1 d-block">Signature</label>
          <code className="d-block bg-body-tertiary p-2 rounded small text-break" style={{ whiteSpace: 'pre-wrap' }}>
            {symbol.signature}
          </code>
        </div>
      )}

      {/* Fields */}
      <div className="mb-3">
        <FieldRow label="Package">
          <span className="font-monospace">{symbol.package_path}</span>
          <CopyButton text={symbol.package_path} size="sm" variant="icon" />
        </FieldRow>
        <FieldRow label="File">
          <span className="font-monospace">{symbol.file_path}:{symbol.start_line}</span>
          <CopyButton text={`${symbol.file_path}:${symbol.start_line}`} size="sm" variant="icon" />
        </FieldRow>
        <FieldRow label="Range">
          L{symbol.start_line}:{symbol.start_col} &ndash; L{symbol.end_line}:{symbol.end_col}
        </FieldRow>
        <FieldRow label="Run">
          #{symbol.run_id}
        </FieldRow>
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

      {/* References */}
      <div>
        <div className="d-flex justify-content-between align-items-center mb-2">
          <label className="text-muted small text-uppercase mb-0">
            References {refs && `(${refs.length})`}
          </label>
        </div>

        {refsLoading && (
          <div className="placeholder-glow">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="mb-1">
                <span className="placeholder col-8" style={{ height: 14 }} />
              </div>
            ))}
          </div>
        )}

        {!refsLoading && refs && refs.length > 0 && (
          <div style={{ maxHeight: 200, overflowY: 'auto' }}>
            {refs.map((r, i) => (
              <RefItem key={`${r.file_path}:${r.start_line}:${i}`} ref={r} />
            ))}
          </div>
        )}

        {!refsLoading && refs && refs.length === 0 && (
          <p className="text-muted small mb-0">No references found</p>
        )}

        {!refsLoading && !refs && !refsAvailable && onComputeRefs && (
          <div className="text-center py-2">
            <p className="text-muted small mb-2">References not computed yet</p>
            <button type="button" className="btn btn-outline-primary btn-sm" onClick={onComputeRefs}>
              Compute References
            </button>
          </div>
        )}

        {!refsLoading && !refs && refsAvailable && (
          <p className="text-muted small mb-0">References available. Select to load.</p>
        )}
      </div>
    </div>
  )
}
