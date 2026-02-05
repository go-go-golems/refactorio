import type { Meta, StoryObj } from '@storybook/react'
import { Breadcrumb } from './Breadcrumb'

const meta: Meta<typeof Breadcrumb> = {
  title: 'Navigation/Breadcrumb',
  component: Breadcrumb,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Breadcrumb>

export const Default: Story = {
  args: {
    items: [
      { label: 'pkg', path: 'pkg' },
      { label: 'handlers', path: 'pkg/handlers' },
      { label: 'command.go', path: 'pkg/handlers/command.go' },
    ],
  },
}

export const Clickable: Story = {
  args: {
    items: [
      { label: 'Root', path: '/' },
      { label: 'pkg', path: 'pkg' },
      { label: 'workbenchapi', path: 'pkg/workbenchapi' },
      { label: 'server.go' },
    ],
    onNavigate: (path: string) => alert(`Navigate to: ${path}`),
  },
}

export const Truncated: Story = {
  args: {
    items: [
      { label: 'Root', path: '/' },
      { label: 'github.com', path: 'github.com' },
      { label: 'go-go-golems', path: 'github.com/go-go-golems' },
      { label: 'glazed', path: 'github.com/go-go-golems/glazed' },
      { label: 'pkg', path: 'github.com/go-go-golems/glazed/pkg' },
      { label: 'handlers', path: 'github.com/go-go-golems/glazed/pkg/handlers' },
      { label: 'command.go' },
    ],
    maxItems: 4,
  },
}

export const SingleItem: Story = {
  args: {
    items: [{ label: 'Root' }],
  },
}
