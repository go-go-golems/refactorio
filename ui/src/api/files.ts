import { api, qs } from './baseApi'
import type { FileEntry, Commit } from '../types/api'

const filesApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getFiles: builder.query<FileEntry[], { workspace_id: string; prefix?: string }>({
      query: (params) => `/files${qs(params)}`,
      transformResponse: (response: { items: FileEntry[] }) => response.items,
      providesTags: ['File'],
    }),
    getFileContent: builder.query<{ content: string; path: string }, { workspace_id: string; path: string; ref?: string }>({
      query: (params) => `/file${qs(params)}`,
    }),
    getFileHistory: builder.query<Commit[], { workspace_id: string; path: string; run_id?: number }>({
      query: (params) => `/files/history${qs(params)}`,
      transformResponse: (response: { items: Commit[] }) => response.items,
    }),
  }),
})

export const {
  useGetFilesQuery,
  useGetFileContentQuery,
  useGetFileHistoryQuery,
} = filesApi
