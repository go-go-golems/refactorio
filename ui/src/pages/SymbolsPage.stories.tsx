import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { SymbolsPage } from './SymbolsPage'
import { withPageContext } from '../stories/decorators'
import { mockSymbols, mockSymbolRefs } from '../mocks/data'

const meta: Meta<typeof SymbolsPage> = {
  title: 'Pages/SymbolsPage',
  component: SymbolsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/symbols', async ({ request }) => {
          await delay(150)
          const url = new URL(request.url)
          const limit = parseInt(url.searchParams.get('limit') || '50')
          const offset = parseInt(url.searchParams.get('offset') || '0')
          const query = url.searchParams.get('q')
          const kind = url.searchParams.get('kind')

          let items = mockSymbols
          if (query) {
            items = items.filter(
              (s) =>
                s.name.toLowerCase().includes(query.toLowerCase()) ||
                s.package_path.toLowerCase().includes(query.toLowerCase()),
            )
          }
          if (kind) {
            items = items.filter((s) => s.kind === kind)
          }
          const paginated = items.slice(offset, offset + limit)
          return HttpResponse.json({ items: paginated, total: items.length, limit, offset })
        }),
        http.get('/api/symbols/:hash/refs', async () => {
          await delay(150)
          return HttpResponse.json({ items: mockSymbolRefs })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof SymbolsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/symbols', async () => {
          await delay(100)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/symbols/:hash/refs', async () => {
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
        http.get('/api/symbols', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [], total: 0, limit: 50, offset: 0 })
        }),
        http.get('/api/symbols/:hash/refs', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}
