import type { Meta, StoryObj } from '@storybook/react'
import { DiffViewer } from './DiffViewer'
import { mockDiffHunks } from '../../mocks/data'
import type { DiffHunk } from '../../types/api'

const meta: Meta<typeof DiffViewer> = {
  title: 'CodeDisplay/DiffViewer',
  component: DiffViewer,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
  },
}

export default meta
type Story = StoryObj<typeof DiffViewer>

const complexHunks: DiffHunk[] = [
  {
    id: 1,
    old_start: 10,
    old_lines: 8,
    new_start: 10,
    new_lines: 12,
    lines: [
      { kind: ' ', line_no_old: 10, line_no_new: 10, text: 'func NewCommandProcessor(opts ...Option) CommandProcessor {' },
      { kind: '-', line_no_old: 11, text: '\timpl := &Processor{' },
      { kind: '+', line_no_new: 11, text: '\timpl := &commandProcessorImpl{' },
      { kind: ' ', line_no_old: 12, line_no_new: 12, text: '\t\tmiddleware: make([]Middleware, 0),' },
      { kind: '+', line_no_new: 13, text: '\t\tvalidators: make([]Validator, 0),' },
      { kind: '+', line_no_new: 14, text: '\t\tlogger:     defaultLogger,' },
      { kind: ' ', line_no_old: 13, line_no_new: 15, text: '\t}' },
      { kind: ' ', line_no_old: 14, line_no_new: 16, text: '\tfor _, opt := range opts {' },
      { kind: '-', line_no_old: 15, text: '\t\topt.Apply(impl)' },
      { kind: '+', line_no_new: 17, text: '\t\topt(impl)' },
      { kind: ' ', line_no_old: 16, line_no_new: 18, text: '\t}' },
      { kind: ' ', line_no_old: 17, line_no_new: 19, text: '\treturn impl' },
    ],
  },
  {
    id: 2,
    old_start: 45,
    old_lines: 5,
    new_start: 49,
    new_lines: 7,
    lines: [
      { kind: ' ', line_no_old: 45, line_no_new: 49, text: '// Process handles the command execution' },
      { kind: '-', line_no_old: 46, text: 'func (p *Processor) Process(ctx context.Context, cmd Command) (Result, error) {' },
      { kind: '+', line_no_new: 50, text: 'func (p *commandProcessorImpl) Process(ctx context.Context, cmd Command) (Result, error) {' },
      { kind: '+', line_no_new: 51, text: '\tif err := p.Validate(cmd); err != nil {' },
      { kind: '+', line_no_new: 52, text: '\t\treturn Result{}, err' },
      { kind: '+', line_no_new: 53, text: '\t}' },
      { kind: ' ', line_no_old: 47, line_no_new: 54, text: '\treturn p.execute(ctx, cmd)' },
      { kind: ' ', line_no_old: 48, line_no_new: 55, text: '}' },
    ],
  },
]

export const Default: Story = {
  args: {
    hunks: mockDiffHunks,
  },
}

export const MultipleHunks: Story = {
  args: {
    hunks: complexHunks,
  },
}

export const WithHighlight: Story = {
  args: {
    hunks: complexHunks,
    highlightQuery: 'commandProcessorImpl',
  },
}

export const NoLineNumbers: Story = {
  args: {
    hunks: mockDiffHunks,
    showLineNumbers: false,
  },
}

export const WithClickableLines: Story = {
  args: {
    hunks: mockDiffHunks,
    onLineClick: (line) => alert(`Clicked: ${line.kind} ${line.text}`),
  },
}

export const AdditionsOnly: Story = {
  args: {
    hunks: [
      {
        id: 1,
        old_start: 0,
        old_lines: 0,
        new_start: 1,
        new_lines: 5,
        lines: [
          { kind: '+', line_no_new: 1, text: '// Package handlers provides command handling' },
          { kind: '+', line_no_new: 2, text: 'package handlers' },
          { kind: '+', line_no_new: 3, text: '' },
          { kind: '+', line_no_new: 4, text: 'import "context"' },
          { kind: '+', line_no_new: 5, text: '' },
        ],
      },
    ],
  },
}

export const DeletionsOnly: Story = {
  args: {
    hunks: [
      {
        id: 1,
        old_start: 1,
        old_lines: 5,
        new_start: 0,
        new_lines: 0,
        lines: [
          { kind: '-', line_no_old: 1, text: '// Deprecated: use handlers package instead' },
          { kind: '-', line_no_old: 2, text: 'package legacy' },
          { kind: '-', line_no_old: 3, text: '' },
          { kind: '-', line_no_old: 4, text: 'import "context"' },
          { kind: '-', line_no_old: 5, text: '' },
        ],
      },
    ],
  },
}

export const Empty: Story = {
  args: {
    hunks: [],
  },
}

export const LongLines: Story = {
  args: {
    hunks: [
      {
        id: 1,
        old_start: 1,
        old_lines: 2,
        new_start: 1,
        new_lines: 2,
        lines: [
          { kind: '-', line_no_old: 1, text: 'func processVeryLongFunctionNameWithManyParameters(ctx context.Context, param1 string, param2 int, param3 bool, param4 []string) error {' },
          { kind: '+', line_no_new: 1, text: 'func processCommand(ctx context.Context, cmd Command) error {' },
          { kind: ' ', line_no_old: 2, line_no_new: 2, text: '\treturn nil' },
        ],
      },
    ],
  },
}
