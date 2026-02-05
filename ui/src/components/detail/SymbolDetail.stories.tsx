import type { Meta, StoryObj } from '@storybook/react'
import { SymbolDetail } from './SymbolDetail'
import { mockSymbols } from '../../mocks/data'
import type { SymbolRef } from '../../types/api'

const meta: Meta<typeof SymbolDetail> = {
  title: 'Detail/SymbolDetail',
  component: SymbolDetail,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 400 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof SymbolDetail>

const mockRefs: SymbolRef[] = [
  { file_path: 'pkg/handlers/command.go', start_line: 45, start_col: 6, end_line: 45, end_col: 22, is_declaration: true },
  { file_path: 'pkg/handlers/command.go', start_line: 67, start_col: 34, end_line: 67, end_col: 50, is_declaration: false },
  { file_path: 'pkg/handlers/middleware.go', start_line: 23, start_col: 12, end_line: 23, end_col: 28, is_declaration: false },
  { file_path: 'cmd/refactorio/api.go', start_line: 89, start_col: 8, end_line: 89, end_col: 24, is_declaration: false },
  { file_path: 'pkg/handlers/types.go', start_line: 12, start_col: 4, end_line: 12, end_col: 20, is_declaration: false },
]

export const Default: Story = {
  args: {
    symbol: mockSymbols[0],
    onOpenInEditor: () => {},
    onAddToPlan: () => {},
  },
}

export const WithRefs: Story = {
  args: {
    symbol: mockSymbols[0],
    refs: mockRefs,
    refsAvailable: true,
    onOpenInEditor: () => {},
  },
}

export const NoRefs: Story = {
  args: {
    symbol: mockSymbols[1],
    refsAvailable: false,
    onComputeRefs: () => alert('Computing refs...'),
  },
}

export const RefsLoading: Story = {
  args: {
    symbol: mockSymbols[0],
    refsLoading: true,
    refsAvailable: true,
  },
}

export const Exported: Story = {
  args: {
    symbol: mockSymbols[0],
    onOpenInEditor: () => {},
    onAddToPlan: () => {},
  },
}

export const Private: Story = {
  args: {
    symbol: mockSymbols[2],
    onOpenInEditor: () => {},
  },
}

export const Method: Story = {
  args: {
    symbol: mockSymbols[3],
    refs: mockRefs.slice(0, 2),
    refsAvailable: true,
    onOpenInEditor: () => {},
    onAddToPlan: () => {},
  },
}
