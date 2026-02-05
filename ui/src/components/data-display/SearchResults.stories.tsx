import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { SearchResults } from './SearchResults'
import { mockSearchResults } from '../../mocks/data'
import type { SearchResult } from '../../types/api'

const meta: Meta<typeof SearchResults> = {
  title: 'DataDisplay/SearchResults',
  component: SearchResults,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
  },
  decorators: [
    (Story) => (
      <div style={{ maxWidth: '600px' }}>
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof SearchResults>

export const Default: Story = {
  args: {
    results: mockSearchResults,
    query: 'CommandProcessor',
  },
}

export const Grouped: Story = {
  args: {
    results: mockSearchResults,
    groupByType: true,
    query: 'Command',
  },
}

export const Flat: Story = {
  args: {
    results: mockSearchResults,
    groupByType: false,
    query: 'Command',
  },
}

export const WithSelection: Story = {
  args: {
    results: mockSearchResults,
    selectedId: 'a7b3c9f2',
    query: 'CommandProcessor',
  },
}

export const Loading: Story = {
  args: {
    results: [],
    loading: true,
  },
}

export const Empty: Story = {
  args: {
    results: [],
    query: 'nonexistent',
  },
}

export const SingleType: Story = {
  args: {
    results: mockSearchResults.filter((r) => r.type === 'symbol'),
    groupByType: true,
    query: 'Command',
  },
}

export const ManyResults: Story = {
  args: {
    results: [
      ...mockSearchResults,
      ...mockSearchResults.map((r, i) => ({ ...r, id: `${r.id}-2-${i}`, label: `${r.label} (copy)` })),
      ...mockSearchResults.map((r, i) => ({ ...r, id: `${r.id}-3-${i}`, label: `${r.label} (another)` })),
    ],
    groupByType: true,
    query: 'Command',
  },
}

export const Interactive: Story = {
  render: function InteractiveResults() {
    const [selectedId, setSelectedId] = useState<string>()
    const [selected, setSelected] = useState<SearchResult>()

    return (
      <div>
        <SearchResults
          results={mockSearchResults}
          query="Command"
          selectedId={selectedId}
          onSelect={(r) => {
            setSelectedId(r.id)
            setSelected(r)
          }}
        />
        {selected && (
          <div className="mt-3 p-3 bg-light rounded">
            <h6>Selected:</h6>
            <pre className="small mb-0">{JSON.stringify(selected, null, 2)}</pre>
          </div>
        )}
      </div>
    )
  },
}
