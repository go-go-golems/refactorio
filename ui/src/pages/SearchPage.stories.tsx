import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { SearchPage } from './SearchPage'
import { withPageContext } from '../stories/decorators'
import { mockSearchResults } from '../mocks/data'

const meta: Meta<typeof SearchPage> = {
  title: 'Pages/SearchPage',
  component: SearchPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.post('/api/search', async ({ request }) => {
          const body = (await request.json()) as { query?: string }
          const query = body.query || ''
          const filtered = query
            ? mockSearchResults.filter(
                (r) =>
                  r.primary.toLowerCase().includes(query.toLowerCase()) ||
                  r.snippet?.toLowerCase().includes(query.toLowerCase()),
              )
            : mockSearchResults
          return HttpResponse.json({ items: filtered })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof SearchPage>

export const Default: Story = {
  parameters: {
    initialPath: '/search?q=CommandProcessor',
  },
}

export const Empty: Story = {
  parameters: {
    initialPath: '/search?q=nonexistent_query_xyz',
    msw: {
      handlers: [
        http.post('/api/search', () => {
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    initialPath: '/search?q=CommandProcessor',
    msw: {
      handlers: [
        http.post('/api/search', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}
