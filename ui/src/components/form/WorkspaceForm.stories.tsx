import type { Meta, StoryObj } from '@storybook/react'
import { WorkspaceForm } from './WorkspaceForm'
import { mockWorkspaces } from '../../mocks/data'

const meta: Meta<typeof WorkspaceForm> = {
  title: 'Form/WorkspaceForm',
  component: WorkspaceForm,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 400 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof WorkspaceForm>

export const Add: Story = {
  args: {
    onSubmit: (data) => alert(JSON.stringify(data, null, 2)),
    onCancel: () => {},
  },
}

export const Edit: Story = {
  args: {
    workspace: mockWorkspaces[0],
    onSubmit: (data) => alert(JSON.stringify(data, null, 2)),
    onCancel: () => {},
  },
}

export const Validation: Story = {
  args: {
    onSubmit: () => {},
    onCancel: () => {},
  },
  parameters: {
    docs: { description: { story: 'Try submitting the empty form to see validation errors' } },
  },
}

export const Loading: Story = {
  args: {
    workspace: mockWorkspaces[0],
    onSubmit: () => {},
    onCancel: () => {},
    loading: true,
  },
}

export const WithError: Story = {
  args: {
    workspace: mockWorkspaces[0],
    onSubmit: () => {},
    onCancel: () => {},
    error: 'Database file not found at the specified path',
  },
}
