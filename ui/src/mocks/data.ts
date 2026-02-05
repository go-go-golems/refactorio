// Mock data for Refactorio Workbench API

import type {
  Workspace,
  DBInfo,
  Run,
  RunSummary,
  Session,
  Symbol,
  SymbolRef,
  CodeUnit,
  Commit,
  CommitFile,
  DiffFile,
  DiffHunk,
  DocTerm,
  DocHit,
  FileEntry,
  SearchResult,
} from '../types/api'

export const mockWorkspaces: Workspace[] = [
  {
    id: 'glazed',
    name: 'glazed',
    db_path: '/Users/dev/glazed.db',
    repo_root: '/Users/dev/src/glazed',
    created_at: '2026-02-01T10:00:00Z',
    updated_at: '2026-02-05T09:00:00Z',
  },
  {
    id: 'refactorio',
    name: 'refactorio',
    db_path: '/Users/dev/refactorio.db',
    repo_root: '/Users/dev/src/refactorio',
    created_at: '2026-02-03T14:00:00Z',
    updated_at: '2026-02-05T08:30:00Z',
  },
]

export const mockDBInfo: DBInfo = {
  db_path: '/tmp/refactorio.db',
  schema_version: 12,
  tables: {
    meta_runs: true,
    symbols: true,
    symbol_occurrences: true,
    symbol_refs: true,
    code_unit_snapshots: true,
    commits: true,
    commit_files: true,
    diff_files: true,
    diff_hunks: true,
    diff_lines: true,
    doc_hits: true,
    files: true,
  },
  fts_tables: {
    symbols_fts: true,
    code_units_fts: true,
    commits_fts: true,
    diff_lines_fts: true,
    doc_hits_fts: true,
  },
  features: {
    gopls_refs: true,
    doc_hits: true,
    code_units: true,
  },
}

export const mockRuns: Run[] = [
  {
    id: 44,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T08:00:00Z',
    finished_at: '2026-02-05T08:05:00Z',
  },
  {
    id: 43,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T07:55:00Z',
    finished_at: '2026-02-05T08:00:00Z',
  },
  {
    id: 42,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T07:50:00Z',
    finished_at: '2026-02-05T07:55:00Z',
  },
  {
    id: 41,
    status: 'failed',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    error_json: '{"message": "indexing parser failed"}',
    started_at: '2026-02-05T07:45:00Z',
    finished_at: '2026-02-05T07:46:00Z',
  },
]

export const mockRunSummary: RunSummary = {
  run_id: 44,
  counts: {
    symbol_occurrences: 234,
    code_unit_snapshots: 189,
    commits: 47,
    diff_files: 23,
    diff_lines: 891,
    doc_hits: 45,
  },
}

export const mockSessions: Session[] = [
  {
    id: 'main-head20-head-a7b3',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    runs: {
      commits: 42,
      diff: 43,
      symbols: 44,
      code_units: 44,
      doc_hits: 45,
    },
    availability: {
      commits: true,
      diff: true,
      symbols: true,
      code_units: true,
      doc_hits: true,
      gopls_refs: false,
    },
    last_updated: '2026-02-05T09:00:00Z',
  },
]

export const mockSymbols: Symbol[] = [
  {
    symbol_hash: 'a7b3c9f2',
    name: 'CommandProcessor',
    kind: 'type',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'type CommandProcessor interface { Process(ctx context.Context, cmd Command) (Result, error) }',
    is_exported: true,
    file: 'pkg/handlers/command.go',
    line: 45,
    col: 1,
    run_id: 44,
  },
  {
    symbol_hash: 'b8c4d0e3',
    name: 'NewCommandProcessor',
    kind: 'func',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'func NewCommandProcessor(opts ...Option) CommandProcessor',
    is_exported: true,
    file: 'pkg/handlers/command.go',
    line: 67,
    col: 1,
    run_id: 44,
  },
  {
    symbol_hash: 'c9d5e1f4',
    name: 'commandProcessorImpl',
    kind: 'type',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'type commandProcessorImpl struct { ... }',
    is_exported: false,
    file: 'pkg/handlers/command.go',
    line: 89,
    col: 1,
    run_id: 44,
  },
  {
    symbol_hash: 'd0e6f2g5',
    name: 'Process',
    kind: 'method',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'func (p *commandProcessorImpl) Process(ctx context.Context, cmd Command) (Result, error)',
    is_exported: true,
    file: 'pkg/handlers/command.go',
    line: 127,
    col: 1,
    run_id: 44,
  },
]

export const mockSymbolRefs: SymbolRef[] = [
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/command.go', line: 45, col: 6, is_decl: true, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/handlers/factory.go', line: 12, col: 10, is_decl: false, source: 'gopls' },
  { run_id: 44, commit_hash: 'abc1234', symbol_hash: 'a7b3c9f2', path: 'pkg/api/server.go', line: 87, col: 15, is_decl: false, source: 'gopls' },
]

