import { api, qs } from './baseApi'
import type { Session } from '../types/api'

const sessionsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getSessions: builder.query<Session[], string>({
      query: (workspaceId) => `/sessions${qs({ workspace_id: workspaceId })}`,
      transformResponse: (response: { sessions: Session[] }) => response.sessions,
      providesTags: ['Session'],
    }),
    getSession: builder.query<Session, { id: string; workspace_id: string }>({
      query: ({ id, ...params }) => `/sessions/${id}${qs(params)}`,
      providesTags: (_r, _e, { id }) => [{ type: 'Session', id }],
    }),
  }),
})

export const {
  useGetSessionsQuery,
  useGetSessionQuery,
} = sessionsApi
