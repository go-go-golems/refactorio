import type { Preview } from '@storybook/react'
import { initialize, mswLoader } from 'msw-storybook-addon'
import 'bootstrap/dist/css/bootstrap.min.css'
import '../src/index.css'
import { handlers as apiHandlers } from '../src/mocks/handlers'

// Initialize MSW
initialize()

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    layout: 'centered',
    msw: {
      handlers: apiHandlers,
    },
  },
  loaders: [mswLoader],
}

export default preview
