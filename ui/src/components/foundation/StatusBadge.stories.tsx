import type { Meta, StoryObj } from '@storybook/react'
import { StatusBadge } from './StatusBadge'

const meta: Meta<typeof StatusBadge> = {
  title: 'Foundation/StatusBadge',
  component: StatusBadge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: ['success', 'running', 'failed', 'pending', 'warning'],
    },
    size: { control: 'radio', options: ['sm', 'md'] },
  },
  parameters: {
    layout: 'centered',
  },
}

export default meta
type Story = StoryObj<typeof StatusBadge>

export const Success: Story = {
  args: {
    status: 'success',
  },
}

export const Running: Story = {
  args: {
    status: 'running',
  },
}

export const Failed: Story = {
  args: {
    status: 'failed',
  },
}

export const Pending: Story = {
  args: {
    status: 'pending',
  },
}

export const Warning: Story = {
  args: {
    status: 'warning',
  },
}

export const WithCustomLabel: Story = {
  args: {
    status: 'success',
    label: '1,847 commits',
  },
}

export const AllStatuses: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
      <StatusBadge status="success" />
      <StatusBadge status="running" />
      <StatusBadge status="failed" />
      <StatusBadge status="pending" />
      <StatusBadge status="warning" />
    </div>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
      <StatusBadge status="success" size="sm" />
      <StatusBadge status="success" size="md" />
    </div>
  ),
}

export const NoPulse: Story = {
  args: {
    status: 'running',
    pulse: false,
  },
}
