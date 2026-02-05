import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { EntityTable, type Column } from './EntityTable'
import { EntityIcon, StatusBadge, CopyButton } from '../foundation'
import { mockSymbols, mockRuns, mockCommits } from '../../mocks/data'
import type { Symbol, Run, Commit } from '../../types/api'

const meta: Meta<typeof EntityTable> = {
  title: 'DataDisplay/EntityTable',
  component: EntityTable,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
  },
}

export default meta

// Symbol table configuration
const symbolColumns: Column<Symbol>[] = [
  {
    key: 'name',
    header: 'Name',
    sortable: true,
    render: (s) => (
      <div className="d-flex align-items-center gap-2">
        <EntityIcon type="symbol" kind={s.kind} size="sm" />
        <span className="fw-medium">{s.name}</span>
      </div>
    ),
  },
  {
    key: 'kind',
    header: 'Kind',
    width: '80px',
    sortable: true,
    render: (s) => <span className="badge bg-secondary-subtle text-secondary">{s.kind}</span>,
  },
  {
    key: 'pkg',
    header: 'Package',
    render: (s) => <code className="small text-muted">{s.pkg.split('/').pop()}</code>,
  },
  {
    key: 'file',
    header: 'Location',
    render: (s) => (
      <span className="text-muted small">
        {s.file}:{s.line}
      </span>
    ),
  },
  {
    key: 'is_exported',
    header: 'Exp',
    width: '50px',
    render: (s) => (s.is_exported ? '✓' : ''),
  },
]

export const Symbols: StoryObj<typeof EntityTable<Symbol>> = {
  args: {
    columns: symbolColumns,
    data: mockSymbols,
    getItemId: (s) => s.symbol_hash,
  },
}

export const SymbolsWithSelection: StoryObj<typeof EntityTable<Symbol>> = {
  args: {
    columns: symbolColumns,
    data: mockSymbols,
    getItemId: (s) => s.symbol_hash,
    selectedId: 'a7b3c9f2',
    onSelect: () => {},
  },
}

export const SymbolsWithPagination: StoryObj<typeof EntityTable<Symbol>> = {
  args: {
    columns: symbolColumns,
    data: mockSymbols,
    getItemId: (s) => s.symbol_hash,
    pagination: {
      total: 12456,
      limit: 50,
      offset: 0,
      onChange: () => {},
    },
  },
}

// Run table configuration
const runColumns: Column<Run>[] = [
  {
    key: 'id',
    header: 'ID',
    width: '60px',
    render: (r) => <span className="font-monospace">#{r.id}</span>,
  },
  {
    key: 'status',
    header: 'Status',
    width: '100px',
    render: (r) => <StatusBadge status={r.status} size="sm" />,
  },
  {
    key: 'git_range',
    header: 'Git Range',
    render: (r) => (
      <code className="small">
        {r.git_from || '?'}→{r.git_to || '?'}
      </code>
    ),
  },
  {
    key: 'started_at',
    header: 'Started',
    render: (r) => <span className="text-muted small">{new Date(r.started_at).toLocaleString()}</span>,
  },
]

export const Runs: StoryObj<typeof EntityTable<Run>> = {
  args: {
    columns: runColumns,
    data: mockRuns,
    getItemId: (r) => String(r.id),
  },
}

// Commit table configuration
const commitColumns: Column<Commit>[] = [
  {
    key: 'hash',
    header: 'Hash',
    width: '90px',
    render: (c) => (
      <div className="d-flex align-items-center gap-1">
        <code className="small">{c.hash.slice(0, 7)}</code>
        <CopyButton text={c.hash} size="sm" />
      </div>
    ),
  },
  {
    key: 'subject',
    header: 'Subject',
    sortable: true,
    render: (c) => <span className="text-truncate d-block" style={{ maxWidth: '400px' }}>{c.subject}</span>,
  },
  {
    key: 'author_name',
    header: 'Author',
    width: '120px',
    sortable: true,
  },
  {
    key: 'committer_date',
    header: 'Date',
    width: '150px',
    sortable: true,
    render: (c) => <span className="text-muted small">{new Date(c.committer_date ?? c.author_date ?? '').toLocaleDateString()}</span>,
  },
]

export const Commits: StoryObj<typeof EntityTable<Commit>> = {
  args: {
    columns: commitColumns,
    data: mockCommits,
    getItemId: (c) => c.hash,
  },
}

export const Loading: StoryObj<typeof EntityTable<Symbol>> = {
  args: {
    columns: symbolColumns,
    data: [],
    loading: true,
    getItemId: (s) => s.symbol_hash,
  },
}

export const Empty: StoryObj<typeof EntityTable<Symbol>> = {
  args: {
    columns: symbolColumns,
    data: [],
    getItemId: (s) => s.symbol_hash,
    emptyMessage: 'No symbols match your search',
  },
}

export const Interactive: StoryObj<typeof EntityTable<Symbol>> = {
  render: function InteractiveTable() {
    const [selectedId, setSelectedId] = useState<string>()
    const [sortColumn, setSortColumn] = useState<string>('name')
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
    const [offset, setOffset] = useState(0)

    const handleSort = (column: string, direction: 'asc' | 'desc') => {
      setSortColumn(column)
      setSortDirection(direction)
    }

    const sortedData = [...mockSymbols].sort((a, b) => {
      const aVal = String((a as Record<string, unknown>)[sortColumn] ?? '')
      const bVal = String((b as Record<string, unknown>)[sortColumn] ?? '')
      const cmp = aVal.localeCompare(bVal)
      return sortDirection === 'asc' ? cmp : -cmp
    })

    return (
      <div style={{ height: '400px' }}>
        <EntityTable
          columns={symbolColumns}
          data={sortedData}
          getItemId={(s) => s.symbol_hash}
          selectedId={selectedId}
          onSelect={(s) => setSelectedId(s.symbol_hash)}
          sortColumn={sortColumn}
          sortDirection={sortDirection}
          onSort={handleSort}
          pagination={{
            total: 100,
            limit: 10,
            offset,
            onChange: setOffset,
          }}
        />
        {selectedId && (
          <div className="mt-2 p-2 bg-light rounded">
            Selected: <code>{selectedId}</code>
          </div>
        )}
      </div>
    )
  },
}
