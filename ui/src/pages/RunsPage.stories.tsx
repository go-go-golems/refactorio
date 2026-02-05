import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { RunsPage } from './RunsPage'
import { withPageContext } from '../stories/decorators'
import { mockRuns, mockRunSummary } from '../mocks/data'

const meta: Meta<typeof RunsPage> = {
  title: 'Pages/RunsPage',
  component: RunsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/runs', async ({ request }) => {
          await delay(150)
          const url = new URL(request.url)
          const limit = parseInt(url.searchParams.get('limit') || '50')
          const offset = parseInt(url.searchParams.get('offset') || '0')
          const items = mockRuns.slice(offset, offset + limit)
          return HttpResponse.json({ items, total: mockRuns.length, limit, offset })
        }),
        http.get('/api/runs/:id/summary', async () => {
          await delay(150)
          return HttpResponse.json(mockRunSummary)
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof RunsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/runs', async () => {
          await delay(100)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/runs/:id/summary', async () => {
          await delay(100)
          return HttpResponse.json(mockRunSummary)
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/runs', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/runs/:id/summary', async () => {
          await delay(999999)
          return HttpResponse.json(mockRunSummary)
        }),
      ],
    },
  },
}
