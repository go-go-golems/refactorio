import { api, qs } from './baseApi'
import type { Symbol, SymbolRef } from '../types/api'

const symbolsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getSymbols: builder.query<Symbol[], { workspace_id: string; run_id?: number; kind?: string; exported_only?: boolean; name?: string; pkg?: string; path?: string; limit?: number; offset?: number }>({
      query: (params) => `/symbols${qs(params)}`,
      transformResponse: (response: { items: Symbol[] }) => response.items,
      providesTags: ['Symbol'],
    }),
    getSymbol: builder.query<Symbol, { hash: string; workspace_id: string; run_id?: number }>({
      query: ({ hash, ...params }) => `/symbols/${hash}${qs(params)}`,
      providesTags: (_r, _e, { hash }) => [{ type: 'Symbol', id: hash }],
    }),
    getSymbolRefs: builder.query<SymbolRef[], { hash: string; workspace_id: string; run_id?: number; limit?: number; offset?: number }>({
      query: ({ hash, ...params }) => `/symbols/${hash}/refs${qs(params)}`,
      transformResponse: (response: { items: SymbolRef[] }) => response.items,
    }),
  }),
})

export const {
  useGetSymbolsQuery,
  useGetSymbolQuery,
  useGetSymbolRefsQuery,
} = symbolsApi
