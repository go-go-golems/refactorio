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
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/command.go', line: 45, col: 6, is_decl: true, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/command.go', line: 67, col: 34, is_decl: false, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/middleware.go', line: 23, col: 12, is_decl: false, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'cmd/refactorio/api.go', line: 89, col: 8, is_decl: false, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/types.go', line: 12, col: 4, is_decl: false, source: 'gopls' },
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
