import { useState, useCallback } from 'react'
import type { Workspace } from '../../types/api'

export interface WorkspaceFormData {
  id: string
  name: string
  db_path: string
  repo_root: string
}

export interface WorkspaceFormProps {
  /** Existing workspace for edit mode */
  workspace?: Workspace
  /** Called on form submit */
  onSubmit: (data: WorkspaceFormData) => void
  /** Called on cancel */
  onCancel?: () => void
  /** Whether submission is in progress */
  loading?: boolean
  /** Error message from server */
  error?: string
}

function validate(data: WorkspaceFormData): Record<string, string> {
  const errors: Record<string, string> = {}
  if (!data.id.trim()) errors.id = 'ID is required'
  if (!data.name.trim()) errors.name = 'Name is required'
  if (!data.db_path.trim()) errors.db_path = 'Database path is required'
  if (data.db_path && !data.db_path.endsWith('.db') && !data.db_path.endsWith('.sqlite')) {
    errors.db_path = 'Path should end with .db or .sqlite'
  }
  return errors
}

export function WorkspaceForm({
  workspace,
  onSubmit,
  onCancel,
  loading = false,
  error,
}: WorkspaceFormProps) {
  const [data, setData] = useState<WorkspaceFormData>({
    id: workspace?.id ?? '',
    name: workspace?.name ?? '',
    db_path: workspace?.db_path ?? '',
    repo_root: workspace?.repo_root ?? '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [touched, setTouched] = useState<Set<string>>(new Set())

  const handleChange = useCallback((field: keyof WorkspaceFormData, value: string) => {
    setData((prev) => ({ ...prev, [field]: value }))
    setTouched((prev) => new Set([...prev, field]))
  }, [])

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      const validationErrors = validate(data)
      setErrors(validationErrors)
      setTouched(new Set(['id', 'name', 'db_path', 'repo_root']))
      if (Object.keys(validationErrors).length === 0) {
        onSubmit(data)
      }
    },
    [data, onSubmit],
  )

  const isEdit = !!workspace

  return (
    <form onSubmit={handleSubmit} noValidate>
      <h6 className="mb-3">{isEdit ? 'Edit Workspace' : 'Add Workspace'}</h6>

      {error && (
        <div className="alert alert-danger small py-2" role="alert">
          {error}
        </div>
      )}

      <div className="mb-3">
        <label htmlFor="ws-id" className="form-label small">
          ID <span className="text-danger">*</span>
        </label>
        <input
          type="text"
          className={`form-control form-control-sm font-monospace ${touched.has('id') && errors.id ? 'is-invalid' : ''}`}
          id="ws-id"
          value={data.id}
          onChange={(e) => handleChange('id', e.target.value)}
          placeholder="e.g. glazed"
          disabled={loading || isEdit}
        />
        {touched.has('id') && errors.id && (
          <div className="invalid-feedback">{errors.id}</div>
        )}
        <div className="form-text">Stable workspace identifier (cannot be changed later).</div>
      </div>

      <div className="mb-3">
        <label htmlFor="ws-name" className="form-label small">
          Name <span className="text-danger">*</span>
        </label>
        <input
          type="text"
          className={`form-control form-control-sm ${touched.has('name') && errors.name ? 'is-invalid' : ''}`}
          id="ws-name"
          value={data.name}
          onChange={(e) => handleChange('name', e.target.value)}
          placeholder="e.g. glazed"
          disabled={loading}
        />
        {touched.has('name') && errors.name && (
          <div className="invalid-feedback">{errors.name}</div>
        )}
      </div>

      <div className="mb-3">
        <label htmlFor="ws-db-path" className="form-label small">
          Database Path <span className="text-danger">*</span>
        </label>
        <input
          type="text"
          className={`form-control form-control-sm font-monospace ${touched.has('db_path') && errors.db_path ? 'is-invalid' : ''}`}
          id="ws-db-path"
          value={data.db_path}
          onChange={(e) => handleChange('db_path', e.target.value)}
          placeholder="/path/to/index.sqlite"
          disabled={loading}
        />
        {touched.has('db_path') && errors.db_path && (
          <div className="invalid-feedback">{errors.db_path}</div>
        )}
        <div className="form-text">Path to the refactorindex SQLite database</div>
      </div>

      <div className="mb-3">
        <label htmlFor="ws-repo-root" className="form-label small">
          Repository Root
        </label>
        <input
          type="text"
          className="form-control form-control-sm font-monospace"
          id="ws-repo-root"
          value={data.repo_root}
          onChange={(e) => handleChange('repo_root', e.target.value)}
          placeholder="/path/to/repo (optional)"
          disabled={loading}
        />
        <div className="form-text">Optional. Needed for file content and git operations.</div>
      </div>

      <div className="d-flex gap-2">
        <button type="submit" className="btn btn-primary btn-sm" disabled={loading}>
          {loading && <span className="spinner-border spinner-border-sm me-1" role="status" />}
          {isEdit ? 'Save Changes' : 'Add Workspace'}
        </button>
        {onCancel && (
          <button type="button" className="btn btn-outline-secondary btn-sm" onClick={onCancel} disabled={loading}>
            Cancel
          </button>
        )}
      </div>
    </form>
  )
}
