import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector, selectActiveWorkspaceId, setActiveWorkspace } from '../store'
import { useGetWorkspacesQuery, useCreateWorkspaceMutation, useUpdateWorkspaceMutation } from '../api/client'
import { WorkspaceSelector } from '../components/selection/WorkspaceSelector'
import { WorkspaceForm, type WorkspaceFormData } from '../components/form/WorkspaceForm'
import type { Workspace } from '../types/api'

export function WorkspacePage() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const activeWorkspaceId = useAppSelector(selectActiveWorkspaceId)
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Workspace | undefined>(undefined)

  const { data: workspacesData, isLoading } = useGetWorkspacesQuery()
  const [createWorkspace, { isLoading: creating }] = useCreateWorkspaceMutation()
  const [updateWorkspace, { isLoading: updating }] = useUpdateWorkspaceMutation()

  const workspaces = workspacesData ?? []
  const activeWorkspace = workspaces.find((w) => w.id === activeWorkspaceId)

  const handleSelect = (ws: Workspace) => {
    dispatch(setActiveWorkspace(ws.id))
    navigate('/')
  }

  const handleSubmit = async (data: WorkspaceFormData) => {
    try {
      if (editing) {
        const { id: _id, ...patch } = data
        await updateWorkspace({ id: editing.id, data: patch }).unwrap()
      } else {
        const result = await createWorkspace(data).unwrap()
        dispatch(setActiveWorkspace(result.id))
      }
      setShowForm(false)
      setEditing(undefined)
      if (!editing) navigate('/')
    } catch (err) {
      console.error('Workspace save failed:', err)
    }
  }

  return (
    <div className="p-4" style={{ maxWidth: 800, margin: '0 auto' }}>
      <h4 className="mb-2">Workspaces</h4>
      <p className="text-muted mb-4">Select an existing workspace or create a new one.</p>

      {isLoading ? (
        <div className="placeholder-glow">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="placeholder col-12 mb-2" style={{ height: 60 }} />
          ))}
        </div>
      ) : (
        <>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <span className="text-muted small">{workspaces.length} workspace(s)</span>
            <button type="button" className="btn btn-primary btn-sm" onClick={() => { setEditing(undefined); setShowForm(true) }}>
              New Workspace
            </button>
          </div>

          <WorkspaceSelector
            workspaces={workspaces}
            selected={activeWorkspace}
            onSelect={handleSelect}
            onEdit={(ws) => { setEditing(ws); setShowForm(true) }}
            onAdd={() => { setEditing(undefined); setShowForm(true) }}
          />

          {showForm && (
            <div className="mt-4 border-top pt-4">
              <h5 className="mb-3">{editing ? 'Edit Workspace' : 'New Workspace'}</h5>
              <WorkspaceForm
                workspace={editing}
                onSubmit={handleSubmit}
                onCancel={() => { setShowForm(false); setEditing(undefined) }}
                loading={creating || updating}
              />
            </div>
          )}
        </>
      )}
    </div>
  )
}
