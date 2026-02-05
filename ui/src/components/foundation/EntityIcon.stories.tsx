import type { Meta, StoryObj } from '@storybook/react'
import { EntityIcon } from './EntityIcon'

const meta: Meta<typeof EntityIcon> = {
  title: 'Foundation/EntityIcon',
  component: EntityIcon,
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: 'select',
      options: ['symbol', 'code_unit', 'commit', 'diff', 'doc', 'file', 'folder', 'run', 'session'],
    },
    kind: {
      control: 'select',
      options: ['func', 'type', 'method', 'const', 'var'],
    },
    size: { control: 'radio', options: ['sm', 'md', 'lg'] },
  },
  parameters: {
    layout: 'centered',
  },
}

export default meta
type Story = StoryObj<typeof EntityIcon>

export const Symbol: Story = {
  args: {
    type: 'symbol',
  },
}

export const SymbolFunc: Story = {
  args: {
    type: 'symbol',
    kind: 'func',
  },
}

export const SymbolType: Story = {
  args: {
    type: 'symbol',
    kind: 'type',
  },
}

export const SymbolMethod: Story = {
  args: {
    type: 'symbol',
    kind: 'method',
  },
}

export const CodeUnit: Story = {
  args: {
    type: 'code_unit',
  },
}

export const Commit: Story = {
  args: {
    type: 'commit',
  },
}

export const Diff: Story = {
  args: {
    type: 'diff',
  },
}

export const Doc: Story = {
  args: {
    type: 'doc',
  },
}

export const File: Story = {
  args: {
    type: 'file',
  },
}

export const Folder: Story = {
  args: {
    type: 'folder',
  },
}

export const AllTypes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" />
        <span style={{ fontSize: '0.75rem' }}>symbol</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="code_unit" />
        <span style={{ fontSize: '0.75rem' }}>code_unit</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="commit" />
        <span style={{ fontSize: '0.75rem' }}>commit</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="diff" />
        <span style={{ fontSize: '0.75rem' }}>diff</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="doc" />
        <span style={{ fontSize: '0.75rem' }}>doc</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="file" />
        <span style={{ fontSize: '0.75rem' }}>file</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="folder" />
        <span style={{ fontSize: '0.75rem' }}>folder</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="run" />
        <span style={{ fontSize: '0.75rem' }}>run</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="session" />
        <span style={{ fontSize: '0.75rem' }}>session</span>
      </div>
    </div>
  ),
}

export const SymbolKinds: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" kind="func" />
        <span style={{ fontSize: '0.75rem' }}>func</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" kind="type" />
        <span style={{ fontSize: '0.75rem' }}>type</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" kind="method" />
        <span style={{ fontSize: '0.75rem' }}>method</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" kind="const" />
        <span style={{ fontSize: '0.75rem' }}>const</span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '0.25rem' }}>
        <EntityIcon type="symbol" kind="var" />
        <span style={{ fontSize: '0.75rem' }}>var</span>
      </div>
    </div>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
      <EntityIcon type="symbol" size="sm" />
      <EntityIcon type="symbol" size="md" />
      <EntityIcon type="symbol" size="lg" />
    </div>
  ),
}
