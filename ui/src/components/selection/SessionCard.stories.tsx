import type { Meta, StoryObj } from '@storybook/react'
import { SessionCard } from './SessionCard'
import { mockSessions } from '../../mocks/data'
import type { Session } from '../../types/api'

const meta: Meta<typeof SessionCard> = {
  title: 'Selection/SessionCard',
  component: SessionCard,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  decorators: [
    (Story) => (
      <div style={{ width: '420px' }}>
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof SessionCard>

export const Default: Story = {
  args: {
    session: mockSessions[0],
  },
}

export const Selected: Story = {
  args: {
    session: mockSessions[0],
    selected: true,
    onClick: () => {},
  },
}

export const WithEdit: Story = {
  args: {
    session: mockSessions[0],
    onClick: () => alert('Card clicked'),
    onEdit: () => alert('Edit clicked'),
  },
}

export const AllAvailable: Story = {
  args: {
    session: {
      ...mockSessions[0],
      availability: {
        commits: true,
        diff: true,
        symbols: true,
        code_units: true,
        doc_hits: true,
        gopls_refs: true,
        tree_sitter: true,
      },
    } as Session,
  },
}

export const MostlyMissing: Story = {
  args: {
    session: {
      ...mockSessions[0],
      availability: {
        commits: true,
        diff: false,
        symbols: false,
        code_units: false,
        doc_hits: false,
        gopls_refs: false,
        tree_sitter: false,
      },
    } as Session,
  },
}

export const MultipleCards: Story = {
  render: () => (
    <div className="d-flex flex-column gap-3">
      <SessionCard session={mockSessions[0]} selected />
      <SessionCard
        session={{
          ...mockSessions[0],
          id: 'session-2',
          git_from: 'v1.0.0',
          git_to: 'v2.0.0',
          availability: {
            commits: true,
            diff: true,
            symbols: true,
            code_units: false,
            doc_hits: false,
            gopls_refs: false,
            tree_sitter: false,
          },
        }}
      />
    </div>
  ),
}
