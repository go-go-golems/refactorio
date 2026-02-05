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
          return HttpResponse.json({ items: mockSessions })
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
            tables: {
              ...mockDBInfo.tables,
              symbol_occurrences: false,
              code_unit_snapshots: false,
              commits: false,
              diff_files: false,
              doc_hits: false,
            },
          })
        }),
        http.get('/api/sessions', async () => {
          await delay(100)
          return HttpResponse.json({ items: [] })
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
          return HttpResponse.json({ items: mockSessions })
        }),
        http.get('/api/runs', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
      ],
    },
  },
}
