import { useEffect, useState } from 'react'
import { useGetDiffRunsQuery, useGetDiffFilesQuery, useGetDiffFileQuery } from '../api/client'
import { useSessionContext } from '../hooks/useSessionContext'
import { EntityTable, type Column } from '../components/data-display/EntityTable'
import { DiffViewer } from '../components/code-display/DiffViewer'
import { StatusBadge } from '../components/foundation/StatusBadge'
import type { DiffRun, DiffFile } from '../types/api'

const runColumns: Column<DiffRun>[] = [
  { key: 'id', header: 'Run', width: '60px', render: (r) => <span className="font-monospace">#{r.id}</span> },
  { key: 'git_range', header: 'Range', render: (r) => <span className="small">{[r.git_from, r.git_to].filter(Boolean).join(' → ')}</span> },
]

const fileColumns: Column<DiffFile>[] = [
  {
    key: 'status', header: '', width: '32px',
    render: (f) => <StatusBadge status={f.status === 'A' ? 'success' : f.status === 'D' ? 'failed' : 'warning'} label={f.status} size="sm" />,
  },
  { key: 'path', header: 'File', render: (f) => <span className="font-monospace small text-truncate d-block" style={{ maxWidth: 250 }}>{f.path}</span> },
]

export function DiffsPage() {
  const { workspaceId, sessionId, activeSession } = useSessionContext()
  const [selectedRun, setSelectedRun] = useState<DiffRun | null>(null)
  const [selectedFile, setSelectedFile] = useState<DiffFile | null>(null)
  const diffAvailable = Boolean(activeSession?.runs.diff)

  const { data: runs, isLoading: runsLoading } = useGetDiffRunsQuery(
    { workspace_id: workspaceId!, session_id: sessionId ?? undefined },
    { skip: !workspaceId || !sessionId },
  )
  const runRows = diffAvailable ? (runs ?? []) : []

  const { data: files, isLoading: filesLoading } = useGetDiffFilesQuery(
    { run_id: selectedRun?.id ?? 0, workspace_id: workspaceId! },
    { skip: !selectedRun || !workspaceId },
  )

  const { data: hunks, isFetching: hunksLoading } = useGetDiffFileQuery(
    { run_id: selectedRun?.id ?? 0, workspace_id: workspaceId!, path: selectedFile?.path ?? '' },
    { skip: !selectedRun || !selectedFile || !workspaceId },
  )

  useEffect(() => {
    setSelectedRun(null)
    setSelectedFile(null)
  }, [sessionId])

  useEffect(() => {
    if (!selectedRun && runRows.length > 0) {
      setSelectedRun(runRows[0])
    }
  }, [runRows, selectedRun])

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      {/* Left: Run list */}
      <div style={{ width: 280, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
        <h6 className="mb-2">Diff Runs</h6>
        <EntityTable
          columns={runColumns}
          data={runRows}
          loading={runsLoading && diffAvailable}
          selectedId={selectedRun ? String(selectedRun.id) : undefined}
          onSelect={(r) => { setSelectedRun(r); setSelectedFile(null) }}
          getItemId={(r) => String(r.id)}
          emptyMessage={diffAvailable ? 'No diff runs' : 'No diff data for this session'}
        />
      </div>

      {/* Middle: File list */}
      {selectedRun && (
        <div style={{ width: 320, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
          <h6 className="mb-2">Files in Run #{selectedRun.id}</h6>
          <EntityTable
            columns={fileColumns}
            data={files ?? []}
            loading={filesLoading}
            selectedId={selectedFile?.path}
            onSelect={setSelectedFile}
            getItemId={(f) => f.path}
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
                <span className="font-monospace small">{selectedFile.path}</span>
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
