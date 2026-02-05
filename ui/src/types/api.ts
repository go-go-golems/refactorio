// API Types for Refactorio Workbench

export interface Workspace {
  id: string
  name: string
  db_path: string
  repo_root?: string
  created_at: string
  updated_at: string
}

export interface DBInfo {
  workspace_id?: string
  db_path: string
  repo_root?: string
  schema_version: number
  tables: Record<string, boolean>
  fts_tables: Record<string, boolean>
  features: Record<string, boolean>
  views?: Record<string, boolean>
}

export interface Run {
  id: number
  status: 'running' | 'success' | 'failed'
  root_path?: string
  git_from?: string
  git_to?: string
  tool_version?: string
  args_json?: string
  error_json?: string
  started_at: string
  finished_at?: string
  sources_dir?: string
}

export interface RunSummary {
  run_id: number
  counts: Record<string, number>
}

export interface Session {
  id: string
  workspace_id?: string
  root_path?: string
  git_from?: string
  git_to?: string
  runs: SessionRuns
  availability: Record<string, boolean>
  last_updated?: string
}

export interface SessionRuns {
  commits?: number
  diff?: number
  symbols?: number
  code_units?: number
  doc_hits?: number
  gopls_refs?: number
}

export interface Symbol {
  symbol_hash: string
  name: string
  kind: string
  pkg: string
  recv?: string
  signature?: string
  file: string
  line: number
  col: number
  is_exported: boolean
  run_id: number
}

export interface SymbolRef {
  run_id: number
  commit_hash?: string
  symbol_hash: string
  path: string
  line: number
  col: number
  is_decl: boolean
  source: string
}

export interface CodeUnit {
  run_id: number
  unit_hash: string
  kind: string
  name: string
  recv?: string
  pkg: string
  file: string
  start_line: number
  start_col: number
  end_line: number
  end_col: number
  body_hash?: string
  signature?: string
}

export interface CodeUnitDetail extends CodeUnit {
  body_text: string
  doc_text?: string
}

export interface Commit {
  run_id: number
  hash: string
  author_name?: string
  author_email?: string
  author_date?: string
  committer_date?: string
  subject?: string
  body?: string
}

export interface CommitFile {
  path: string
  status: string
  old_path?: string
  new_path?: string
  blob_old?: string
  blob_new?: string
}

export interface DiffRun {
  id: number
  root_path?: string
  git_from?: string
  git_to?: string
}

export interface DiffFile {
  run_id: number
  status: string
  path: string
  old_path?: string
  new_path?: string
}

export interface DiffHunk {
  id: number
  old_start: number
  old_lines: number
  new_start: number
  new_lines: number
  lines: DiffLine[]
}

export interface DiffLine {
  kind: '+' | '-' | ' '
  line_no_old?: number
  line_no_new?: number
  text: string
}

export interface DocTerm {
  term: string
  count: number
}

export interface DocHit {
  run_id: number
  term: string
  path: string
  line: number
  col: number
  match_text: string
}

export interface FileEntry {
  path: string
  kind: 'file' | 'dir'
  ext?: string
  exists?: boolean
  is_binary?: boolean
}

export interface SearchResult {
  type: 'symbol' | 'code_unit' | 'commit' | 'diff' | 'doc' | 'file'
  primary: string
  secondary?: string
  path?: string
  line?: number
  col?: number
  snippet?: string
  run_id?: number
  commit_hash?: string
  payload?: unknown
}

export interface SearchRequest {
  query: string
  types?: string[]
  session_id?: string
  filters?: {
    path?: string
    pkg?: string
    symbol_kind?: string
    term?: string
  }
  run_ids?: Record<string, number>
  limit?: number
  offset?: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total?: number
  limit: number
  offset: number
}

export interface ErrorResponse {
  error: string
  code?: string
  details?: unknown
}
