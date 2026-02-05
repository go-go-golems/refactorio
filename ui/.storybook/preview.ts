import type { Preview } from '@storybook/react'
import { initialize, mswLoader } from 'msw-storybook-addon'
import 'bootstrap/dist/css/bootstrap.min.css'
import '../src/index.css'

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
  },
  loaders: [mswLoader],
}

export default preview
