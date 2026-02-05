import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { WorkspacePage } from './WorkspacePage'
import { withPageContext } from '../stories/decorators'
import { mockWorkspaces } from '../mocks/data'

const meta: Meta<typeof WorkspacePage> = {
  title: 'Pages/WorkspacePage',
  component: WorkspacePage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/workspaces', () => {
          return HttpResponse.json({ items: mockWorkspaces })
        }),
        http.post('/api/workspaces', async ({ request }) => {
          const body = (await request.json()) as Record<string, unknown>
          return HttpResponse.json(
            {
              id: 'new-workspace',
              name: body.name || 'new-workspace',
              db_path: body.db_path || '/tmp/new.db',
              repo_root: body.repo_root || '/tmp/repo',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
            { status: 201 },
          )
        }),
        http.patch('/api/workspaces/:id', async ({ request, params }) => {
          const body = (await request.json()) as Record<string, unknown>
          const existing = mockWorkspaces.find((w) => w.id === params.id)
          return HttpResponse.json({
            ...existing,
            ...body,
            id: params.id,
            updated_at: new Date().toISOString(),
          })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof WorkspacePage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/workspaces', () => {
          return HttpResponse.json({ items: [] })
        }),
        http.post('/api/workspaces', async ({ request }) => {
          const body = (await request.json()) as Record<string, unknown>
          return HttpResponse.json(
            {
              id: 'new-workspace',
              name: body.name || 'new-workspace',
              db_path: body.db_path || '/tmp/new.db',
              repo_root: body.repo_root || '/tmp/repo',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
            { status: 201 },
          )
        }),
        http.patch('/api/workspaces/:id', async ({ request, params }) => {
          const body = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            ...body,
            id: params.id,
            updated_at: new Date().toISOString(),
          })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/workspaces', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
        http.post('/api/workspaces', async () => {
          await delay(999999)
          return HttpResponse.json({}, { status: 201 })
        }),
        http.patch('/api/workspaces/:id', async () => {
          await delay(999999)
          return HttpResponse.json({})
        }),
      ],
    },
  },
}
