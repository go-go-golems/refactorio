import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetDiffRunsQuery, useGetDiffFilesQuery, useGetDiffFileQuery } from '../api/client'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { DiffViewer } from '../components/code-display/DiffViewer'
import { StatusBadge } from '../components/foundation/StatusBadge'
import type { DiffRun, DiffFile } from '../types/api'

const runColumns: Column<DiffRun>[] = [
  { key: 'run_id', header: 'Run', width: '60px', render: (r) => <span className="font-monospace">#{r.run_id}</span> },
  { key: 'git_range', header: 'Range', render: (r) => <span className="small">{r.git_from} → {r.git_to}</span> },
  { key: 'files', header: 'Files', width: '60px', render: (r) => <span className="small">{r.files_count}</span> },
]

const fileColumns: Column<DiffFile>[] = [
  {
    key: 'status', header: '', width: '32px',
    render: (f) => <StatusBadge status={f.status === 'A' ? 'success' : f.status === 'D' ? 'failed' : 'warning'} label={f.status} size="sm" />,
  },
  { key: 'path', header: 'File', render: (f) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 250 }}>{f.file_path}</span> },
  {
    key: 'changes', header: '+/-', width: '80px',
    render: (f) => <span className="small"><span className="text-success">+{f.additions}</span> <span className="text-danger">-{f.deletions}</span></span>,
  },
]

export function DiffsPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [selectedRun, setSelectedRun] = useState<DiffRun | null>(null)
  const [selectedFile, setSelectedFile] = useState<DiffFile | null>(null)

  const { data: runs, isLoading: runsLoading } = useGetDiffRunsQuery(
    { workspace_id: workspaceId! },
    { skip: !workspaceId },
  )

  const { data: files, isLoading: filesLoading } = useGetDiffFilesQuery(
    { run_id: selectedRun?.run_id ?? 0, workspace_id: workspaceId! },
    { skip: !selectedRun || !workspaceId },
  )

  const { data: hunks, isFetching: hunksLoading } = useGetDiffFileQuery(
    { run_id: selectedRun?.run_id ?? 0, workspace_id: workspaceId!, file_path: selectedFile?.file_path ?? '' },
    { skip: !selectedRun || !selectedFile || !workspaceId },
  )

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      {/* Left: Run list */}
      <div style={{ width: 280, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
        <h6 className="mb-2">Diff Runs</h6>
        <EntityTable
          columns={runColumns}
          data={runs ?? []}
          loading={runsLoading}
          selectedId={selectedRun ? String(selectedRun.run_id) : undefined}
          onSelect={(r) => { setSelectedRun(r); setSelectedFile(null) }}
          getItemId={(r) => String(r.run_id)}
          emptyMessage="No diff runs"
        />
      </div>

      {/* Middle: File list */}
      {selectedRun && (
        <div style={{ width: 320, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
          <h6 className="mb-2">Files in Run #{selectedRun.run_id}</h6>
          <EntityTable
            columns={fileColumns}
            data={files ?? []}
            loading={filesLoading}
            selectedId={selectedFile?.file_path}
            onSelect={setSelectedFile}
            getItemId={(f) => f.file_path}
            emptyMessage="No files"
          />
        </div>
      )}

      {/* Right: Diff viewer */}
      <div className="flex-grow-1 p-3 overflow-auto">
        {selectedFile ? (
          hunksLoading ? (
            <div className="text-muted p-4">Loading diff...</div>
          ) : hunks && hunks.length > 0 ? (
            <>
              <div className="d-flex justify-content-between align-items-center mb-2">
                <span className="font-monospace small">{selectedFile.file_path}</span>
                <button type="button" className="btn-close btn-sm" onClick={() => setSelectedFile(null)} />
              </div>
              <DiffViewer hunks={hunks} />
            </>
          ) : (
            <div className="text-muted p-4">No diff data</div>
          )
        ) : (
          <div className="text-muted p-4 text-center">Select a file to view its diff</div>
        )}
      </div>
    </div>
  )
}
