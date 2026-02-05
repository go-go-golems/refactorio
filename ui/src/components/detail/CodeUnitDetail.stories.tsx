import type { Meta, StoryObj } from '@storybook/react'
import { CodeUnitDetail } from './CodeUnitDetail'
import { mockCodeUnits } from '../../mocks/data'
import type { CodeUnitDetail as CodeUnitDetailType, CodeUnit } from '../../types/api'

const meta: Meta<typeof CodeUnitDetail> = {
  title: 'Detail/CodeUnitDetail',
  component: CodeUnitDetail,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ maxWidth: 500 }}><Story /></div>],
}

export default meta
type Story = StoryObj<typeof CodeUnitDetail>

const funcUnit: CodeUnitDetailType = {
  ...mockCodeUnits[0],
  body_text: `func NewCommandProcessor(opts ...Option) CommandProcessor {
\tp := &commandProcessorImpl{
\t\tregistry: NewRegistry(),
\t\tlogger:   slog.Default(),
\t}
\tfor _, opt := range opts {
\t\topt(p)
\t}
\treturn p
}`,
  doc_text: '// NewCommandProcessor creates a new CommandProcessor with the given options.\n// It initializes a default registry and logger.',
}

const methodUnit: CodeUnitDetailType = {
  ...mockCodeUnits[1],
  body_text: `func (p *commandProcessorImpl) Process(ctx context.Context, cmd Command) (Result, error) {
\tif err := p.validate(cmd); err != nil {
\t\treturn Result{}, fmt.Errorf("validation: %w", err)
\t}
\thandler, ok := p.registry.Get(cmd.Name())
\tif !ok {
\t\treturn Result{}, ErrUnknownCommand
\t}
\treturn handler.Execute(ctx, cmd)
}`,
}

const typeUnit: CodeUnitDetailType = {
  unit_hash: 'cu_type1',
  run_id: 44,
  kind: 'type',
  name: 'CommandProcessor',
  pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
  file: 'pkg/handlers/command.go',
  start_line: 45,
  start_col: 1,
  end_line: 52,
  end_col: 2,
  body_hash: 'type_body_hash_1',
  body_text: `type CommandProcessor interface {
\tProcess(ctx context.Context, cmd Command) (Result, error)
\tValidate(cmd Command) error
\tClose() error
}`,
  doc_text: '// CommandProcessor handles command execution and validation.',
}

const historyVersions: CodeUnit[] = [
  { ...mockCodeUnits[0], body_hash: 'hash_v3_abc', run_id: 44 },
  { ...mockCodeUnits[0], body_hash: 'hash_v2_def', run_id: 43 },
  { ...mockCodeUnits[0], body_hash: 'hash_v1_ghi', run_id: 42 },
]

export const Function: Story = {
  args: {
    codeUnit: funcUnit,
    onOpenInEditor: () => {},
    onAddToPlan: () => {},
  },
}

export const Type: Story = {
  args: {
    codeUnit: typeUnit,
    onOpenInEditor: () => {},
  },
}

export const Method: Story = {
  args: {
    codeUnit: methodUnit,
    onOpenInEditor: () => {},
    onAddToPlan: () => {},
  },
}

export const WithHistory: Story = {
  args: {
    codeUnit: funcUnit,
    history: historyVersions,
    onDiff: (h1, h2) => alert(`Diff: ${h1} vs ${h2}`),
    onOpenInEditor: () => {},
  },
}

export const WithDoc: Story = {
  args: {
    codeUnit: funcUnit,
    onOpenInEditor: () => {},
  },
}
