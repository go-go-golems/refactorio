import { useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import { FilterPanel, type FilterConfig } from './FilterPanel'

const meta: Meta<typeof FilterPanel> = {
  title: 'Search/FilterPanel',
  component: FilterPanel,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 280 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof FilterPanel>

const typeFilters: FilterConfig[] = [
  {
    key: 'types',
    label: 'Entity Type',
    type: 'checkbox-group',
    options: [
      { value: 'symbol', label: 'Symbols' },
      { value: 'code_unit', label: 'Code Units' },
      { value: 'commit', label: 'Commits' },
      { value: 'diff', label: 'Diffs' },
      { value: 'doc', label: 'Doc Hits' },
      { value: 'file', label: 'Files' },
    ],
  },
]

const kindFilters: FilterConfig[] = [
  {
    key: 'kind',
    label: 'Symbol Kind',
    type: 'select',
    options: [
      { value: 'func', label: 'Function' },
      { value: 'type', label: 'Type' },
      { value: 'method', label: 'Method' },
      { value: 'const', label: 'Constant' },
      { value: 'var', label: 'Variable' },
    ],
  },
]

const combinedFilters: FilterConfig[] = [
  ...typeFilters,
  ...kindFilters,
  { key: 'path', label: 'File Path', type: 'text', placeholder: 'e.g. pkg/handlers/' },
  { key: 'date', label: 'Date Range', type: 'date-range' },
]

export const TypeFilters: Story = {
  args: {
    filters: typeFilters,
    values: { types: ['symbol', 'code_unit'] },
  },
}

export const KindFilters: Story = {
  args: {
    filters: kindFilters,
    values: { kind: 'func' },
  },
}

export const PathFilter: Story = {
  args: {
    filters: [{ key: 'path', label: 'File Path', type: 'text', placeholder: 'e.g. pkg/handlers/' }],
    values: { path: '' },
  },
}

export const Combined: Story = {
  args: {
    filters: combinedFilters,
    values: { types: [], kind: '', path: '', date: {} },
  },
}

export const Applied: Story = {
  args: {
    filters: combinedFilters,
    values: {
      types: ['symbol', 'commit'],
      kind: 'func',
      path: 'pkg/handlers',
      date: { from: '2026-02-01', to: '2026-02-05' },
    },
  },
}

export const Interactive: Story = {
  render: () => {
    const [values, setValues] = useState<Record<string, unknown>>({
      types: [],
      kind: '',
      path: '',
      date: {},
    })

    return (
      <div>
        <FilterPanel
          filters={combinedFilters}
          values={values}
          onChange={(key, value) => setValues((prev) => ({ ...prev, [key]: value }))}
          onReset={() => setValues({ types: [], kind: '', path: '', date: {} })}
        />
        <pre className="mt-3 small bg-light p-2 rounded">{JSON.stringify(values, null, 2)}</pre>
      </div>
    )
  },
}
