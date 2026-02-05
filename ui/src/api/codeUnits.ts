import { api, qs } from './baseApi'
import type { CodeUnit, CodeUnitDetail } from '../types/api'

const codeUnitsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getCodeUnits: builder.query<CodeUnit[], { workspace_id: string; run_id?: number; kind?: string; name?: string; pkg?: string; path?: string; body_q?: string; limit?: number; offset?: number }>({
      query: (params) => `/code-units${qs(params)}`,
      transformResponse: (response: { items: CodeUnit[] }) => response.items,
      providesTags: ['CodeUnit'],
    }),
    getCodeUnit: builder.query<CodeUnitDetail, { hash: string; workspace_id: string; run_id?: number }>({
      query: ({ hash, ...params }) => `/code-units/${hash}${qs(params)}`,
      providesTags: (_r, _e, { hash }) => [{ type: 'CodeUnit', id: hash }],
    }),
    getCodeUnitHistory: builder.query<CodeUnit[], { hash: string; workspace_id: string }>({
      query: ({ hash, ...params }) => `/code-units/${hash}/history${qs(params)}`,
      transformResponse: (response: { items: CodeUnit[] }) => response.items,
    }),
  }),
})

export const {
  useGetCodeUnitsQuery,
  useGetCodeUnitQuery,
  useGetCodeUnitHistoryQuery,
} = codeUnitsApi
