import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useSearchUnifiedQuery } from '../api/client'
import { GlobalSearchBar } from '../components/search/GlobalSearchBar'
import { SearchResults } from '../components/data-display/SearchResults'

export function SearchPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get('q') || '')

  const { data: results, isLoading, isFetching } = useSearchUnifiedQuery(
    { workspace_id: workspaceId!, request: { query } },
    { skip: !workspaceId || !query },
  )

  useEffect(() => {
    const q = searchParams.get('q')
    if (q && q !== query) setQuery(q)
  }, [searchParams]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSubmit = (value: string) => {
    setQuery(value)
    if (value) setSearchParams({ q: value })
    else setSearchParams({})
  }

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="p-4 d-flex flex-column h-100">
      <div className="mb-4">
        <GlobalSearchBar
          value={query}
          onChange={setQuery}
          onSubmit={handleSubmit}
          placeholder="Search symbols, code units, commits, diffs..."
          loading={isFetching}
          autoFocus
        />
      </div>
      <div className="flex-grow-1 overflow-auto">
        {!query ? (
          <div className="text-center text-muted py-5">Enter a search query to begin</div>
        ) : isLoading ? (
          <SearchResults results={[]} loading query={query} />
        ) : results && results.length > 0 ? (
          <SearchResults results={results} query={query} />
        ) : (
          <div className="text-center text-muted py-5">No results found for &ldquo;{query}&rdquo;</div>
        )}
      </div>
    </div>
  )
}
