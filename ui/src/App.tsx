import { useEffect } from 'react'
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom'
import { AppShell } from './components/layout/AppShell'
import type { SidebarSection } from './components/layout/Sidebar'
import { useGetWorkspacesQuery, useGetSessionsQuery } from './api/client'
import { useAppDispatch, useAppSelector, setActiveWorkspace, setActiveSession, selectActiveWorkspaceId, selectActiveSessionId, selectSidebarCollapsed, toggleSidebar } from './store'
import { DashboardPage } from './pages/DashboardPage'
import { RunsPage } from './pages/RunsPage'
import { SymbolsPage } from './pages/SymbolsPage'
import { CodeUnitsPage } from './pages/CodeUnitsPage'
import { CommitsPage } from './pages/CommitsPage'
import { DiffsPage } from './pages/DiffsPage'
import { DocsPage } from './pages/DocsPage'
import { FilesPage } from './pages/FilesPage'
import { SearchPage } from './pages/SearchPage'
import { WorkspacePage } from './pages/WorkspacePage'

const sidebarSections: SidebarSection[] = [
  {
    id: 'overview',
    label: 'Overview',
    items: [
      { id: 'dashboard', label: 'Dashboard', path: '/' },
      { id: 'runs', label: 'Runs', path: '/runs' },
    ],
  },
  {
    id: 'explore',
    label: 'Explore',
    items: [
      { id: 'symbols', label: 'Symbols', path: '/symbols' },
      { id: 'code-units', label: 'Code Units', path: '/code-units' },
      { id: 'commits', label: 'Commits', path: '/commits' },
      { id: 'diffs', label: 'Diffs', path: '/diffs' },
      { id: 'docs', label: 'Docs/Terms', path: '/docs' },
      { id: 'files', label: 'Files', path: '/files' },
    ],
  },
  {
    id: 'tools',
    label: 'Tools',
    items: [
      { id: 'search', label: 'Search', path: '/search' },
    ],
  },
]

function sessionOptionLabel(session: { id: string; git_from?: string; git_to?: string }) {
  const range = [session.git_from, session.git_to].filter(Boolean).join(' \u2192 ')
  if (range) return range
  const runMatch = session.id.match(/:run-(\d+)$/)
  if (runMatch) return `Session #${runMatch[1]}`
  return 'Unnamed Session'
}

export default function App() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()

  const activeWorkspaceId = useAppSelector(selectActiveWorkspaceId)
  const activeSessionId = useAppSelector(selectActiveSessionId)
  const sidebarCollapsed = useAppSelector(selectSidebarCollapsed)

  const { data: workspaces } = useGetWorkspacesQuery()
  const { data: sessions } = useGetSessionsQuery(activeWorkspaceId!, { skip: !activeWorkspaceId })

  // Auto-select first workspace if none selected
  useEffect(() => {
    if (!activeWorkspaceId && workspaces?.length) {
      dispatch(setActiveWorkspace(workspaces[0].id))
    }
  }, [activeWorkspaceId, workspaces, dispatch])

  // Auto-select first session when sessions load
  useEffect(() => {
    if (!activeSessionId && sessions?.length) {
      dispatch(setActiveSession(sessions[0].id))
    }
  }, [activeSessionId, sessions, dispatch])

  const activeWorkspace = workspaces?.find((w) => w.id === activeWorkspaceId)
  const activeSession = sessions?.find((s) => s.id === activeSessionId)
  const workspaceOptions = (workspaces ?? []).map((w) => ({ id: w.id, label: w.name }))
  const sessionOptions = (sessions ?? []).map((s) => ({
    id: s.id,
    label: sessionOptionLabel(s),
  }))

  // Derive active sidebar item from current path
  const activePath = location.pathname
  const activeItem = sidebarSections
    .flatMap((s) => s.items)
    .find((item) => {
      if (item.path === '/') return activePath === '/'
      return activePath.startsWith(item.path)
    })?.id

  // If no workspace, redirect to workspace setup
  if (workspaces && workspaces.length === 0 && location.pathname !== '/workspace') {
    return (
      <AppShell>
        <WorkspacePage />
      </AppShell>
    )
  }

  return (
    <AppShell
      sidebarProps={{
        sections: sidebarSections,
        activeItem,
        onNavigate: (path) => navigate(path),
        collapsed: sidebarCollapsed,
      }}
      topbarProps={{
        workspaceName: activeWorkspace?.name,
        sessionName: activeSession ? [activeSession.git_from, activeSession.git_to].filter(Boolean).join(' \u2192 ') : undefined,
        workspaceOptions,
        sessionOptions,
        selectedWorkspaceId: activeWorkspaceId,
        selectedSessionId: activeSessionId,
        onWorkspaceSelect: (workspaceId) => dispatch(setActiveWorkspace(workspaceId)),
        onSessionSelect: (sessionId) => dispatch(setActiveSession(sessionId)),
        onWorkspaceClick: () => navigate('/workspace'),
        onSessionClick: undefined,
        onSearch: (q) => navigate(`/search?q=${encodeURIComponent(q)}`),
      }}
      sidebarCollapsed={sidebarCollapsed}
      onSidebarToggle={() => dispatch(toggleSidebar())}
    >
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/runs" element={<RunsPage />} />
        <Route path="/symbols" element={<SymbolsPage />} />
        <Route path="/code-units" element={<CodeUnitsPage />} />
        <Route path="/commits" element={<CommitsPage />} />
        <Route path="/diffs" element={<DiffsPage />} />
        <Route path="/docs" element={<DocsPage />} />
        <Route path="/files" element={<FilesPage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/workspace" element={<WorkspacePage />} />
      </Routes>
    </AppShell>
  )
}
