import type { Meta, StoryObj } from '@storybook/react'
import { WorkspaceSelector } from './WorkspaceSelector'
import { mockWorkspaces } from '../../mocks/data'

const meta: Meta<typeof WorkspaceSelector> = {
  title: 'Selection/WorkspaceSelector',
  component: WorkspaceSelector,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 400 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof WorkspaceSelector>

export const Default: Story = {
  args: {
    workspaces: mockWorkspaces,
    onSelect: () => {},
  },
}

export const Empty: Story = {
  args: {
    workspaces: [],
    onAdd: () => alert('Add workspace'),
    onSelect: () => {},
  },
}

export const Loading: Story = {
  args: {
    workspaces: [],
    loading: true,
    onSelect: () => {},
  },
}

export const WithSelected: Story = {
  args: {
    workspaces: mockWorkspaces,
    selected: mockWorkspaces[0],
    onSelect: () => {},
    onEdit: () => {},
  },
}

export const WithActions: Story = {
  args: {
    workspaces: mockWorkspaces,
    selected: mockWorkspaces[1],
    onSelect: () => {},
    onAdd: () => alert('Add workspace'),
    onEdit: (ws) => alert(`Edit: ${ws.name}`),
  },
}
