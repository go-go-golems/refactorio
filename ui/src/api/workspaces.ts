import { api, qs } from './baseApi'
import type { Workspace, WorkspaceConfig, DBInfo } from '../types/api'

const workspacesApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getWorkspaces: builder.query<WorkspaceConfig, void>({
      query: () => '/workspaces',
      providesTags: ['Workspace'],
    }),
    getDBInfo: builder.query<DBInfo, string>({
      query: (workspaceId) => `/db/info${qs({ workspace_id: workspaceId })}`,
    }),
    getWorkspace: builder.query<Workspace, string>({
      query: (id) => `/workspaces/${id}`,
      providesTags: (_r, _e, id) => [{ type: 'Workspace', id }],
    }),
    createWorkspace: builder.mutation<Workspace, Partial<Workspace>>({
      query: (body) => ({ url: '/workspaces', method: 'POST', body }),
      invalidatesTags: ['Workspace'],
    }),
    updateWorkspace: builder.mutation<Workspace, { id: string; data: Partial<Workspace> }>({
      query: ({ id, data }) => ({ url: `/workspaces/${id}`, method: 'PATCH', body: data }),
      invalidatesTags: (_r, _e, { id }) => [{ type: 'Workspace', id }, 'Workspace'],
    }),
    deleteWorkspace: builder.mutation<void, string>({
      query: (id) => ({ url: `/workspaces/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Workspace'],
    }),
  }),
})

export const {
  useGetWorkspacesQuery,
  useGetWorkspaceQuery,
  useGetDBInfoQuery,
  useCreateWorkspaceMutation,
  useUpdateWorkspaceMutation,
  useDeleteWorkspaceMutation,
} = workspacesApi
