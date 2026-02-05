import type { Meta, StoryObj } from '@storybook/react'
import { Topbar } from './Topbar'

const meta: Meta<typeof Topbar> = {
  title: 'Layout/Topbar',
  component: Topbar,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}

export default meta
type Story = StoryObj<typeof Topbar>

export const Default: Story = {
  args: {
    workspaceName: 'glazed',
    sessionName: 'main: HEAD~20→HEAD',
    onWorkspaceClick: () => alert('Workspace clicked'),
    onSessionClick: () => alert('Session clicked'),
    onSearch: (q) => alert(`Search: ${q}`),
    onCommandPalette: () => alert('Command palette'),
  },
}

export const NoWorkspace: Story = {
  args: {
    onWorkspaceClick: () => alert('Workspace clicked'),
  },
}

export const WorkspaceOnly: Story = {
  args: {
    workspaceName: 'glazed',
    onWorkspaceClick: () => alert('Workspace clicked'),
    onSessionClick: () => alert('Session clicked'),
  },
}

export const Loading: Story = {
  args: {
    workspaceName: 'glazed',
    sessionName: 'Loading...',
  },
}
