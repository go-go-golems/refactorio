import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { Pagination } from './Pagination'

const meta: Meta<typeof Pagination> = {
  title: 'Foundation/Pagination',
  component: Pagination,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
}

export default meta
type Story = StoryObj<typeof Pagination>

export const Default: Story = {
  args: {
    total: 1234,
    limit: 50,
    offset: 0,
    onChange: () => {},
  },
}

export const MiddlePage: Story = {
  args: {
    total: 1234,
    limit: 50,
    offset: 300,
    onChange: () => {},
  },
}

export const LastPage: Story = {
  args: {
    total: 1234,
    limit: 50,
    offset: 1200,
    onChange: () => {},
  },
}

export const Compact: Story = {
  args: {
    total: 1234,
    limit: 50,
    offset: 100,
    compact: true,
    onChange: () => {},
  },
}

export const NoTotal: Story = {
  args: {
    limit: 50,
    offset: 100,
    showTotal: false,
    onChange: () => {},
  },
}

export const FewPages: Story = {
  args: {
    total: 150,
    limit: 50,
    offset: 0,
    onChange: () => {},
  },
}

export const Interactive: Story = {
  render: function InteractivePagination() {
    const [offset, setOffset] = useState(0)
    return (
      <div>
        <Pagination
          total={500}
          limit={25}
          offset={offset}
          onChange={setOffset}
        />
        <p className="mt-3 text-muted small">Current offset: {offset}</p>
      </div>
    )
  },
}

export const InteractiveCompact: Story = {
  render: function InteractiveCompactPagination() {
    const [offset, setOffset] = useState(0)
    return (
      <div>
        <Pagination
          total={500}
          limit={25}
          offset={offset}
          onChange={setOffset}
          compact
        />
        <p className="mt-3 text-muted small">Current offset: {offset}</p>
      </div>
    )
  },
}
