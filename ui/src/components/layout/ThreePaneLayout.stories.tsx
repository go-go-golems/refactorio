import type { Meta, StoryObj } from '@storybook/react'
import { ThreePaneLayout, ThreePaneLayoutWithToggle } from './ThreePaneLayout'

const meta: Meta<typeof ThreePaneLayout> = {
  title: 'Layout/ThreePaneLayout',
  component: ThreePaneLayout,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  decorators: [
    (Story) => (
      <div style={{ height: '500px' }}>
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof ThreePaneLayout>

const LeftPane = () => (
  <div className="p-3">
    <h6 className="text-muted mb-3">FILTERS</h6>
    <div className="mb-2">
      <label className="form-label small">Type</label>
      <div className="form-check">
        <input className="form-check-input" type="checkbox" defaultChecked />
        <label className="form-check-label">Symbols</label>
      </div>
      <div className="form-check">
        <input className="form-check-input" type="checkbox" defaultChecked />
        <label className="form-check-label">Code Units</label>
      </div>
      <div className="form-check">
        <input className="form-check-input" type="checkbox" />
        <label className="form-check-label">Commits</label>
      </div>
    </div>
  </div>
)

const CenterPane = () => (
  <div className="p-3">
    <h5>Results</h5>
    <div className="list-group">
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="list-group-item list-group-item-action">
          <div className="d-flex justify-content-between">
            <strong>CommandProcessor</strong>
            <small className="text-muted">type</small>
          </div>
          <small className="text-muted">pkg/handlers/command.go:45</small>
        </div>
      ))}
    </div>
  </div>
)

const RightPane = () => (
  <div className="p-3">
    <h6 className="text-muted mb-3">INSPECTOR</h6>
    <div className="mb-3">
      <strong>CommandProcessor</strong>
      <span className="badge bg-primary ms-2">type</span>
    </div>
    <div className="mb-2">
      <small className="text-muted d-block">Package</small>
      <code className="small">github.com/example/pkg/handlers</code>
    </div>
    <div className="mb-2">
      <small className="text-muted d-block">File</small>
      <code className="small">pkg/handlers/command.go:45</code>
    </div>
    <div className="mt-3">
      <button className="btn btn-primary btn-sm me-2">Open Detail</button>
      <button className="btn btn-outline-secondary btn-sm">Copy Hash</button>
    </div>
  </div>
)

export const Default: Story = {
  args: {
    left: <LeftPane />,
    center: <CenterPane />,
    right: <RightPane />,
  },
}

export const TwoPanes: Story = {
  args: {
    left: <LeftPane />,
    center: <CenterPane />,
    showRight: false,
  },
}

export const CustomWidths: Story = {
  args: {
    left: <LeftPane />,
    center: <CenterPane />,
    right: <RightPane />,
    leftWidth: 200,
    rightWidth: 400,
  },
}

export const WithToggle: Story = {
  render: () => (
    <ThreePaneLayoutWithToggle
      left={<LeftPane />}
      center={<CenterPane />}
      right={<RightPane />}
    />
  ),
}
