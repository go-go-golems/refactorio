import { useEffect, useState } from 'react'
import { useGetSymbolsQuery, useGetSymbolRefsQuery } from '../api/client'
import { useSessionContext } from '../hooks/useSessionContext'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel } from '../components/detail/InspectorPanel'
import { SymbolDetail } from '../components/detail/SymbolDetail'
import { EntityIcon } from '../components/foundation/EntityIcon'
import type { Symbol } from '../types/api'

const columns: Column<Symbol>[] = [
  { key: 'kind', header: '', width: '32px', render: (s) => <EntityIcon type="symbol" kind={s.kind} size="sm" /> },
  { key: 'name', header: 'Name', render: (s) => <span className="font-monospace fw-medium">{s.name}</span>, sortable: true },
  { key: 'kind_label', header: 'Kind', width: '80px', render: (s) => <span className="small text-muted">{s.kind}</span> },
  { key: 'pkg', header: 'Package', render: (s) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 200 }}>{s.pkg}</span> },
  { key: 'file', header: 'File', render: (s) => <span className="font-monospace small">{s.file}:{s.line}</span> },
]

function symbolRowID(s: Symbol): string {
  return `${s.symbol_hash}:${s.file}:${s.line}:${s.col}:${s.run_id}`
}

export function SymbolsPage() {
  const { workspaceId, sessionId, activeSession } = useSessionContext()
  const [offset, setOffset] = useState(0)
  const [selectedSymbol, setSelectedSymbol] = useState<Symbol | null>(null)
  const [kindFilter, setKindFilter] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 50
  const symbolsRunId = activeSession?.runs.symbols
  const refsRunId = activeSession?.runs.gopls_refs ?? activeSession?.runs.symbols
  const symbolsAvailable = Boolean(symbolsRunId)

  const { data: symbols, isLoading } = useGetSymbolsQuery(
    {
      workspace_id: workspaceId!,
      run_id: symbolsRunId,
      limit,
      offset,
      kind: kindFilter || undefined,
      name: searchQuery || undefined,
    },
    { skip: !workspaceId || !symbolsRunId },
  )
  const symbolRows = symbolsAvailable ? (symbols ?? []) : []

  const { data: refs, isFetching: refsLoading } = useGetSymbolRefsQuery(
    { hash: selectedSymbol?.symbol_hash ?? '', workspace_id: workspaceId!, run_id: refsRunId },
    { skip: !selectedSymbol || !workspaceId || !refsRunId },
  )

  useEffect(() => {
    setSelectedSymbol(null)
    setOffset(0)
  }, [sessionId])

  if (!workspaceId) {
    return <div className="p-4 text-muted">Select a workspace first.</div>
  }

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h4 className="mb-0">Symbols</h4>
          <div className="d-flex gap-2">
            <select className="form-select form-select-sm" style={{ width: 120 }} value={kindFilter} onChange={(e) => { setKindFilter(e.target.value); setOffset(0) }}>
              <option value="">All kinds</option>
              <option value="func">func</option>
              <option value="type">type</option>
              <option value="method">method</option>
              <option value="const">const</option>
              <option value="var">var</option>
            </select>
            <input
              type="search"
              className="form-control form-control-sm"
              placeholder="Search symbols..."
              style={{ width: 200 }}
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setOffset(0) }}
            />
          </div>
        </div>
        <EntityTable
          columns={columns}
          data={symbolRows}
          loading={isLoading && symbolsAvailable}
          selectedId={selectedSymbol ? symbolRowID(selectedSymbol) : undefined}
          onSelect={(s) => setSelectedSymbol(s)}
          getItemId={symbolRowID}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage={symbolsAvailable ? 'No symbols found' : 'No symbol data for this session'}
        />
      </div>
      {selectedSymbol && (
        <div style={{ width: 380, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel
            title={selectedSymbol.name}
            subtitle={selectedSymbol.kind}
            onClose={() => setSelectedSymbol(null)}
          >
            <SymbolDetail
              symbol={selectedSymbol}
              refs={refs}
              refsLoading={refsLoading}
              refsAvailable={Boolean(refsRunId)}
            />
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
