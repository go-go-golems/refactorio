import type { Meta, StoryObj } from '@storybook/react'
import { http, HttpResponse, delay } from 'msw'
import { DocsPage } from './DocsPage'
import { withPageContext } from '../stories/decorators'
import { mockDocTerms, mockDocHits } from '../mocks/data'

const meta: Meta<typeof DocsPage> = {
  title: 'Pages/DocsPage',
  component: DocsPage,
  decorators: [withPageContext],
  parameters: {
    layout: 'fullscreen',
    msw: {
      handlers: [
        http.get('/api/docs/terms', () => {
          return HttpResponse.json({ items: mockDocTerms })
        }),
        http.get('/api/docs/hits', ({ request }) => {
          const url = new URL(request.url)
          const term = url.searchParams.get('term')
          const filtered = term
            ? mockDocHits.filter((hit) => hit.term === term)
            : mockDocHits
          return HttpResponse.json({ items: filtered })
        }),
      ],
    },
  },
}

export default meta
type Story = StoryObj<typeof DocsPage>

export const Default: Story = {}

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/docs/terms', () => {
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/docs/hits', () => {
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
        http.get('/api/docs/terms', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
        http.get('/api/docs/hits', async () => {
          await delay(999999)
          return HttpResponse.json({ items: [] })
        }),
      ],
    },
  },
}
