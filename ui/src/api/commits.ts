import { api, qs } from './baseApi'
import type { Commit, CommitFile } from '../types/api'

const commitsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getCommits: builder.query<Commit[], { workspace_id: string; run_id?: number; author?: string; path?: string; q?: string; limit?: number; offset?: number }>({
      query: (params) => `/commits${qs(params)}`,
      transformResponse: (response: { items: Commit[] }) => response.items,
      providesTags: ['Commit'],
    }),
    getCommit: builder.query<Commit, { hash: string; workspace_id: string }>({
      query: ({ hash, ...params }) => `/commits/${hash}${qs(params)}`,
      providesTags: (_r, _e, { hash }) => [{ type: 'Commit', id: hash }],
    }),
    getCommitFiles: builder.query<CommitFile[], { hash: string; workspace_id: string }>({
      query: ({ hash, ...params }) => `/commits/${hash}/files${qs(params)}`,
      transformResponse: (response: { items: CommitFile[] }) => response.items,
    }),
  }),
})

export const {
  useGetCommitsQuery,
  useGetCommitQuery,
  useGetCommitFilesQuery,
} = commitsApi
