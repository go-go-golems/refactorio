import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetCodeUnitsQuery, useGetCodeUnitQuery } from '../api/client'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel } from '../components/detail/InspectorPanel'
import { CodeUnitDetail } from '../components/detail/CodeUnitDetail'
import { EntityIcon } from '../components/foundation/EntityIcon'
import type { CodeUnit } from '../types/api'

const columns: Column<CodeUnit>[] = [
  { key: 'kind_icon', header: '', width: '32px', render: (u) => <EntityIcon type="code_unit" kind={u.kind} size="sm" /> },
  { key: 'name', header: 'Name', render: (u) => <span className="font-monospace fw-medium">{u.receiver ? `(${u.receiver}).${u.name}` : u.name}</span>, sortable: true },
  { key: 'kind', header: 'Kind', width: '80px', render: (u) => <span className="small text-muted">{u.kind}</span> },
  { key: 'package_path', header: 'Package', render: (u) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 200 }}>{u.package_path}</span> },
  { key: 'location', header: 'File', render: (u) => <span className="font-monospace small">{u.file_path}:{u.start_line}</span> },
]

export function CodeUnitsPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<CodeUnit | null>(null)
  const [kindFilter, setKindFilter] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 50

  const { data: units, isLoading } = useGetCodeUnitsQuery(
    { workspace_id: workspaceId!, limit, offset, kind: kindFilter || undefined, q: searchQuery || undefined },
    { skip: !workspaceId },
  )

  const { data: detail, isFetching: detailLoading } = useGetCodeUnitQuery(
    { hash: selected?.code_unit_hash ?? '', workspace_id: workspaceId! },
    { skip: !selected || !workspaceId },
  )

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
        <EntityTable columns={columns} data={units ?? []} loading={isLoading} selectedId={selected?.code_unit_hash} onSelect={setSelected} getItemId={(u) => u.code_unit_hash} pagination={{ limit, offset, onChange: setOffset }} emptyMessage="No code units found" />
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
