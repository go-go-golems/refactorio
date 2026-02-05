import type { Meta, StoryObj } from '@storybook/react'
import { CopyButton } from './CopyButton'

const meta: Meta<typeof CopyButton> = {
  title: 'Foundation/CopyButton',
  component: CopyButton,
  tags: ['autodocs'],
  argTypes: {
    size: { control: 'radio', options: ['sm', 'md'] },
    variant: { control: 'radio', options: ['icon', 'text', 'outline'] },
  },
  parameters: {
    layout: 'centered',
  },
}

export default meta
type Story = StoryObj<typeof CopyButton>

export const Default: Story = {
  args: {
    text: 'pkg/handlers/command.go:45',
  },
}

export const WithLabel: Story = {
  args: {
    text: 'pkg/handlers/command.go:45',
    label: 'Copy path',
  },
}

export const SmallSize: Story = {
  args: {
    text: 'a7b3c9f2',
    label: 'Copy hash',
    size: 'sm',
  },
}

export const OutlineVariant: Story = {
  args: {
    text: 'a7b3c9f2',
    label: 'Copy',
    variant: 'outline',
  },
}

export const TextVariant: Story = {
  args: {
    text: 'github.com/example/pkg/handlers',
    label: 'Copy package',
    variant: 'text',
  },
}

export const AllVariants: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
      <CopyButton text="icon" variant="icon" />
      <CopyButton text="text" variant="text" label="Text" />
      <CopyButton text="outline" variant="outline" label="Outline" />
    </div>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
      <CopyButton text="small" size="sm" label="Small" variant="outline" />
      <CopyButton text="medium" size="md" label="Medium" variant="outline" />
    </div>
  ),
}
