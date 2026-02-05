import { api, qs } from './baseApi'
import type { DiffRun, DiffFile, DiffHunk } from '../types/api'

const diffsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getDiffRuns: builder.query<DiffRun[], { workspace_id: string; session_id?: string }>({
      query: (params) => `/diff-runs${qs(params)}`,
      transformResponse: (response: { items: DiffRun[] }) => response.items,
      providesTags: ['DiffRun'],
    }),
    getDiffFiles: builder.query<DiffFile[], { run_id: number; workspace_id: string }>({
      query: ({ run_id, ...params }) => `/diff/${run_id}/files${qs(params)}`,
      transformResponse: (response: { items: DiffFile[] }) => response.items,
    }),
    getDiffFile: builder.query<DiffHunk[], { run_id: number; workspace_id: string; file_path: string }>({
      query: ({ run_id, ...params }) => `/diff/${run_id}/file${qs(params)}`,
      transformResponse: (response: { hunks: DiffHunk[] }) => response.hunks,
    }),
  }),
})

export const {
  useGetDiffRunsQuery,
  useGetDiffFilesQuery,
  useGetDiffFileQuery,
} = diffsApi
