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
  { path: 'cmd', is_dir: true, children_count: 3 },
  { path: 'internal', is_dir: true, children_count: 4 },
  { path: 'pkg', is_dir: true, children_count: 8 },
  { path: 'go.mod', is_dir: false },
  { path: 'go.sum', is_dir: false },
  { path: 'Makefile', is_dir: false },
  { path: 'README.md', is_dir: false },
]

const childrenMap: Record<string, FileEntry[]> = {
  cmd: [
    { path: 'cmd/refactorio', is_dir: true, children_count: 2 },
    { path: 'cmd/tools', is_dir: true, children_count: 1 },
  ],
  'cmd/refactorio': [
    { path: 'cmd/refactorio/api.go', is_dir: false },
    { path: 'cmd/refactorio/root.go', is_dir: false },
    { path: 'cmd/refactorio/main.go', is_dir: false },
  ],
  'cmd/tools': [
    { path: 'cmd/tools/seed.go', is_dir: false },
  ],
  pkg: [
    { path: 'pkg/handlers', is_dir: true, children_count: 5 },
    { path: 'pkg/workbenchapi', is_dir: true, children_count: 12 },
    { path: 'pkg/refactorindex', is_dir: true, children_count: 8 },
  ],
  'pkg/handlers': [
    { path: 'pkg/handlers/command.go', is_dir: false },
    { path: 'pkg/handlers/middleware.go', is_dir: false },
    { path: 'pkg/handlers/types.go', is_dir: false },
  ],
  'pkg/workbenchapi': [
    { path: 'pkg/workbenchapi/server.go', is_dir: false },
    { path: 'pkg/workbenchapi/routes.go', is_dir: false },
    { path: 'pkg/workbenchapi/symbols.go', is_dir: false },
    { path: 'pkg/workbenchapi/diffs.go', is_dir: false },
  ],
  internal: [
    { path: 'internal/config', is_dir: true, children_count: 2 },
    { path: 'internal/util.go', is_dir: false },
  ],
  'internal/config': [
    { path: 'internal/config/config.go', is_dir: false },
    { path: 'internal/config/config_test.go', is_dir: false },
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
