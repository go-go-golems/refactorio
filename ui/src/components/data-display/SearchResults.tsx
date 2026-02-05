import { useState } from 'react'
import type { SearchResult } from '../../types/api'
import { EntityIcon } from '../foundation'

export interface SearchResultsProps {
  /** Search results */
  results: SearchResult[]
  /** Group results by type */
  groupByType?: boolean
  /** Currently selected result ID */
  selectedId?: string
  /** Called when a result is clicked */
  onSelect?: (result: SearchResult) => void
  /** Loading state */
  loading?: boolean
  /** Search query (for highlighting) */
  query?: string
  /** Custom class name */
  className?: string
}

interface GroupedResults {
  type: SearchResult['type']
  label: string
  results: SearchResult[]
}

const typeLabels: Record<SearchResult['type'], string> = {
  symbol: 'Symbols',
  code_unit: 'Code Units',
  commit: 'Commits',
  diff: 'Diffs',
  doc: 'Docs',
  file: 'Files',
}

function highlightMatch(text: string, query?: string): React.ReactNode {
  if (!query || !text) return text
  const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  const parts = text.split(regex)
  return parts.map((part, i) =>
    regex.test(part) ? (
      <mark key={i} className="bg-warning-subtle px-0">
        {part}
      </mark>
    ) : (
      part
    )
  )
}

function SkeletonResult() {
  return (
    <div className="list-group-item">
      <div className="placeholder-glow">
        <span className="placeholder col-4 mb-1"></span>
        <span className="placeholder col-8"></span>
      </div>
    </div>
  )
}

function buildResultId(result: SearchResult): string {
  return [
    result.type,
    result.primary,
    result.path ?? '',
    result.line ?? '',
    result.col ?? '',
    result.run_id ?? '',
    result.commit_hash ?? '',
  ].join('|')
}

function ResultItem({
  result,
  selected,
  onClick,
  query,
}: {
  result: SearchResult
  selected: boolean
  onClick?: () => void
  query?: string
}) {
  return (
    <button
      type="button"
      className={`list-group-item list-group-item-action d-flex align-items-start gap-2 ${selected ? 'active' : ''}`}
      onClick={onClick}
    >
      <EntityIcon
        type={result.type}
        kind={result.type === 'symbol' ? (result.payload as { kind?: string })?.kind : undefined}
        size="sm"
      />
      <div className="flex-grow-1 min-width-0">
        <div className="d-flex justify-content-between align-items-center">
          <span className={`fw-medium ${selected ? '' : ''}`}>
            {highlightMatch(result.primary, query)}
          </span>
          <span className={`badge ${selected ? 'bg-light text-dark' : 'bg-secondary-subtle text-secondary'}`}>
            {result.type}
          </span>
        </div>
        {result.snippet && (
          <div className={`small ${selected ? 'text-light' : 'text-muted'} text-truncate`}>
            {highlightMatch(result.snippet, query)}
          </div>
        )}
        {result.path && (
          <div className={`small ${selected ? 'text-light opacity-75' : 'text-muted'}`}>
            <code className="small">{result.path}{result.line ? `:${result.line}` : ''}</code>
          </div>
        )}
      </div>
    </button>
  )
}

function ResultGroup({
  group,
  selectedId,
  onSelect,
  query,
  defaultOpen = true,
}: {
  group: GroupedResults
  selectedId?: string
  onSelect?: (result: SearchResult) => void
  query?: string
  defaultOpen?: boolean
}) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className="mb-3">
      <button
        type="button"
        className="btn btn-link text-decoration-none w-100 d-flex align-items-center justify-content-between p-2 text-body-secondary"
        onClick={() => setOpen(!open)}
      >
        <span className="d-flex align-items-center gap-2">
          <EntityIcon type={group.type} size="sm" />
          <span className="fw-semibold">{group.label}</span>
          <span className="badge bg-secondary-subtle text-secondary">{group.results.length}</span>
        </span>
        <span style={{ transform: open ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }}>
          ›
        </span>
      </button>
      {open && (
        <div className="list-group list-group-flush">
          {group.results.map((result) => (
            <ResultItem
              key={buildResultId(result)}
              result={result}
              selected={buildResultId(result) === selectedId}
              onClick={() => onSelect?.(result)}
              query={query}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function SearchResults({
  results,
  groupByType = true,
  selectedId,
  onSelect,
  loading = false,
  query,
  className = '',
}: SearchResultsProps) {
  if (loading) {
    return (
      <div className={`list-group list-group-flush ${className}`}>
        {Array.from({ length: 5 }).map((_, i) => (
          <SkeletonResult key={i} />
        ))}
      </div>
    )
  }

  if (results.length === 0) {
    return (
      <div className={`text-center text-muted py-5 ${className}`}>
        <p className="mb-1">No results found</p>
        {query && <p className="small">Try adjusting your search query</p>}
      </div>
    )
  }

  if (groupByType) {
    // Group results by type
    const groups: GroupedResults[] = []
    const typeOrder: SearchResult['type'][] = ['symbol', 'code_unit', 'diff', 'commit', 'doc', 'file']

    for (const type of typeOrder) {
      const typeResults = results.filter((r) => r.type === type)
      if (typeResults.length > 0) {
        groups.push({
          type,
          label: typeLabels[type],
          results: typeResults,
        })
      }
    }

    return (
      <div className={className}>
        {groups.map((group) => (
          <ResultGroup
            key={group.type}
            group={group}
            selectedId={selectedId}
            onSelect={onSelect}
            query={query}
          />
        ))}
      </div>
    )
  }

  // Flat list
  return (
    <div className={`list-group list-group-flush ${className}`}>
      {results.map((result) => (
        <ResultItem
          key={buildResultId(result)}
          result={result}
          selected={buildResultId(result) === selectedId}
          onClick={() => onSelect?.(result)}
          query={query}
        />
      ))}
    </div>
  )
}
