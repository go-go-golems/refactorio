import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useGetDocTermsQuery, useGetDocHitsQuery } from '../api/client'
import { useSessionContext } from '../hooks/useSessionContext'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel, InspectorSection } from '../components/detail/InspectorPanel'
import type { DocTerm } from '../types/api'
import { parseDocDrillInParams } from '../features/search-drill-in'

const columns: Column<DocTerm>[] = [
  { key: 'term', header: 'Term', render: (t) => <span className="fw-medium">{t.term}</span>, sortable: true },
  { key: 'count', header: 'Count', width: '80px', render: (t) => <span className="text-muted">{t.count}</span>, sortable: true },
]

export function DocsPage() {
  const { workspaceId, sessionId, activeSession } = useSessionContext()
  const [searchParams] = useSearchParams()
  const drillIn = parseDocDrillInParams(searchParams)
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<DocTerm | null>(null)
  const limit = 50
  const docsRunId = drillIn.runId ?? activeSession?.runs.doc_hits
  const docsAvailable = Boolean(docsRunId)

  const { data: terms, isLoading } = useGetDocTermsQuery(
    { workspace_id: workspaceId!, run_id: docsRunId, limit, offset },
    { skip: !workspaceId || !docsRunId },
  )
  const termRows = docsAvailable ? (terms ?? []) : []

  const { data: hits, isFetching: hitsLoading } = useGetDocHitsQuery(
    { workspace_id: workspaceId!, run_id: docsRunId, term: selected?.term, path: drillIn.path },
    { skip: !selected || !workspaceId || !docsRunId },
  )

  useEffect(() => {
    setSelected(null)
    setOffset(0)
  }, [sessionId])

  useEffect(() => {
    if (!drillIn.term) return
    const fromRows = termRows.find((term) => term.term === drillIn.term)
    if (fromRows) {
      setSelected(fromRows)
      return
    }
    setSelected({ term: drillIn.term, count: 0 })
  }, [drillIn.term, termRows])

  useEffect(() => {
    if (!hits || hits.length === 0 || !drillIn.path || !drillIn.line) return
    const targetId = `doc-hit-${drillIn.path}:${drillIn.line}:${drillIn.col ?? ''}`
    const node = document.getElementById(targetId)
    if (node) {
      node.scrollIntoView({ block: 'center', behavior: 'smooth' })
    }
  }, [hits, drillIn.path, drillIn.line, drillIn.col])

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <h4 className="mb-3">Docs / Terms</h4>
        {drillIn.term && !hitsLoading && selected && hits && hits.length === 0 && (
          <div className="alert alert-warning py-2">
            Target doc hit for term <code>{drillIn.term}</code> was not found in the current scope.
          </div>
        )}
        <EntityTable
          columns={columns}
          data={termRows}
          loading={isLoading}
          selectedId={selected?.term}
          onSelect={setSelected}
          getItemId={(t) => t.term}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage={docsAvailable ? 'No document terms found' : 'No doc hit data for this session'}
        />
      </div>
      {selected && (
        <div style={{ width: 380, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel title={selected.term} subtitle={`${selected.count} hits`} onClose={() => setSelected(null)} loading={hitsLoading}>
            {hits && hits.length > 0 ? (
              <InspectorSection title="Hits" defaultOpen>
                <div style={{ maxHeight: 400, overflowY: 'auto' }}>
                  {hits.map((hit, i) => (
                    <div
                      id={`doc-hit-${hit.path}:${hit.line}:${hit.col}`}
                      key={`${hit.path}:${hit.line}:${hit.col}:${i}`}
                      className={`border-bottom py-2 px-1 ${
                        drillIn.path === hit.path &&
                        drillIn.line === hit.line &&
                        (drillIn.col == null || drillIn.col === hit.col)
                          ? 'bg-warning-subtle'
                          : ''
                      }`}
                    >
                      <div className="font-monospace small text-truncate">{hit.path}:{hit.line}</div>
                      <div className="small text-muted">{hit.match_text}</div>
                    </div>
                  ))}
                </div>
              </InspectorSection>
            ) : !hitsLoading ? (
              <div className="p-3 text-muted small">No hits found</div>
            ) : null}
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
