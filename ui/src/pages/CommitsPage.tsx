import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetCommitsQuery, useGetCommitFilesQuery } from '../api/client'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel } from '../components/detail/InspectorPanel'
import { CommitDetail } from '../components/detail/CommitDetail'
import type { Commit } from '../types/api'

const columns: Column<Commit>[] = [
  { key: 'hash', header: 'Hash', width: '90px', render: (c) => <span className="font-monospace">{c.commit_hash.slice(0, 7)}</span> },
  { key: 'subject', header: 'Subject', render: (c) => <span>{c.subject}</span>, sortable: true },
  { key: 'author', header: 'Author', render: (c) => <span className="small">{c.author_name}</span>, sortable: true },
  { key: 'date', header: 'Date', width: '180px', render: (c) => <span className="small">{new Date(c.commit_date).toLocaleString()}</span>, sortable: true },
]

export function CommitsPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<Commit | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 50

  const { data: commits, isLoading } = useGetCommitsQuery(
    { workspace_id: workspaceId!, limit, offset, q: searchQuery || undefined },
    { skip: !workspaceId },
  )

  const { data: files, isFetching: filesLoading } = useGetCommitFilesQuery(
    { hash: selected?.commit_hash ?? '', workspace_id: workspaceId! },
    { skip: !selected || !workspaceId },
  )

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h4 className="mb-0">Commits</h4>
          <input
            type="search"
            className="form-control form-control-sm"
            placeholder="Search commits..."
            style={{ width: 200 }}
            value={searchQuery}
            onChange={(e) => { setSearchQuery(e.target.value); setOffset(0) }}
          />
        </div>
        <EntityTable
          columns={columns}
          data={commits ?? []}
          loading={isLoading}
          selectedId={selected?.commit_hash}
          onSelect={setSelected}
          getItemId={(c) => c.commit_hash}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage="No commits found"
        />
      </div>
      {selected && (
        <div style={{ width: 400, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel title={selected.subject} subtitle={selected.commit_hash.slice(0, 7)} onClose={() => setSelected(null)} loading={filesLoading}>
            <CommitDetail commit={selected} files={files} />
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
