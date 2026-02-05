import { useEffect, useState } from 'react'
import { useGetCodeUnitsQuery, useGetCodeUnitQuery } from '../api/client'
import { useSessionContext } from '../hooks/useSessionContext'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel } from '../components/detail/InspectorPanel'
import { CodeUnitDetail } from '../components/detail/CodeUnitDetail'
import { EntityIcon } from '../components/foundation/EntityIcon'
import type { CodeUnit } from '../types/api'

const columns: Column<CodeUnit>[] = [
  { key: 'kind_icon', header: '', width: '32px', render: (u) => <EntityIcon type="code_unit" kind={u.kind} size="sm" /> },
  { key: 'name', header: 'Name', render: (u) => <span className="font-monospace fw-medium">{u.recv ? `(${u.recv}).${u.name}` : u.name}</span>, sortable: true },
  { key: 'kind', header: 'Kind', width: '80px', render: (u) => <span className="small text-muted">{u.kind}</span> },
  { key: 'pkg', header: 'Package', render: (u) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 200 }}>{u.pkg}</span> },
  { key: 'location', header: 'File', render: (u) => <span className="font-monospace small">{u.file}:{u.start_line}</span> },
]

export function CodeUnitsPage() {
  const { workspaceId, sessionId, activeSession } = useSessionContext()
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<CodeUnit | null>(null)
  const [kindFilter, setKindFilter] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 50
  const codeUnitsRunId = activeSession?.runs.code_units
  const codeUnitsAvailable = Boolean(codeUnitsRunId)

  const { data: units, isLoading } = useGetCodeUnitsQuery(
    { workspace_id: workspaceId!, run_id: codeUnitsRunId, limit, offset, kind: kindFilter || undefined, name: searchQuery || undefined },
    { skip: !workspaceId || !codeUnitsRunId },
  )

  const { data: detail, isFetching: detailLoading } = useGetCodeUnitQuery(
    { hash: selected?.unit_hash ?? '', workspace_id: workspaceId!, run_id: codeUnitsRunId },
    { skip: !selected || !workspaceId || !codeUnitsRunId },
  )

  useEffect(() => {
    setSelected(null)
    setOffset(0)
  }, [sessionId])

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h4 className="mb-0">Code Units</h4>
          <div className="d-flex gap-2">
            <select className="form-select form-select-sm" style={{ width: 120 }} value={kindFilter} onChange={(e) => { setKindFilter(e.target.value); setOffset(0) }}>
              <option value="">All kinds</option>
              <option value="func">func</option>
              <option value="type">type</option>
              <option value="method">method</option>
            </select>
            <input type="search" className="form-control form-control-sm" placeholder="Search..." style={{ width: 200 }} value={searchQuery} onChange={(e) => { setSearchQuery(e.target.value); setOffset(0) }} />
          </div>
        </div>
        <EntityTable
          columns={columns}
          data={units ?? []}
          loading={isLoading && codeUnitsAvailable}
          selectedId={selected?.unit_hash}
          onSelect={setSelected}
          getItemId={(u) => u.unit_hash}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage={codeUnitsAvailable ? 'No code units found' : 'No code units data for this session'}
        />
      </div>
      {selected && (
        <div style={{ width: 400, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel title={selected.name} subtitle={selected.kind} onClose={() => setSelected(null)} loading={detailLoading}>
            {detail && <CodeUnitDetail codeUnit={detail} />}
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
