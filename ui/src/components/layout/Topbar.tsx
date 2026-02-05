export interface TopbarProps {
  /** Current workspace name */
  workspaceName?: string
  /** Current session name */
  sessionName?: string
  /** Workspace options for selector */
  workspaceOptions?: Array<{ id: string; label: string }>
  /** Session options for selector */
  sessionOptions?: Array<{ id: string; label: string }>
  /** Selected workspace ID */
  selectedWorkspaceId?: string | null
  /** Selected session ID */
  selectedSessionId?: string | null
  /** Called when workspace selection changes */
  onWorkspaceSelect?: (workspaceId: string | null) => void
  /** Called when session selection changes */
  onSessionSelect?: (sessionId: string | null) => void
  /** Called when workspace selector is clicked */
  onWorkspaceClick?: () => void
  /** Called when session selector is clicked */
  onSessionClick?: () => void
  /** Called when search is submitted */
  onSearch?: (query: string) => void
  /** Called when command palette is triggered */
  onCommandPalette?: () => void
  /** Custom class name */
  className?: string
}

export function Topbar({
  workspaceName,
  sessionName,
  workspaceOptions,
  sessionOptions,
  selectedWorkspaceId,
  selectedSessionId,
  onWorkspaceSelect,
  onSessionSelect,
  onWorkspaceClick,
  onSessionClick,
  onSearch,
  onCommandPalette,
  className = '',
}: TopbarProps) {
  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      onSearch?.(e.currentTarget.value)
    }
  }

  return (
    <header className={`topbar navbar navbar-expand bg-body border-bottom px-3 ${className}`}>
      <div className="container-fluid p-0">
        {/* Brand */}
        <span className="navbar-brand d-flex align-items-center gap-2 mb-0">
          <span style={{ fontSize: '1.25rem' }}>🔧</span>
          <span className="fw-semibold">Refactorio Workbench</span>
        </span>

        {/* Selectors */}
        <div className="d-flex align-items-center gap-2 mx-3">
          {workspaceOptions && onWorkspaceSelect ? (
            <select
              className="form-select form-select-sm"
              style={{ minWidth: 180 }}
              value={selectedWorkspaceId ?? ''}
              onChange={(e) => onWorkspaceSelect(e.target.value || null)}
              aria-label="Workspace"
            >
              <option value="" disabled>No workspace</option>
              {workspaceOptions.map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </select>
          ) : (
            <button
              type="button"
              className="btn btn-outline-secondary btn-sm d-flex align-items-center gap-1"
              onClick={onWorkspaceClick}
              disabled={!onWorkspaceClick}
            >
              {workspaceName || 'No workspace'}
              <span className="opacity-50">▾</span>
            </button>
          )}

          {sessionOptions && onSessionSelect ? (
            <select
              className="form-select form-select-sm"
              style={{ minWidth: 220 }}
              value={selectedSessionId ?? ''}
              onChange={(e) => onSessionSelect(e.target.value || null)}
              disabled={!selectedWorkspaceId || sessionOptions.length === 0}
              aria-label="Session"
            >
              <option value="">No session</option>
              {sessionOptions.map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </select>
          ) : (
            <button
              type="button"
              className="btn btn-outline-secondary btn-sm d-flex align-items-center gap-1"
              onClick={onSessionClick}
              disabled={!onSessionClick || !workspaceName}
            >
              {sessionName || 'No session'}
              <span className="opacity-50">▾</span>
            </button>
          )}
        </div>

        {/* Search */}
        <div className="flex-grow-1 mx-3" style={{ maxWidth: '500px' }}>
          <div className="input-group input-group-sm">
            <span className="input-group-text bg-body border-end-0">
              🔍
            </span>
            <input
              type="search"
              className="form-control border-start-0"
              placeholder="Search symbols, files, diffs..."
              onKeyDown={handleSearchKeyDown}
              aria-label="Search"
            />
          </div>
        </div>

        {/* Command Palette */}
        <button
          type="button"
          className="btn btn-outline-secondary btn-sm d-flex align-items-center gap-1"
          onClick={onCommandPalette}
          title="Command Palette (Ctrl+K)"
        >
          <span className="d-none d-md-inline">⌘K</span>
        </button>
      </div>

      <style>{`
        .topbar {
          height: 56px;
        }
      `}</style>
    </header>
  )
}
