import { api, qs } from './baseApi'
import type { SearchResult, SearchRequest, Symbol, CodeUnitDetail, Commit } from '../types/api'

const searchApi = api.injectEndpoints({
  endpoints: (builder) => ({
    searchUnified: builder.query<SearchResult[], { workspace_id: string; request: SearchRequest }>({
      query: ({ workspace_id, request }) => ({
        url: `/search${qs({ workspace_id })}`,
        method: 'POST',
        body: request,
      }),
      transformResponse: (response: { items: SearchResult[] }) => response.items,
    }),
    searchSymbols: builder.query<Symbol[], { workspace_id: string; q: string; run_id?: number; limit?: number }>({
      query: (params) => `/search/symbols${qs(params)}`,
      transformResponse: (response: { items: Symbol[] }) => response.items,
    }),
    searchCodeUnits: builder.query<CodeUnitDetail[], { workspace_id: string; q: string; run_id?: number; limit?: number }>({
      query: (params) => `/search/code-units${qs(params)}`,
      transformResponse: (response: { items: CodeUnitDetail[] }) => response.items,
    }),
    searchCommits: builder.query<Commit[], { workspace_id: string; q: string; run_id?: number; limit?: number }>({
      query: (params) => `/search/commits${qs(params)}`,
      transformResponse: (response: { items: Commit[] }) => response.items,
    }),
  }),
})

export const {
  useSearchUnifiedQuery,
  useSearchSymbolsQuery,
  useSearchCodeUnitsQuery,
  useSearchCommitsQuery,
} = searchApi
