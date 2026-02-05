import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { Sidebar, createNavItems } from './Sidebar'

const meta: Meta<typeof Sidebar> = {
  title: 'Layout/Sidebar',
  component: Sidebar,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  decorators: [
    (Story) => (
      <div style={{ height: '600px', display: 'flex' }}>
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof Sidebar>

const defaultSections = createNavItems()

export const Default: Story = {
  args: {
    sections: defaultSections,
    activeItem: 'dashboard',
    onNavigate: () => {},
  },
}

export const WithActiveItem: Story = {
  args: {
    sections: defaultSections,
    activeItem: 'symbols',
    onNavigate: () => {},
  },
}

export const Collapsed: Story = {
  args: {
    sections: defaultSections,
    activeItem: 'dashboard',
    collapsed: true,
    onNavigate: () => {},
  },
}

export const WithBadges: Story = {
  args: {
    sections: [
      {
        id: 'explore',
        label: 'Explore',
        collapsible: true,
        defaultOpen: true,
        items: [
          { id: 'symbols', label: 'Symbols', path: '/symbols', badge: '12,456' },
          { id: 'commits', label: 'Commits', path: '/commits', badge: '1,847' },
          { id: 'diffs', label: 'Diffs', path: '/diffs', badge: 142 },
        ],
      },
    ],
    activeItem: 'symbols',
    onNavigate: () => {},
  },
}

export const Interactive: Story = {
  render: function InteractiveSidebar() {
    const [activeItem, setActiveItem] = useState('dashboard')
    const [collapsed, setCollapsed] = useState(false)

    return (
      <div style={{ display: 'flex', height: '100%' }}>
        <Sidebar
          sections={defaultSections}
          activeItem={activeItem}
          collapsed={collapsed}
          onNavigate={(path) => {
            const item = defaultSections
              .flatMap((s) => s.items)
              .find((i) => i.path === path)
            if (item) setActiveItem(item.id)
          }}
        />
        <div className="p-3">
          <button
            className="btn btn-outline-secondary btn-sm"
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? 'Expand' : 'Collapse'}
          </button>
          <p className="mt-2 text-muted">Active: {activeItem}</p>
        </div>
      </div>
    )
  },
}
