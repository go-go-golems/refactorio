import { api, qs } from './baseApi'
import type { Run, RunSummary } from '../types/api'

const runsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getRuns: builder.query<Run[], { workspace_id: string; limit?: number; offset?: number }>({
      query: (params) => `/runs${qs(params)}`,
      transformResponse: (response: { items: Run[] }) => response.items,
      providesTags: ['Run'],
    }),
    getRun: builder.query<Run, { id: number; workspace_id: string }>({
      query: ({ id, ...params }) => `/runs/${id}${qs(params)}`,
      providesTags: (_r, _e, { id }) => [{ type: 'Run', id }],
    }),
    getRunSummary: builder.query<RunSummary, { id: number; workspace_id: string }>({
      query: ({ id, ...params }) => `/runs/${id}/summary${qs(params)}`,
    }),
  }),
})

export const {
  useGetRunsQuery,
  useGetRunQuery,
  useGetRunSummaryQuery,
} = runsApi
