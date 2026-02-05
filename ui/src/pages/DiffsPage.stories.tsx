import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { DiffsPage } from './DiffsPage'
import { withPageContext } from '../stories/decorators'
import { mockDiffFiles, mockDiffHunks } from '../mocks/data'

const meta: Meta<typeof DiffsPage> = {
  title: 'Pages/DiffsPage',
  component: DiffsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/diff-runs', () => {
          return HttpResponse.json({
            items: [
              {
                run_id: 43,
                root_path: '/Users/dev/src/glazed',
                git_from: 'HEAD~20',
                git_to: 'HEAD',
                files_count: 3,
              },
            ],
          })
        }),
        http.get('/api/diff/:runId/files', () => {
          return HttpResponse.json({ items: mockDiffFiles })
        }),
        http.get('/api/diff/:runId/file', () => {
          return HttpResponse.json({ hunks: mockDiffHunks })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof DiffsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/diff-runs', () => {
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/diff/:runId/files', () => {
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/diff/:runId/file', () => {
          return HttpResponse.json({ hunks: [] })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/diff-runs', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/diff/:runId/files', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/diff/:runId/file', async () => {
          await delay(999999)
          return HttpResponse.json({ hunks: [] })
        }),
      ],
    },
  },
}
