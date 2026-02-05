import type {
  Workspace,
  WorkspaceConfig,
  DBInfo,
  Run,
  RunSummary,
  Session,
  Symbol,
  SymbolRef,
  CodeUnit,
  CodeUnitDetail,
  Commit,
  CommitFile,
  DiffRun,
  DiffFile,
  DiffHunk,
  DocTerm,
  DocHit,
  FileEntry,
  SearchResult,
  SearchRequest,
  ErrorResponse,
} from '../types/api'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText })) as ErrorResponse
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

function qs(params: Record<string, string | number | boolean | undefined>): string {
  const parts = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
  return parts.length > 0 ? `?${parts.join('&')}` : ''
}

// Workspaces
export const workspaces = {
  list: () => request<WorkspaceConfig>('/workspaces'),
  get: (id: string) => request<Workspace>(`/workspaces/${id}`),
  create: (data: Partial<Workspace>) => request<Workspace>('/workspaces', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<Workspace>) => request<Workspace>(`/workspaces/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/workspaces/${id}`, { method: 'DELETE' }),
}

// DB Info
export const db = {
  info: (workspaceId: string) => request<DBInfo>(`/db/info${qs({ workspace_id: workspaceId })}`),
}

// Runs
export const runs = {
  list: (params: { workspace_id: string; limit?: number; offset?: number }) =>
    request<Run[]>(`/runs${qs(params)}`),
  get: (id: number, workspaceId: string) =>
    request<Run>(`/runs/${id}${qs({ workspace_id: workspaceId })}`),
  summary: (id: number, workspaceId: string) =>
    request<RunSummary>(`/runs/${id}/summary${qs({ workspace_id: workspaceId })}`),
}

// Sessions
export const sessions = {
  list: (workspaceId: string) =>
    request<Session[]>(`/sessions${qs({ workspace_id: workspaceId })}`),
  get: (id: string, workspaceId: string) =>
    request<Session>(`/sessions/${id}${qs({ workspace_id: workspaceId })}`),
}

// Symbols
export const symbols = {
  list: (params: { workspace_id: string; run_id?: number; kind?: string; exported?: boolean; q?: string; limit?: number; offset?: number }) =>
    request<Symbol[]>(`/symbols${qs(params)}`),
  get: (hash: string, workspaceId: string) =>
    request<Symbol>(`/symbols/${hash}${qs({ workspace_id: workspaceId })}`),
  refs: (hash: string, params: { workspace_id: string; run_id?: number; limit?: number; offset?: number }) =>
    request<SymbolRef[]>(`/symbols/${hash}/refs${qs(params)}`),
}

// Code Units
export const codeUnits = {
  list: (params: { workspace_id: string; run_id?: number; kind?: string; q?: string; limit?: number; offset?: number }) =>
    request<CodeUnit[]>(`/code-units${qs(params)}`),
  get: (hash: string, workspaceId: string) =>
    request<CodeUnitDetail>(`/code-units/${hash}${qs({ workspace_id: workspaceId })}`),
  history: (hash: string, workspaceId: string) =>
    request<CodeUnit[]>(`/code-units/${hash}/history${qs({ workspace_id: workspaceId })}`),
}

// Commits
export const commits = {
  list: (params: { workspace_id: string; run_id?: number; author?: string; path?: string; q?: string; limit?: number; offset?: number }) =>
    request<Commit[]>(`/commits${qs(params)}`),
  get: (hash: string, workspaceId: string) =>
    request<Commit>(`/commits/${hash}${qs({ workspace_id: workspaceId })}`),
  files: (hash: string, workspaceId: string) =>
    request<CommitFile[]>(`/commits/${hash}/files${qs({ workspace_id: workspaceId })}`),
}

// Diffs
export const diffs = {
  runs: (params: { workspace_id: string; session_id?: string }) =>
    request<DiffRun[]>(`/diff-runs${qs(params)}`),
  files: (runId: number, workspaceId: string) =>
    request<DiffFile[]>(`/diff/${runId}/files${qs({ workspace_id: workspaceId })}`),
  file: (runId: number, params: { workspace_id: string; file_path: string }) =>
    request<DiffHunk[]>(`/diff/${runId}/file${qs(params)}`),
}

// Docs
export const docs = {
  terms: (params: { workspace_id: string; run_id?: number; limit?: number; offset?: number }) =>
    request<DocTerm[]>(`/docs/terms${qs(params)}`),
  hits: (params: { workspace_id: string; term?: string; run_id?: number; path_prefix?: string; limit?: number; offset?: number }) =>
    request<DocHit[]>(`/docs/hits${qs(params)}`),
}

// Files
export const files = {
  list: (params: { workspace_id: string; prefix?: string }) =>
    request<FileEntry[]>(`/files${qs(params)}`),
  content: (params: { workspace_id: string; path: string; ref?: string }) =>
    request<{ content: string; path: string }>(`/file${qs(params)}`),
  history: (params: { workspace_id: string; path: string }) =>
    request<Commit[]>(`/files/history${qs(params)}`),
}

// Search
export const search = {
  unified: (data: SearchRequest, workspaceId: string) =>
    request<SearchResult[]>(`/search${qs({ workspace_id: workspaceId })}`, { method: 'POST', body: JSON.stringify(data) }),
  symbols: (params: { workspace_id: string; q: string; run_id?: number; limit?: number }) =>
    request<SearchResult[]>(`/search/symbols${qs(params)}`),
  codeUnits: (params: { workspace_id: string; q: string; run_id?: number; limit?: number }) =>
    request<SearchResult[]>(`/search/code-units${qs(params)}`),
  commits: (params: { workspace_id: string; q: string; run_id?: number; limit?: number }) =>
    request<SearchResult[]>(`/search/commits${qs(params)}`),
}
