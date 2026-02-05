import { create } from 'zustand'
import type { Workspace, DBInfo, Session } from '../types/api'
import * as api from '../api/client'

interface WorkspaceState {
  // Data
  workspaces: Workspace[]
  activeWorkspace: Workspace | null
  dbInfo: DBInfo | null
  sessions: Session[]
  activeSession: Session | null

  // Loading states
  loadingWorkspaces: boolean
  loadingDbInfo: boolean
  loadingSessions: boolean

  // Error
  error: string | null

  // Actions
  fetchWorkspaces: () => Promise<void>
  selectWorkspace: (workspace: Workspace) => Promise<void>
  selectSession: (session: Session) => void
  clearError: () => void
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  workspaces: [],
  activeWorkspace: null,
  dbInfo: null,
  sessions: [],
  activeSession: null,
  loadingWorkspaces: false,
  loadingDbInfo: false,
  loadingSessions: false,
  error: null,

  fetchWorkspaces: async () => {
    set({ loadingWorkspaces: true, error: null })
    try {
      const config = await api.workspaces.list()
      set({ workspaces: config.workspaces, loadingWorkspaces: false })
    } catch (err) {
      set({ error: (err as Error).message, loadingWorkspaces: false })
    }
  },

  selectWorkspace: async (workspace: Workspace) => {
    set({
      activeWorkspace: workspace,
      dbInfo: null,
      sessions: [],
      activeSession: null,
      loadingDbInfo: true,
      loadingSessions: true,
      error: null,
    })

    // Fetch DB info and sessions in parallel
    try {
      const [dbInfo, sessions] = await Promise.all([
        api.db.info(workspace.id),
        api.sessions.list(workspace.id),
      ])
      set({
        dbInfo,
        sessions,
        loadingDbInfo: false,
        loadingSessions: false,
        // Auto-select first session if available
        activeSession: sessions.length > 0 ? sessions[0] : null,
      })
    } catch (err) {
      set({
        error: (err as Error).message,
        loadingDbInfo: false,
        loadingSessions: false,
      })
    }
  },

  selectSession: (session: Session) => {
    set({ activeSession: session })
  },

  clearError: () => set({ error: null }),
}))
