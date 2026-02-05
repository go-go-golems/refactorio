import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { CodeUnitsPage } from './CodeUnitsPage'
import { withPageContext } from '../stories/decorators'
import { mockCodeUnits } from '../mocks/data'

const meta: Meta<typeof CodeUnitsPage> = {
  title: 'Pages/CodeUnitsPage',
  component: CodeUnitsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/code-units', async ({ request }) => {
          await delay(150)
          const url = new URL(request.url)
          const limit = parseInt(url.searchParams.get('limit') || '50')
          const offset = parseInt(url.searchParams.get('offset') || '0')
          const kind = url.searchParams.get('kind')
          const query = url.searchParams.get('name')

          let items = mockCodeUnits
          if (kind) {
            items = items.filter((u) => u.kind === kind)
          }
          if (query) {
            items = items.filter((u) => u.name.toLowerCase().includes(query.toLowerCase()))
          }
          const paginated = items.slice(offset, offset + limit)
          return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
        }),
        http.get('/api/code-units/:hash', async ({ params }) => {
          await delay(150)
          const unit = mockCodeUnits.find((u) => u.unit_hash === params.hash)
          if (!unit) {
            return HttpResponse.json({ error: 'Code unit not found' }, { status: 404 })
          }
          return HttpResponse.json({
            ...unit,
            body_text: `func NewCommandProcessor(opts ...Option) CommandProcessor {
  impl := &commandProcessorImpl{
    middleware: make([]Middleware, 0),
    validators: make([]Validator, 0),
  }
  for _, opt := range opts {
    opt(impl)
  }
  return impl
}`,
            doc_text: '// NewCommandProcessor creates a new CommandProcessor with the given options.',
          })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof CodeUnitsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/code-units', async () => {
          await delay(100)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/code-units/:hash', async () => {
          await delay(100)
          return HttpResponse.json({ error: 'Code unit not found' }, { status: 404 })
        }),
      ],
    },
  },
}

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/code-units', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/code-units/:hash', async () => {
          await delay(999999)
          return HttpResponse.json({})
        }),
      ],
    },
  },
}
