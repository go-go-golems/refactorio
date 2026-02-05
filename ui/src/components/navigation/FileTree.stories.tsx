import type { Meta, StoryObj } from '@storybook/react'
import { FileTree } from './FileTree'
import type { FileEntry } from '../../types/api'

const meta: Meta<typeof FileTree> = {
  title: 'Navigation/FileTree',
  component: FileTree,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof FileTree>

const rootEntries: FileEntry[] = [
  { path: 'cmd', kind: 'dir' },
  { path: 'internal', kind: 'dir' },
  { path: 'pkg', kind: 'dir' },
  { path: 'go.mod', kind: 'file' },
  { path: 'go.sum', kind: 'file' },
  { path: 'Makefile', kind: 'file' },
  { path: 'README.md', kind: 'file' },
]

const childrenMap: Record<string, FileEntry[]> = {
  cmd: [
    { path: 'cmd/refactorio', kind: 'dir' },
    { path: 'cmd/tools', kind: 'dir' },
  ],
  'cmd/refactorio': [
    { path: 'cmd/refactorio/api.go', kind: 'file' },
    { path: 'cmd/refactorio/root.go', kind: 'file' },
    { path: 'cmd/refactorio/main.go', kind: 'file' },
  ],
  'cmd/tools': [
    { path: 'cmd/tools/seed.go', kind: 'file' },
  ],
  pkg: [
    { path: 'pkg/handlers', kind: 'dir' },
    { path: 'pkg/workbenchapi', kind: 'dir' },
    { path: 'pkg/refactorindex', kind: 'dir' },
  ],
  'pkg/handlers': [
    { path: 'pkg/handlers/command.go', kind: 'file' },
    { path: 'pkg/handlers/middleware.go', kind: 'file' },
    { path: 'pkg/handlers/types.go', kind: 'file' },
  ],
  'pkg/workbenchapi': [
    { path: 'pkg/workbenchapi/server.go', kind: 'file' },
    { path: 'pkg/workbenchapi/routes.go', kind: 'file' },
    { path: 'pkg/workbenchapi/symbols.go', kind: 'file' },
    { path: 'pkg/workbenchapi/diffs.go', kind: 'file' },
  ],
  internal: [
    { path: 'internal/config', kind: 'dir' },
    { path: 'internal/util.go', kind: 'file' },
  ],
  'internal/config': [
    { path: 'internal/config/config.go', kind: 'file' },
    { path: 'internal/config/config_test.go', kind: 'file' },
  ],
}

export const Default: Story = {
  args: {
    entries: rootEntries,
    childrenMap,
  },
}

export const WithBadges: Story = {
  args: {
    entries: rootEntries,
    childrenMap,
    badges: {
      'pkg/handlers/command.go': 3,
      'pkg/handlers/middleware.go': 1,
      pkg: 4,
      'pkg/handlers': 4,
    },
  },
}

export const DeepNesting: Story = {
  args: {
    entries: rootEntries,
    childrenMap,
    expandedPaths: new Set(['cmd', 'cmd/refactorio', 'pkg', 'pkg/handlers']),
  },
}

export const Selected: Story = {
  args: {
    entries: rootEntries,
    childrenMap,
    expandedPaths: new Set(['pkg', 'pkg/handlers']),
    selectedPath: 'pkg/handlers/command.go',
  },
}

export const Loading: Story = {
  args: {
    entries: [],
    loading: true,
  },
}

export const Empty: Story = {
  args: {
    entries: [],
  },
}
