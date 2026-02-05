import type { Meta, StoryObj } from '@storybook/react'
import { InspectorPanel, InspectorSection } from './InspectorPanel'
import { CopyButton, StatusBadge, EntityIcon } from '../foundation'

const meta: Meta<typeof InspectorPanel> = {
  title: 'Detail/InspectorPanel',
  component: InspectorPanel,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  decorators: [
    (Story) => (
      <div style={{ height: '500px', width: '320px', border: '1px solid #dee2e6' }}>
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof InspectorPanel>

export const SymbolInspector: Story = {
  args: {
    title: 'CommandProcessor',
    subtitle: 'pkg/handlers/command.go:45',
    actions: (
      <>
        <button className="btn btn-primary btn-sm">Open Detail</button>
        <button className="btn btn-outline-secondary btn-sm">Add to Plan</button>
      </>
    ),
    children: (
      <>
        <InspectorSection title="Info">
          <div className="mb-2">
            <small className="text-muted d-block">Kind</small>
            <span className="d-flex align-items-center gap-1">
              <EntityIcon type="symbol" kind="type" size="sm" />
              <span>type</span>
              <StatusBadge status="success" label="exported" size="sm" />
            </span>
          </div>
          <div className="mb-2">
            <small className="text-muted d-block">Package</small>
            <code className="small">github.com/go-go-golems/glazed/pkg/handlers</code>
          </div>
          <div className="mb-2">
            <small className="text-muted d-block">Hash</small>
            <span className="d-flex align-items-center gap-1">
              <code className="small">a7b3c9f2</code>
              <CopyButton text="a7b3c9f2" size="sm" />
            </span>
          </div>
        </InspectorSection>
        <InspectorSection title="Signature">
          <pre className="bg-body-tertiary p-2 rounded small mb-0" style={{ whiteSpace: 'pre-wrap' }}>
            {'type CommandProcessor interface {\n  Process(ctx context.Context, cmd Command) (Result, error)\n  Validate(cmd Command) error\n}'}
          </pre>
        </InspectorSection>
        <InspectorSection title="References" collapsible defaultOpen={false}>
          <p className="text-muted small mb-0">Gopls refs not computed</p>
          <button className="btn btn-outline-primary btn-sm mt-1">Compute References</button>
        </InspectorSection>
      </>
    ),
  },
}

export const CommitInspector: Story = {
  args: {
    title: 'Rename CommandProcessor to Handler',
    subtitle: 'abc1234 by Alice',
    actions: (
      <>
        <button className="btn btn-outline-secondary btn-sm">View Diff</button>
        <CopyButton text="abc1234" label="Copy Hash" size="sm" variant="outline" />
      </>
    ),
    children: (
      <>
        <InspectorSection title="Details">
          <div className="mb-2">
            <small className="text-muted d-block">Author</small>
            <span>Alice &lt;alice@example.com&gt;</span>
          </div>
          <div className="mb-2">
            <small className="text-muted d-block">Date</small>
            <span>Feb 5, 2026 6:00 AM</span>
          </div>
          <div className="mb-2">
            <small className="text-muted d-block">Files Changed</small>
            <span>14 files (+245 / -89)</span>
          </div>
        </InspectorSection>
        <InspectorSection title="Body">
          <p className="small">This change renames the CommandProcessor interface to Handler for consistency with the rest of the codebase.</p>
        </InspectorSection>
      </>
    ),
  },
}

export const Loading: Story = {
  args: {
    title: 'Loading...',
    loading: true,
    children: null,
  },
}

export const WithClose: Story = {
  args: {
    title: 'CommandProcessor',
    subtitle: 'type',
    onClose: () => alert('Close'),
    children: <p className="text-muted">Panel content here</p>,
  },
}

export const EmptyState: Story = {
  args: {
    title: 'No Selection',
    children: (
      <div className="text-center text-muted py-4">
        <p className="mb-1">No item selected</p>
        <small>Click on an item to view details</small>
      </div>
    ),
  },
}
