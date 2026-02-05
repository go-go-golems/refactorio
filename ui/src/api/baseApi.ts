import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

export function qs(params: Record<string, string | number | boolean | undefined>): string {
  const parts = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
  return parts.length > 0 ? `?${parts.join('&')}` : ''
}

export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Workspace', 'Session', 'Run', 'Symbol', 'CodeUnit', 'Commit', 'DiffRun', 'DocTerm', 'File'],
  endpoints: () => ({}),
})
