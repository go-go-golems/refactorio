import type { Meta, StoryObj } from '@storybook/react'
import { SessionSelector } from './SessionSelector'
import { mockSessions } from '../../mocks/data'
import type { Session } from '../../types/api'

const meta: Meta<typeof SessionSelector> = {
  title: 'Selection/SessionSelector',
  component: SessionSelector,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 400 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof SessionSelector>

const multipleSessions: Session[] = [
  ...mockSessions,
  {
    id: 'feature-v2-a1b2',
    root_path: '/Users/dev/src/glazed',
    git_from: 'v1.0.0',
    git_to: 'v2.0.0',
    runs: { commits: 50, diff: 51, symbols: 52 },
    availability: {
      commits: true,
      diff: true,
      symbols: true,
      code_units: false,
      doc_hits: false,
      gopls_refs: false,
      tree_sitter: false,
    },
    last_updated: '2026-02-04T12:00:00Z',
  },
  {
    id: 'bugfix-fix123-c3d4',
    root_path: '/Users/dev/src/glazed',
    git_from: 'main~5',
    git_to: 'main',
    runs: { commits: 60, diff: 61 },
    availability: {
      commits: true,
      diff: true,
      symbols: false,
      code_units: false,
      doc_hits: false,
      gopls_refs: false,
      tree_sitter: false,
    },
    last_updated: '2026-02-03T16:00:00Z',
  },
]

export const Default: Story = {
  args: {
    sessions: multipleSessions,
    onSelect: () => {},
  },
}

export const WithAvailability: Story = {
  args: {
    sessions: multipleSessions,
    selected: multipleSessions[0],
    onSelect: () => {},
  },
}

export const SingleSession: Story = {
  args: {
    sessions: mockSessions,
    selected: mockSessions[0],
    onSelect: () => {},
  },
}

export const Loading: Story = {
  args: {
    sessions: [],
    loading: true,
    onSelect: () => {},
  },
}

export const Empty: Story = {
  args: {
    sessions: [],
    onSelect: () => {},
  },
}
