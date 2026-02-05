import type { Workspace } from '../../types/api'

export interface WorkspaceSelectorProps {
  /** Available workspaces */
  workspaces: Workspace[]
  /** Currently selected workspace */
  selected?: Workspace
  /** Called when a workspace is selected */
  onSelect: (workspace: Workspace) => void
  /** Called when "Add workspace" is clicked */
  onAdd?: () => void
  /** Called when edit is clicked for a workspace */
  onEdit?: (workspace: Workspace) => void
  /** Whether the list is loading */
  loading?: boolean
}

function SkeletonList() {
  return (
    <div className="placeholder-glow">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="list-group-item py-2">
          <span className="placeholder col-5 mb-1 d-block" style={{ height: 14 }} />
          <span className="placeholder col-8" style={{ height: 12 }} />
        </div>
      ))}
    </div>
  )
}

export function WorkspaceSelector({
  workspaces,
  selected,
  onSelect,
  onAdd,
  onEdit,
  loading = false,
}: WorkspaceSelectorProps) {
  if (loading) {
    return (
      <div className="workspace-selector">
        <div className="fw-semibold small mb-2">Select Workspace</div>
        <SkeletonList />
      </div>
    )
  }

  if (workspaces.length === 0) {
    return (
      <div className="workspace-selector text-center p-4">
        <p className="text-muted mb-2">No workspaces configured</p>
        {onAdd && (
          <button type="button" className="btn btn-primary btn-sm" onClick={onAdd}>
            + Add Workspace
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="workspace-selector">
      <div className="d-flex justify-content-between align-items-center mb-2">
        <span className="fw-semibold small">Select Workspace</span>
        {onAdd && (
          <button type="button" className="btn btn-outline-primary btn-sm" onClick={onAdd}>
            + Add
          </button>
        )}
      </div>

      <div className="list-group">
        {workspaces.map((ws) => {
          const isSelected = selected?.id === ws.id
          return (
            <button
              key={ws.id}
              type="button"
              className={`list-group-item list-group-item-action d-flex justify-content-between align-items-start ${
                isSelected ? 'active' : ''
              }`}
              onClick={() => onSelect(ws)}
            >
              <div className="me-2">
                <div className={`fw-medium ${isSelected ? '' : ''}`}>{ws.name}</div>
                <small className={isSelected ? 'text-white-50' : 'text-muted'}>
                  {ws.db_path}
                </small>
                {ws.repo_root && (
                  <small className={`d-block ${isSelected ? 'text-white-50' : 'text-muted'}`}>
                    {ws.repo_root}
                  </small>
                )}
              </div>
              {onEdit && (
                <button
                  type="button"
                  className={`btn btn-sm ${isSelected ? 'btn-outline-light' : 'btn-outline-secondary'}`}
                  onClick={(e) => {
                    e.stopPropagation()
                    onEdit(ws)
                  }}
                  aria-label={`Edit ${ws.name}`}
                >
                  Edit
                </button>
              )}
            </button>
          )
        })}
      </div>
    </div>
  )
}
