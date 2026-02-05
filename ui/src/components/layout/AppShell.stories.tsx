import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { AppShell } from './AppShell'
import { createNavItems } from './Sidebar'

const meta: Meta<typeof AppShell> = {
  title: 'Layout/AppShell',
  component: AppShell,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
}

export default meta
type Story = StoryObj<typeof AppShell>

const defaultSections = createNavItems()

export const Default: Story = {
  args: {
    topbarProps: {
      workspaceName: 'glazed',
      sessionName: 'main: HEAD~20→HEAD',
    },
    sidebarProps: {
      sections: defaultSections,
      activeItem: 'dashboard',
      onNavigate: () => {},
    },
    children: (
      <div className="p-4">
        <h1>Dashboard</h1>
        <p className="text-muted">Welcome to Refactorio Workbench</p>
      </div>
    ),
  },
}

export const NoWorkspace: Story = {
  args: {
    topbarProps: {},
    sidebarProps: {
      sections: defaultSections,
      activeItem: 'dashboard',
      onNavigate: () => {},
    },
    children: (
      <div className="p-4 text-center mt-5">
        <h2>No Workspace Selected</h2>
        <p className="text-muted">Select or create a workspace to get started</p>
        <button className="btn btn-primary">Add Workspace</button>
      </div>
    ),
  },
}

export const CollapsedSidebar: Story = {
  args: {
    topbarProps: {
      workspaceName: 'glazed',
      sessionName: 'main: HEAD~20→HEAD',
    },
    sidebarProps: {
      sections: defaultSections,
      activeItem: 'symbols',
      onNavigate: () => {},
    },
    sidebarCollapsed: true,
    children: (
      <div className="p-4">
        <h1>Symbols</h1>
        <p className="text-muted">Browse symbol definitions</p>
      </div>
    ),
  },
}

export const Interactive: Story = {
  render: function InteractiveAppShell() {
    const [activeItem, setActiveItem] = useState('dashboard')
    const [collapsed, setCollapsed] = useState(false)

    return (
      <AppShell
        topbarProps={{
          workspaceName: 'glazed',
          sessionName: 'main: HEAD~20→HEAD',
          onWorkspaceClick: () => alert('Workspace'),
          onSessionClick: () => alert('Session'),
          onSearch: (q) => alert(`Search: ${q}`),
          onCommandPalette: () => alert('Command palette'),
        }}
        sidebarProps={{
          sections: defaultSections,
          activeItem,
          onNavigate: (path) => {
            const item = defaultSections.flatMap((s) => s.items).find((i) => i.path === path)
            if (item) setActiveItem(item.id)
          },
        }}
        sidebarCollapsed={collapsed}
        onSidebarToggle={() => setCollapsed(!collapsed)}
      >
        <div className="p-4">
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h1 className="mb-0">{activeItem.charAt(0).toUpperCase() + activeItem.slice(1)}</h1>
            <button
              className="btn btn-outline-secondary btn-sm"
              onClick={() => setCollapsed(!collapsed)}
            >
              {collapsed ? 'Expand Sidebar' : 'Collapse Sidebar'}
            </button>
          </div>
          <p className="text-muted">Content for {activeItem}</p>
        </div>
      </AppShell>
    )
  },
}
