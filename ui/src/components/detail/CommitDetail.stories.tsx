import type { Meta, StoryObj } from '@storybook/react'
import { CommitDetail } from './CommitDetail'
import { mockCommits, mockDiffFiles } from '../../mocks/data'
import type { CommitFile } from '../../types/api'

const meta: Meta<typeof CommitDetail> = {
  title: 'Detail/CommitDetail',
  component: CommitDetail,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 500 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof CommitDetail>

const commitFiles: CommitFile[] = [
  { file_path: 'pkg/handlers/command.go', status: 'M', additions: 45, deletions: 23 },
  { file_path: 'pkg/handlers/middleware.go', status: 'M', additions: 12, deletions: 5 },
  { file_path: 'pkg/handlers/types.go', status: 'A', additions: 89, deletions: 0 },
  { file_path: 'pkg/handlers/old_types.go', status: 'D', additions: 0, deletions: 34 },
  { file_path: 'pkg/handlers/command_test.go', status: 'M', additions: 22, deletions: 8 },
]

export const Default: Story = {
  args: {
    commit: mockCommits[0],
    onViewDiff: () => {},
  },
}

export const WithFiles: Story = {
  args: {
    commit: mockCommits[0],
    files: commitFiles,
    onFileClick: (f) => alert(`Click: ${f.file_path}`),
    onViewDiff: () => {},
  },
}

export const LongMessage: Story = {
  args: {
    commit: {
      ...mockCommits[0],
      body: `This change renames the CommandProcessor interface to Handler for
consistency with the rest of the codebase.

Key changes:
- Renamed CommandProcessor -> Handler
- Updated all references in pkg/handlers/
- Updated middleware registration to use new type
- Added migration guide in docs/

Breaking changes:
- Anyone implementing CommandProcessor must now implement Handler
- The Process() method signature remains the same

Fixes #1234`,
    },
    files: commitFiles,
    onFileClick: () => {},
    onViewDiff: () => {},
  },
}

export const SimpleCommit: Story = {
  args: {
    commit: mockCommits[1],
  },
}
