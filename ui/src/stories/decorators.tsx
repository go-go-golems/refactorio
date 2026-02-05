import type { Decorator } from '@storybook/react'
import { Provider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import { api } from '../api/baseApi'
import { uiReducer } from '../store/uiSlice'

// Import all slice files to register endpoints before store creation
import '../api/workspaces'
import '../api/runs'
import '../api/sessions'
import '../api/symbols'
import '../api/codeUnits'
import '../api/commits'
import '../api/diffs'
import '../api/docs'
import '../api/files'
import '../api/search'

function createMockStore() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return configureStore({
    reducer: {
      [api.reducerPath]: api.reducer,
      ui: uiReducer,
    } as Record<string, any>, // eslint-disable-line @typescript-eslint/no-explicit-any
    middleware: (getDefaultMiddleware: any) => // eslint-disable-line @typescript-eslint/no-explicit-any
      getDefaultMiddleware().concat(api.middleware),
    preloadedState: {
      ui: {
        activeWorkspaceId: 'glazed',
        activeSessionId: 'main-head20-head-a7b3',
        sidebarCollapsed: false,
        inspectorOpen: false,
        searchQuery: '',
        commandPaletteOpen: false,
      },
    } as Record<string, unknown>,
  })
}

export const withPageContext: Decorator = (Story, context) => {
  const store = createMockStore()
  const initialPath = (context.parameters?.initialPath as string) || '/'

  return (
    <Provider store={store}>
      <MemoryRouter initialEntries={[initialPath]}>
        <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
          <Story />
        </div>
      </MemoryRouter>
    </Provider>
  )
}
