import type { Meta, StoryObj } from '@storybook/react'
import { CodeViewer } from './CodeViewer'

const meta: Meta<typeof CodeViewer> = {
  title: 'CodeDisplay/CodeViewer',
  component: CodeViewer,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
  },
}

export default meta
type Story = StoryObj<typeof CodeViewer>

const goCode = `package handlers

import (
	"context"
)

// CommandProcessor handles command execution
type CommandProcessor interface {
	Process(ctx context.Context, cmd Command) (Result, error)
	Validate(cmd Command) error
}

// NewCommandProcessor creates a new CommandProcessor with the given options.
func NewCommandProcessor(opts ...Option) CommandProcessor {
	impl := &commandProcessorImpl{
		middleware: make([]Middleware, 0),
		validators: make([]Validator, 0),
	}
	for _, opt := range opts {
		opt(impl)
	}
	return impl
}`

const typeScriptCode = `interface SearchResult {
  type: 'symbol' | 'code_unit' | 'commit' | 'diff' | 'doc' | 'file'
  id: string
  label: string
  snippet?: string
  location?: string
  payload: unknown
}

export function SearchResults({
  results,
  groupByType = true,
  selectedId,
  onSelect,
  loading = false,
}: SearchResultsProps) {
  if (loading) {
    return <LoadingSkeleton />
  }
  return <ResultsList results={results} />
}`

export const Default: Story = {
  args: {
    content: goCode,
    language: 'go',
  },
}

export const WithHighlightedLines: Story = {
  args: {
    content: goCode,
    language: 'go',
    highlightLines: [8, 9, 10, 11],
  },
}

export const WithHighlightRange: Story = {
  args: {
    content: goCode,
    language: 'go',
    highlightRanges: [
      { startLine: 14, endLine: 23, className: 'highlight-focus' },
    ],
  },
}

export const MultipleHighlights: Story = {
  args: {
    content: goCode,
    language: 'go',
    highlightRanges: [
      { startLine: 7, endLine: 11, className: 'highlight-add' },
      { startLine: 14, endLine: 14, className: 'highlight-focus' },
    ],
  },
}

export const TypeScript: Story = {
  args: {
    content: typeScriptCode,
    language: 'typescript',
  },
}

export const WithClickableLines: Story = {
  args: {
    content: goCode,
    language: 'go',
    onLineClick: (line) => alert(`Clicked line ${line}`),
  },
}

export const NoLineNumbers: Story = {
  args: {
    content: goCode,
    language: 'go',
    showLineNumbers: false,
  },
}

export const CustomStartLine: Story = {
  args: {
    content: goCode.split('\n').slice(6, 12).join('\n'),
    language: 'go',
    startLine: 7,
    highlightLines: [8],
  },
}

export const WrapLines: Story = {
  args: {
    content: 'This is a very long line that should wrap when wrapLines is enabled. '.repeat(5),
    wrapLines: true,
  },
}

export const WithMaxHeight: Story = {
  args: {
    content: goCode + '\n' + goCode,
    language: 'go',
    maxHeight: '300px',
  },
}

export const EmptyContent: Story = {
  args: {
    content: '',
    language: 'go',
  },
}
