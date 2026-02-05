// API Types for Refactorio Workbench

export interface Workspace {
  id: string
  name: string
  db_path: string
  repo_root?: string
  created_at: string
  updated_at: string
}

export interface WorkspaceConfig {
  workspaces: Workspace[]
}

export interface DBInfo {
  schema_version: number
  tables: string[]
  fts_tables: string[]
  features: {
    gopls_refs: boolean
    tree_sitter: boolean
    doc_hits: boolean
    code_units: boolean
  }
  row_counts: Record<string, number>
}

export interface Run {
  run_id: number
  status: 'running' | 'success' | 'failed'
  root_path: string
  git_from?: string
  git_to?: string
  args_json?: string
  error_json?: string
  started_at: string
  finished_at?: string
}

export interface RunSummary {
  run_id: number
  symbols_count: number
  code_units_count: number
  commits_count: number
  diff_files_count: number
  diff_lines_count: number
  doc_hits_count: number
}

export interface Session {
  id: string
  root_path: string
  git_from?: string
  git_to?: string
  runs: SessionRuns
  availability: SessionAvailability
  last_updated: string
}

export interface SessionRuns {
  commits?: number
  diff?: number
  symbols?: number
  code_units?: number
  doc_hits?: number
  gopls_refs?: number
  tree_sitter?: number
}

export interface SessionAvailability {
  commits: boolean
  diff: boolean
  symbols: boolean
  code_units: boolean
  doc_hits: boolean
  gopls_refs: boolean
  tree_sitter: boolean
}

export interface Symbol {
  symbol_hash: string
  name: string
  kind: string
  package_path: string
  signature?: string
  exported: boolean
  file_path: string
  start_line: number
  start_col: number
  end_line: number
  end_col: number
  run_id: number
}

export interface SymbolRef {
  file_path: string
  start_line: number
  start_col: number
  end_line: number
  end_col: number
  is_declaration: boolean
}

export interface CodeUnit {
  code_unit_hash: string
  kind: string
  name: string
  receiver?: string
  package_path: string
  file_path: string
  start_line: number
  start_col: number
  end_line: number
  end_col: number
  body_hash: string
  run_id: number
}

export interface CodeUnitDetail extends CodeUnit {
  body: string
  doc_comment?: string
}

export interface Commit {
  commit_hash: string
  subject: string
  body?: string
  author_name: string
  author_email: string
  commit_date: string
  run_id: number
}

export interface CommitFile {
  file_path: string
  status: string
  additions: number
  deletions: number
}

export interface DiffRun {
  run_id: number
  root_path: string
  git_from: string
  git_to: string
  files_count: number
}

export interface DiffFile {
  file_path: string
  status: string
  hunks_count: number
  additions: number
  deletions: number
}

export interface DiffHunk {
  hunk_id: number
  old_start: number
  old_count: number
  new_start: number
  new_count: number
  lines: DiffLine[]
}

export interface DiffLine {
  kind: '+' | '-' | ' '
  old_line?: number
  new_line?: number
  content: string
}

export interface DocTerm {
  term: string
  count: number
}

export interface DocHit {
  file_path: string
  term: string
  line: number
  col: number
  match_text: string
  run_id: number
}

export interface FileEntry {
  path: string
  is_dir: boolean
  children_count?: number
}

export interface SearchResult {
  type: 'symbol' | 'code_unit' | 'commit' | 'diff' | 'doc' | 'file'
  id: string
  label: string
  snippet?: string
  location?: string
  payload: unknown
}

export interface SearchRequest {
  query: string
  types?: string[]
  run_id?: number
  session_id?: string
  file_path?: string
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
