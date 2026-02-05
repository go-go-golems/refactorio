import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { TabNav } from './TabNav'

const meta: Meta<typeof TabNav> = {
  title: 'Navigation/TabNav',
  component: TabNav,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof TabNav>

export const Default: Story = {
  args: {
    tabs: [
      { id: 'overview', label: 'Overview' },
      { id: 'refs', label: 'References' },
      { id: 'history', label: 'History' },
      { id: 'audit', label: 'Audit' },
    ],
    activeTab: 'overview',
  },
}

export const WithBadges: Story = {
  args: {
    tabs: [
      { id: 'overview', label: 'Overview' },
      { id: 'refs', label: 'References', badge: 23 },
      { id: 'history', label: 'History', badge: 5 },
      { id: 'audit', label: 'Audit', badge: 0 },
    ],
    activeTab: 'refs',
  },
}

export const Disabled: Story = {
  args: {
    tabs: [
      { id: 'overview', label: 'Overview' },
      { id: 'refs', label: 'References', disabled: true },
      { id: 'history', label: 'History' },
      { id: 'audit', label: 'Audit', disabled: true },
    ],
    activeTab: 'overview',
  },
}

export const Interactive: Story = {
  render: () => {
    const [active, setActive] = useState('overview')
    return (
      <div>
        <TabNav
          tabs={[
            { id: 'overview', label: 'Overview' },
            { id: 'snapshot', label: 'Snapshot' },
            { id: 'history', label: 'History', badge: 3 },
            { id: 'diffs', label: 'Diffs', badge: 2 },
            { id: 'related', label: 'Related' },
          ]}
          activeTab={active}
          onChange={setActive}
        />
        <div className="p-3 border border-top-0">
          <p className="mb-0">Active tab: <strong>{active}</strong></p>
        </div>
      </div>
    )
  },
}

export const ManyTabs: Story = {
  args: {
    tabs: [
      { id: 'overview', label: 'Overview' },
      { id: 'definition', label: 'Definition' },
      { id: 'refs', label: 'References', badge: 45 },
      { id: 'callers', label: 'Callers', badge: 12 },
      { id: 'callees', label: 'Callees', badge: 8 },
      { id: 'history', label: 'History', badge: 5 },
      { id: 'audit', label: 'Audit' },
      { id: 'related', label: 'Related' },
    ],
    activeTab: 'refs',
  },
}
