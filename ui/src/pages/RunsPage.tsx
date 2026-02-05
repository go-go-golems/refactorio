import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetRunsQuery, useGetRunSummaryQuery } from '../api/client'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { InspectorPanel, InspectorSection } from '../components/detail/InspectorPanel'
import { StatusBadge } from '../components/foundation/StatusBadge'
import type { Run } from '../types/api'

const columns: Column<Run>[] = [
  { key: 'id', header: 'ID', width: '60px', render: (r) => <span className="font-monospace">#{r.id}</span> },
  {
    key: 'status', header: 'Status', width: '90px',
    render: (r) => <StatusBadge status={r.status === 'success' ? 'success' : r.status === 'failed' ? 'failed' : 'running'} label={r.status} size="sm" />,
  },
  { key: 'root_path', header: 'Root Path', render: (r) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 200 }}>{r.root_path}</span> },
  { key: 'git_range', header: 'Git Range', render: (r) => <span className="small">{[r.git_from, r.git_to].filter(Boolean).join(' \u2192 ')}</span> },
  { key: 'started_at', header: 'Started', render: (r) => <span className="small">{new Date(r.started_at).toLocaleString()}</span>, sortable: true },
]

export function RunsPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [offset, setOffset] = useState(0)
  const [selectedRun, setSelectedRun] = useState<Run | null>(null)
  const limit = 50

  const { data: runs, isLoading } = useGetRunsQuery(
    { workspace_id: workspaceId!, limit, offset },
    { skip: !workspaceId },
  )
  const { data: summary } = useGetRunSummaryQuery(
    { id: selectedRun?.id ?? 0, workspace_id: workspaceId! },
    { skip: !selectedRun || !workspaceId },
  )

  if (!workspaceId) {
    return <div className="p-4 text-muted">Select a workspace first.</div>
  }

  return (
    <div className="d-flex h-100">
      <div className="flex-grow-1 p-4 overflow-auto">
        <h4 className="mb-3">Runs</h4>
        <EntityTable
          columns={columns}
          data={runs ?? []}
          loading={isLoading}
          selectedId={selectedRun ? String(selectedRun.id) : undefined}
          onSelect={(run) => setSelectedRun(run)}
          getItemId={(r) => String(r.id)}
          pagination={{ limit, offset, onChange: setOffset }}
          emptyMessage="No runs found"
        />
      </div>
      {selectedRun && (
        <div style={{ width: 320, flexShrink: 0, borderLeft: '1px solid var(--bs-border-color)' }}>
          <InspectorPanel
            title={`Run #${selectedRun.id}`}
            subtitle={selectedRun.status}
            onClose={() => setSelectedRun(null)}
          >
            <InspectorSection title="Details">
              <div className="small">
                <div className="mb-1"><strong>Status:</strong> <StatusBadge status={selectedRun.status === 'success' ? 'success' : 'failed'} label={selectedRun.status} size="sm" /></div>
                {selectedRun.root_path && <div className="mb-1"><strong>Root:</strong> <span className="font-monospace">{selectedRun.root_path}</span></div>}
                {selectedRun.git_from && <div className="mb-1"><strong>From:</strong> {selectedRun.git_from}</div>}
                {selectedRun.git_to && <div className="mb-1"><strong>To:</strong> {selectedRun.git_to}</div>}
                <div className="mb-1"><strong>Started:</strong> {new Date(selectedRun.started_at).toLocaleString()}</div>
                {selectedRun.finished_at && <div className="mb-1"><strong>Finished:</strong> {new Date(selectedRun.finished_at).toLocaleString()}</div>}
              </div>
            </InspectorSection>
            {summary && (
              <InspectorSection title="Summary" collapsible defaultOpen>
                <div className="small">
                  <div className="d-flex justify-content-between mb-1"><span>Symbols</span><span>{summary.counts?.symbol_occurrences ?? 0}</span></div>
                  <div className="d-flex justify-content-between mb-1"><span>Code Units</span><span>{summary.counts?.code_unit_snapshots ?? 0}</span></div>
                  <div className="d-flex justify-content-between mb-1"><span>Commits</span><span>{summary.counts?.commits ?? 0}</span></div>
                  <div className="d-flex justify-content-between mb-1"><span>Diff Files</span><span>{summary.counts?.diff_files ?? 0}</span></div>
                  <div className="d-flex justify-content-between mb-1"><span>Doc Hits</span><span>{summary.counts?.doc_hits ?? 0}</span></div>
                </div>
              </InspectorSection>
            )}
          </InspectorPanel>
        </div>
      )}
    </div>
  )
}
