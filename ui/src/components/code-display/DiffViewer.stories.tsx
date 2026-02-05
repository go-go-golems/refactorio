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
    hunk_id: 1,
    old_start: 10,
    old_count: 8,
    new_start: 10,
    new_count: 12,
    lines: [
      { kind: ' ', old_line: 10, new_line: 10, content: 'func NewCommandProcessor(opts ...Option) CommandProcessor {' },
      { kind: '-', old_line: 11, content: '\timpl := &Processor{' },
      { kind: '+', new_line: 11, content: '\timpl := &commandProcessorImpl{' },
      { kind: ' ', old_line: 12, new_line: 12, content: '\t\tmiddleware: make([]Middleware, 0),' },
      { kind: '+', new_line: 13, content: '\t\tvalidators: make([]Validator, 0),' },
      { kind: '+', new_line: 14, content: '\t\tlogger:     defaultLogger,' },
      { kind: ' ', old_line: 13, new_line: 15, content: '\t}' },
      { kind: ' ', old_line: 14, new_line: 16, content: '\tfor _, opt := range opts {' },
      { kind: '-', old_line: 15, content: '\t\topt.Apply(impl)' },
      { kind: '+', new_line: 17, content: '\t\topt(impl)' },
      { kind: ' ', old_line: 16, new_line: 18, content: '\t}' },
      { kind: ' ', old_line: 17, new_line: 19, content: '\treturn impl' },
    ],
  },
  {
    hunk_id: 2,
    old_start: 45,
    old_count: 5,
    new_start: 49,
    new_count: 7,
    lines: [
      { kind: ' ', old_line: 45, new_line: 49, content: '// Process handles the command execution' },
      { kind: '-', old_line: 46, content: 'func (p *Processor) Process(ctx context.Context, cmd Command) (Result, error) {' },
      { kind: '+', new_line: 50, content: 'func (p *commandProcessorImpl) Process(ctx context.Context, cmd Command) (Result, error) {' },
      { kind: '+', new_line: 51, content: '\tif err := p.Validate(cmd); err != nil {' },
      { kind: '+', new_line: 52, content: '\t\treturn Result{}, err' },
      { kind: '+', new_line: 53, content: '\t}' },
      { kind: ' ', old_line: 47, new_line: 54, content: '\treturn p.execute(ctx, cmd)' },
      { kind: ' ', old_line: 48, new_line: 55, content: '}' },
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
    onLineClick: (line) => alert(`Clicked: ${line.kind} ${line.content}`),
  },
}

export const AdditionsOnly: Story = {
  args: {
    hunks: [
      {
        hunk_id: 1,
        old_start: 0,
        old_count: 0,
        new_start: 1,
        new_count: 5,
        lines: [
          { kind: '+', new_line: 1, content: '// Package handlers provides command handling' },
          { kind: '+', new_line: 2, content: 'package handlers' },
          { kind: '+', new_line: 3, content: '' },
          { kind: '+', new_line: 4, content: 'import "context"' },
          { kind: '+', new_line: 5, content: '' },
        ],
      },
    ],
  },
}

export const DeletionsOnly: Story = {
  args: {
    hunks: [
      {
        hunk_id: 1,
        old_start: 1,
        old_count: 5,
        new_start: 0,
        new_count: 0,
        lines: [
          { kind: '-', old_line: 1, content: '// Deprecated: use handlers package instead' },
          { kind: '-', old_line: 2, content: 'package legacy' },
          { kind: '-', old_line: 3, content: '' },
          { kind: '-', old_line: 4, content: 'import "context"' },
          { kind: '-', old_line: 5, content: '' },
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
        hunk_id: 1,
        old_start: 1,
        old_count: 2,
        new_start: 1,
        new_count: 2,
        lines: [
          { kind: '-', old_line: 1, content: 'func processVeryLongFunctionNameWithManyParameters(ctx context.Context, param1 string, param2 int, param3 bool, param4 []string) error {' },
          { kind: '+', new_line: 1, content: 'func processCommand(ctx context.Context, cmd Command) error {' },
          { kind: ' ', old_line: 2, new_line: 2, content: '\treturn nil' },
        ],
      },
    ],
  },
}
