import { useEffect, useState } from 'react'
import { useGetCommitsQuery, useGetCommitFilesQuery } from '../api/client'
import { useSessionContext } from '../hooks/useSessionContext'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel } from '../components/detail/InspectorPanel'
import { CommitDetail } from '../components/detail/CommitDetail'
import type { Commit } from '../types/api'

const columns: Column<Commit>[] = [
  { key: 'hash', header: 'Hash', width: '90px', render: (c) => <span className="font-monospace">{c.hash.slice(0, 7)}</span> },
  { key: 'subject', header: 'Subject', render: (c) => <span>{c.subject}</span>, sortable: true },
  { key: 'author', header: 'Author', render: (c) => <span className="small">{c.author_name ?? 'Unknown'}</span>, sortable: true },
  { key: 'date', header: 'Date', width: '180px', render: (c) => <span className="small">{new Date(c.committer_date ?? c.author_date ?? '').toLocaleString()}</span>, sortable: true },
]

export function CommitsPage() {
  const { workspaceId, sessionId, activeSession } = useSessionContext()
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<Commit | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const limit = 50
  const commitsRunId = activeSession?.runs.commits
  const commitsAvailable = Boolean(commitsRunId)

  const { data: commits, isLoading } = useGetCommitsQuery(
    { workspace_id: workspaceId!, run_id: commitsRunId, limit, offset, q: searchQuery || undefined },
    { skip: !workspaceId || !commitsRunId },
  )

  const { data: files, isFetching: filesLoading } = useGetCommitFilesQuery(
    { hash: selected?.hash ?? '', workspace_id: workspaceId! },
    { skip: !selected || !workspaceId },
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
          loading={isLoading && commitsAvailable}
          selectedId={selected?.hash}
          onSelect={setSelected}
          getItemId={(c) => c.hash}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage={commitsAvailable ? 'No commits found' : 'No commits data for this session'}
        />
      </div>
      {selected && (
        <div style={{ width: 400, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)', overflowY: 'auto' }}>
          <InspectorPanel title={selected.subject ?? 'Commit'} subtitle={selected.hash.slice(0, 7)} onClose={() => setSelected(null)} loading={filesLoading}>
            <CommitDetail commit={selected} files={files} />
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
