import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type { RootState } from './store'

interface UIState {
  activeWorkspaceId: string | null
  activeSessionId: string | null
  sidebarCollapsed: boolean
  inspectorOpen: boolean
  searchQuery: string
  commandPaletteOpen: boolean
}

const initialState: UIState = {
  activeWorkspaceId: null,
  activeSessionId: null,
  sidebarCollapsed: false,
  inspectorOpen: false,
  searchQuery: '',
  commandPaletteOpen: false,
}

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setActiveWorkspace(state, action: PayloadAction<string | null>) {
      state.activeWorkspaceId = action.payload
      // Reset session when workspace changes
      state.activeSessionId = null
    },
    setActiveSession(state, action: PayloadAction<string | null>) {
      state.activeSessionId = action.payload
    },
    toggleSidebar(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed
    },
    toggleInspector(state) {
      state.inspectorOpen = !state.inspectorOpen
    },
    openInspector(state) {
      state.inspectorOpen = true
    },
    closeInspector(state) {
      state.inspectorOpen = false
    },
    setSearchQuery(state, action: PayloadAction<string>) {
      state.searchQuery = action.payload
    },
    openCommandPalette(state) {
      state.commandPaletteOpen = true
    },
    closeCommandPalette(state) {
      state.commandPaletteOpen = false
    },
  },
})

export const {
  setActiveWorkspace,
  setActiveSession,
  toggleSidebar,
  toggleInspector,
  openInspector,
  closeInspector,
  setSearchQuery,
  openCommandPalette,
  closeCommandPalette,
} = uiSlice.actions

export const uiReducer = uiSlice.reducer

// Selectors
export const selectActiveWorkspaceId = (state: RootState) => state.ui.activeWorkspaceId
export const selectActiveSessionId = (state: RootState) => state.ui.activeSessionId
export const selectSidebarCollapsed = (state: RootState) => state.ui.sidebarCollapsed
export const selectInspectorOpen = (state: RootState) => state.ui.inspectorOpen
export const selectSearchQuery = (state: RootState) => state.ui.searchQuery
export const selectCommandPaletteOpen = (state: RootState) => state.ui.commandPaletteOpen
