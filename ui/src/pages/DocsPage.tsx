import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetDocTermsQuery, useGetDocHitsQuery } from '../api/client'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel, InspectorSection } from '../components/detail/InspectorPanel'
import type { DocTerm } from '../types/api'

const columns: Column<DocTerm>[] = [
  { key: 'term', header: 'Term', render: (t) => <span className="fw-medium">{t.term}</span>, sortable: true },
  { key: 'count', header: 'Count', width: '80px', render: (t) => <span className="text-muted">{t.count}</span>, sortable: true },
]

export function DocsPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<DocTerm | null>(null)
  const limit = 50

  const { data: terms, isLoading } = useGetDocTermsQuery(
    { workspace_id: workspaceId!, limit, offset },
    { skip: !workspaceId },
  )

  const { data: hits, isFetching: hitsLoading } = useGetDocHitsQuery(
    { workspace_id: workspaceId!, term: selected?.term },
    { skip: !selected || !workspaceId },
  )

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <h4 className="mb-3">Docs / Terms</h4>
        <EntityTable
          columns={columns}
          data={terms ?? []}
          loading={isLoading}
          selectedId={selected?.term}
          onSelect={setSelected}
          getItemId={(t) => t.term}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage="No document terms found"
        />
      </div>
      {selected && (
        <div style={{ width: 380, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel title={selected.term} subtitle={`${selected.count} hits`} onClose={() => setSelected(null)} loading={hitsLoading}>
            {hits && hits.length > 0 ? (
              <InspectorSection title="Hits" defaultOpen>
                <div style={{ maxHeight: 400, overflowY: 'auto' }}>
                  {hits.map((hit, i) => (
                    <div key={i} className="border-bottom py-2 px-1">
                      <div className="font-monospace small text-truncate">{hit.file_path}:{hit.line}</div>
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
