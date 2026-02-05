import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { CommitsPage } from './CommitsPage'
import { withPageContext } from '../stories/decorators'
import { mockCommits, mockCommitFiles } from '../mocks/data'

const meta: Meta<typeof CommitsPage> = {
  title: 'Pages/CommitsPage',
  component: CommitsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/commits', async ({ request }) => {
          await delay(150)
          const url = new URL(request.url)
          const limit = parseInt(url.searchParams.get('limit') || '50')
          const offset = parseInt(url.searchParams.get('offset') || '0')
          const query = url.searchParams.get('q')

          let items = mockCommits
          if (query) {
            const q = query.toLowerCase()
            items = items.filter(
              (c) =>
                (c.subject ?? '').toLowerCase().includes(q) ||
                (c.author_name ?? '').toLowerCase().includes(q),
            )
          }
          const paginated = items.slice(offset, offset + limit)
          return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
        }),
        http.get('/api/commits/:hash/files', async () => {
          await delay(150)
          return HttpResponse.json({ items: mockCommitFiles })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof CommitsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/commits', async () => {
          await delay(100)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/commits/:hash/files', async () => {
          await delay(100)
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/commits', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/commits/:hash/files', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}