export const mockCodeUnits: CodeUnit[] = [
  {
    unit_hash: 'cu_a1b2c3',
    kind: 'func',
    name: 'NewCommandProcessor',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    file: 'pkg/handlers/command.go',
    start_line: 67,
    start_col: 1,
    end_line: 85,
    end_col: 2,
    body_hash: 'body_hash_1',
    run_id: 44,
  },
  {
    unit_hash: 'cu_d4e5f6',
    kind: 'method',
    name: 'Process',
    recv: '*commandProcessorImpl',
    pkg: 'github.com/go-go-golems/glazed/pkg/handlers',
    file: 'pkg/handlers/command.go',
    start_line: 127,
    start_col: 1,
    end_line: 165,
    end_col: 2,
    body_hash: 'body_hash_2',
    run_id: 44,
  },
]

export const mockCommits: Commit[] = [
  {
    hash: 'abc1234',
    subject: 'Rename CommandProcessor to Handler',
    body: 'This change renames the CommandProcessor interface to Handler for consistency with the rest of the codebase.',
    author_name: 'Alice',
    author_email: 'alice@example.com',
    committer_date: '2026-02-05T06:00:00Z',
    run_id: 42,
  },
  {
    hash: 'def5678',
    subject: 'Fix middleware registration order',
    author_name: 'Bob',
    author_email: 'bob@example.com',
    committer_date: '2026-02-05T04:00:00Z',
    run_id: 42,
  },
  {
    hash: '789abcd',
    subject: 'Add context propagation to grpc layer',
    author_name: 'Alice',
    author_email: 'alice@example.com',
    committer_date: '2026-02-04T10:00:00Z',
    run_id: 42,
  },
]

export const mockCommitFiles: CommitFile[] = [
  { path: 'pkg/handlers/command.go', status: 'M' },
  { path: 'pkg/handlers/types.go', status: 'A' },
  { path: 'pkg/handlers/old_handler.go', status: 'D' },
]

export const mockDiffFiles: DiffFile[] = [
  {
    run_id: 44,
    path: 'pkg/handlers/command.go',
    status: 'M',
  },
  {
    run_id: 44,
    path: 'pkg/handlers/middleware.go',
    status: 'M',
  },
  {
    run_id: 44,
    path: 'pkg/handlers/types.go',
    status: 'A',
  },
]

export const mockDiffHunks: DiffHunk[] = [
  {
    id: 1,
    old_start: 45,
    old_lines: 8,
    new_start: 45,
    new_lines: 10,
    lines: [
      { kind: ' ', line_no_old: 45, line_no_new: 45, text: '// CommandProcessor handles command execution' },
      { kind: '-', line_no_old: 46, text: 'type Processor interface {' },
      { kind: '+', line_no_new: 46, text: 'type CommandProcessor interface {' },
      { kind: ' ', line_no_old: 47, line_no_new: 47, text: '\tProcess(ctx context.Context, cmd Command) (Result, error)' },
      { kind: '+', line_no_new: 48, text: '\tValidate(cmd Command) error' },
      { kind: ' ', line_no_old: 48, line_no_new: 49, text: '}' },
    ],
  },
]

export const mockDocTerms: DocTerm[] = [
  { term: 'Processor', count: 45 },
  { term: 'CommandProcessor', count: 23 },
  { term: 'handler', count: 89 },
  { term: 'middleware', count: 34 },
]

export const mockDocHits: DocHit[] = [
  { path: 'pkg/handlers/command.go', term: 'Processor', line: 45, col: 6, match_text: 'type CommandProcessor interface', run_id: 44 },
  { path: 'pkg/api/server.go', term: 'Processor', line: 87, col: 15, match_text: 'var proc CommandProcessor', run_id: 44 },
  { path: 'pkg/handlers/factory.go', term: 'Processor', line: 23, col: 12, match_text: 'func NewProcessor() CommandProcessor', run_id: 44 },
]

export const mockFileTree: FileEntry[] = [
  { path: 'cmd', kind: 'dir' },
  { path: 'pkg', kind: 'dir' },
  { path: 'internal', kind: 'dir' },
  { path: 'go.mod', kind: 'file' },
  { path: 'go.sum', kind: 'file' },
  { path: 'README.md', kind: 'file' },
  { path: 'Makefile', kind: 'file' },
]

export const mockSearchResults: SearchResult[] = [
  {
    type: 'symbol',
    primary: 'CommandProcessor',
    snippet: 'type CommandProcessor interface { ... }',
    path: 'pkg/handlers/command.go',
    line: 45,
    payload: mockSymbols[0],
  },
  {
    type: 'code_unit',
    primary: 'NewCommandProcessor',
    snippet: 'func NewCommandProcessor(opts ...Option) CommandProcessor',
    path: 'pkg/handlers/command.go',
    line: 67,
    payload: mockCodeUnits[0],
  },
  {
    type: 'diff',
    primary: '+type CommandProcessor interface',
    snippet: '+type CommandProcessor interface {',
    path: 'pkg/handlers/command.go',
    line: 46,
    payload: { kind: '+', text: 'type CommandProcessor interface {' },
  },
  {
    type: 'commit',
    primary: 'Rename CommandProcessor to Handler',
    snippet: 'This change renames the CommandProcessor...',
    commit_hash: 'abc1234',
    payload: mockCommits[0],
  },
  {
    type: 'doc',
    primary: 'CommandProcessor',
    snippet: '...the CommandProcessor interface handles...',
    path: 'docs/api.md',
    line: 23,
    payload: { term: 'CommandProcessor', path: 'docs/api.md', line: 23 },
  },
]
