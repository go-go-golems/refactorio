import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { GlobalSearchBar } from './GlobalSearchBar'

const meta: Meta<typeof GlobalSearchBar> = {
  title: 'Search/GlobalSearchBar',
  component: GlobalSearchBar,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof GlobalSearchBar>

export const Default: Story = {
  args: {},
}

export const WithValue: Story = {
  args: {
    value: 'CommandProcessor',
  },
}

export const WithSuggestions: Story = {
  render: () => {
    const [value, setValue] = useState('Command')
    return (
      <GlobalSearchBar
        value={value}
        onChange={setValue}
        onSubmit={(v) => alert(`Search: ${v}`)}
        suggestions={[
          'CommandProcessor',
          'CommandHandler',
          'CommandRegistry',
          'NewCommandProcessor',
          'commandProcessorImpl',
        ]}
      />
    )
  },
}

export const Loading: Story = {
  args: {
    value: 'searching...',
    loading: true,
  },
}

export const Focused: Story = {
  args: {
    autoFocus: true,
    placeholder: 'Type to search (Ctrl+K)',
  },
}

export const Interactive: Story = {
  render: () => {
    const [value, setValue] = useState('')
    const [submitted, setSubmitted] = useState('')
    return (
      <div>
        <GlobalSearchBar
          value={value}
          onChange={setValue}
          onSubmit={setSubmitted}
          suggestions={[
            'CommandProcessor',
            'NewCommandProcessor',
            'Process',
            'middleware',
            'handler',
          ]}
        />
        {submitted && (
          <p className="mt-2 mb-0 text-muted small">
            Submitted: <strong>{submitted}</strong>
          </p>
        )}
      </div>
    )
  },
}
