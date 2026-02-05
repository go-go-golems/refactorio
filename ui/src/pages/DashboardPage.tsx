import { useAppSelector } from '../store'
import { selectActiveWorkspaceId, selectActiveSessionId } from '../store'
import { useGetDBInfoQuery, useGetSessionsQuery, useGetRunsQuery } from '../api/client'
import { SessionCard } from '../components/selection/SessionCard'
import { useAppDispatch, setActiveSession } from '../store'

export function DashboardPage() {
  const dispatch = useAppDispatch()
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const activeSessionId = useAppSelector(selectActiveSessionId)

  const { data: dbInfo, isLoading: dbLoading } = useGetDBInfoQuery(workspaceId!, { skip: !workspaceId })
  const { data: sessions, isLoading: sessionsLoading } = useGetSessionsQuery(workspaceId!, { skip: !workspaceId })
  const { data: runs, isLoading: runsLoading } = useGetRunsQuery(
    { workspace_id: workspaceId!, limit: 5 },
    { skip: !workspaceId },
  )

  if (!workspaceId) {
    return <div className="p-4 text-muted">Select a workspace to get started.</div>
  }

  const hasTable = (name: string) => !!dbInfo?.tables?.[name]

  return (
    <div className="p-4">
      <h4 className="mb-4">Dashboard</h4>

      {/* DB Info summary */}
      {dbLoading ? (
        <div className="placeholder-glow mb-4">
          <span className="placeholder col-6" style={{ height: 20 }} />
        </div>
      ) : dbInfo ? (
        <div className="row mb-4">
          <div className="col-md-3">
            <div className="card">
              <div className="card-body text-center">
                <div className="fs-3 fw-bold">{hasTable('symbol_occurrences') ? 'Yes' : 'No'}</div>
                <small className="text-muted">Symbols</small>
              </div>
            </div>
          </div>
          <div className="col-md-3">
            <div className="card">
              <div className="card-body text-center">
                <div className="fs-3 fw-bold">{hasTable('code_unit_snapshots') ? 'Yes' : 'No'}</div>
                <small className="text-muted">Code Units</small>
              </div>
            </div>
          </div>
          <div className="col-md-3">
            <div className="card">
              <div className="card-body text-center">
                <div className="fs-3 fw-bold">{hasTable('commits') ? 'Yes' : 'No'}</div>
                <small className="text-muted">Commits</small>
              </div>
            </div>
          </div>
          <div className="col-md-3">
            <div className="card">
              <div className="card-body text-center">
                <div className="fs-3 fw-bold">{hasTable('diff_files') ? 'Yes' : 'No'}</div>
                <small className="text-muted">Diff Files</small>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {/* Sessions */}
      <h5 className="mb-3">Sessions</h5>
      {sessionsLoading ? (
        <div className="placeholder-glow">
          <div className="placeholder col-12 mb-2" style={{ height: 120 }} />
          <div className="placeholder col-12" style={{ height: 120 }} />
        </div>
      ) : sessions && sessions.length > 0 ? (
        <div className="row g-3 mb-4">
          {sessions.map((session) => (
            <div key={session.id} className="col-md-6 col-lg-4">
              <SessionCard
                session={session}
                selected={session.id === activeSessionId}
                onClick={() => dispatch(setActiveSession(session.id))}
              />
            </div>
          ))}
        </div>
      ) : (
        <p className="text-muted">No sessions found.</p>
      )}

      {/* Recent Runs */}
      <h5 className="mb-3">Recent Runs</h5>
      {runsLoading ? (
        <div className="placeholder-glow">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="placeholder col-12 mb-1" style={{ height: 32 }} />
          ))}
        </div>
      ) : runs && runs.length > 0 ? (
        <div className="list-group">
          {runs.map((run) => (
            <div key={run.id} className="list-group-item d-flex justify-content-between align-items-center">
              <div>
                <span className="font-monospace me-2">#{run.id}</span>
                <span className={`badge ${run.status === 'success' ? 'bg-success' : run.status === 'failed' ? 'bg-danger' : 'bg-warning'}`}>
                  {run.status}
                </span>
              </div>
              <small className="text-muted">{new Date(run.started_at).toLocaleString()}</small>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-muted">No runs found.</p>
      )}
    </div>
  )
}
