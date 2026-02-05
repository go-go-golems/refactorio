import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { FilesPage } from './FilesPage'
import { withPageContext } from '../stories/decorators'
import { mockFileTree } from '../mocks/data'

const meta: Meta<typeof FilesPage> = {
  title: 'Pages/FilesPage',
  component: FilesPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/files', () => {
          return HttpResponse.json({ items: mockFileTree })
        }),
        http.get('/api/file', ({ request }) => {
          const url = new URL(request.url)
          const path = url.searchParams.get('path') || 'unknown'
          return HttpResponse.json({
            path,
            content:
              'package handlers\n\nimport (\n\t"context"\n\t"fmt"\n)\n\ntype CommandProcessor interface {\n\tProcess(ctx context.Context, cmd Command) (Result, error)\n\tValidate(cmd Command) error\n}\n\nfunc NewCommandProcessor(opts ...Option) CommandProcessor {\n\timpl := &commandProcessorImpl{\n\t\tmiddleware: make([]Middleware, 0),\n\t\tvalidators: make([]Validator, 0),\n\t}\n\tfor _, opt := range opts {\n\t\topt(impl)\n\t}\n\treturn impl\n}\n',
          })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof FilesPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/files', () => {
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/file', () => {
          return HttpResponse.json({ path: '', content: '' })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/files', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/file', async () => {
          await delay(999999)
          return HttpResponse.json({ path: '', content: '' })
        }),
      ],
    },
  },
}
