import { create } from 'zustand'

interface UIState {
  // Sidebar
  sidebarCollapsed: boolean
  toggleSidebar: () => void

  // Inspector panel
  inspectorOpen: boolean
  toggleInspector: () => void
  openInspector: () => void
  closeInspector: () => void

  // Search
  searchQuery: string
  setSearchQuery: (query: string) => void

  // Command palette
  commandPaletteOpen: boolean
  openCommandPalette: () => void
  closeCommandPalette: () => void
}

export const useUIStore = create<UIState>((set) => ({
  sidebarCollapsed: false,
  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),

  inspectorOpen: false,
  toggleInspector: () => set((s) => ({ inspectorOpen: !s.inspectorOpen })),
  openInspector: () => set({ inspectorOpen: true }),
  closeInspector: () => set({ inspectorOpen: false }),

  searchQuery: '',
  setSearchQuery: (query: string) => set({ searchQuery: query }),

  commandPaletteOpen: false,
  openCommandPalette: () => set({ commandPaletteOpen: true }),
  closeCommandPalette: () => set({ commandPaletteOpen: false }),
}))
