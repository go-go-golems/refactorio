import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { DashboardPage } from './DashboardPage'
import { withPageContext } from '../stories/decorators'
import { mockDBInfo, mockRuns, mockSessions } from '../mocks/data'

const meta: Meta<typeof DashboardPage> = {
  title: 'Pages/DashboardPage',
  component: DashboardPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/db/info', async () => {
          await delay(150)
          return HttpResponse.json(mockDBInfo)
        }),
        http.get('/api/sessions', async () => {
          await delay(150)
          return HttpResponse.json({ sessions: mockSessions })
        }),
        http.get('/api/runs', async ({ request }) => {
          await delay(150)
          const url = new URL(request.url)
          const limit = parseInt(url.searchParams.get('limit') || '50')
          const offset = parseInt(url.searchParams.get('offset') || '0')
          const items = mockRuns.slice(offset, offset + limit)
          return HttpResponse.json({ items, total: mockRuns.length, limit, offset })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof DashboardPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/db/info', async () => {
          await delay(100)
          return HttpResponse.json({
            ...mockDBInfo,
            row_counts: {
              symbols: 0,
              code_unit_snapshots: 0,
              commits: 0,
              diff_files: 0,
              doc_hits: 0,
            },
          })
        }),
        http.get('/api/sessions', async () => {
          await delay(100)
          return HttpResponse.json({ sessions: [] })
        }),
        http.get('/api/runs', async () => {
          await delay(100)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/db/info', async () => {
          await delay(999999)
          return HttpResponse.json(mockDBInfo)
        }),
        http.get('/api/sessions', async () => {
          await delay(999999)
          return HttpResponse.json({ sessions: mockSessions })
        }),
        http.get('/api/runs', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
      ],
    },
  },
}
