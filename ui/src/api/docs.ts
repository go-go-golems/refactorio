import { api, qs } from './baseApi'
import type { DocTerm, DocHit } from '../types/api'

const docsApi = api.injectEndpoints({
  endpoints: (builder) => ({
    getDocTerms: builder.query<DocTerm[], { workspace_id: string; run_id?: number; limit?: number; offset?: number }>({
      query: (params) => `/docs/terms${qs(params)}`,
      transformResponse: (response: { items: DocTerm[] }) => response.items,
      providesTags: ['DocTerm'],
    }),
    getDocHits: builder.query<DocHit[], { workspace_id: string; term?: string; run_id?: number; path_prefix?: string; limit?: number; offset?: number }>({
      query: (params) => `/docs/hits${qs(params)}`,
      transformResponse: (response: { items: DocHit[] }) => response.items,
    }),
  }),
})

export const {
  useGetDocTermsQuery,
  useGetDocHitsQuery,
} = docsApi
