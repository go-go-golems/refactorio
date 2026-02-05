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
  schema_version: 12,
  tables: [
    'meta_runs',
    'symbols',
    'symbol_occurrences',
    'symbol_refs',
    'code_unit_snapshots',
    'commits',
    'commit_files',
    'diff_files',
    'diff_hunks',
    'diff_lines',
    'doc_hits',
    'files',
  ],
  fts_tables: [
    'symbols_fts',
    'code_units_fts',
    'commits_fts',
    'diff_lines_fts',
    'doc_hits_fts',
  ],
  features: {
    gopls_refs: true,
    tree_sitter: false,
    doc_hits: true,
    code_units: true,
  },
  row_counts: {
    symbols: 12456,
    code_unit_snapshots: 8234,
    commits: 1847,
    diff_files: 3291,
    doc_hits: 847,
  },
}

export const mockRuns: Run[] = [
  {
    run_id: 44,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T08:00:00Z',
    finished_at: '2026-02-05T08:05:00Z',
  },
  {
    run_id: 43,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T07:55:00Z',
    finished_at: '2026-02-05T08:00:00Z',
  },
  {
    run_id: 42,
    status: 'success',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    started_at: '2026-02-05T07:50:00Z',
    finished_at: '2026-02-05T07:55:00Z',
  },
  {
    run_id: 41,
    status: 'failed',
    root_path: '/Users/dev/src/glazed',
    git_from: 'HEAD~20',
    git_to: 'HEAD',
    error_json: '{"message": "tree-sitter parser failed"}',
    started_at: '2026-02-05T07:45:00Z',
    finished_at: '2026-02-05T07:46:00Z',
  },
]

export const mockRunSummary: RunSummary = {
  run_id: 44,
  symbols_count: 234,
  code_units_count: 189,
  commits_count: 47,
  diff_files_count: 23,
  diff_lines_count: 891,
  doc_hits_count: 45,
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
      tree_sitter: false,
    },
    last_updated: '2026-02-05T09:00:00Z',
  },
]

export const mockSymbols: Symbol[] = [
  {
    symbol_hash: 'a7b3c9f2',
    name: 'CommandProcessor',
    kind: 'type',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'type CommandProcessor interface { Process(ctx context.Context, cmd Command) (Result, error) }',
    exported: true,
    file_path: 'pkg/handlers/command.go',
    start_line: 45,
    start_col: 1,
    end_line: 52,
    end_col: 2,
    run_id: 44,
  },
  {
    symbol_hash: 'b8c4d0e3',
    name: 'NewCommandProcessor',
    kind: 'func',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'func NewCommandProcessor(opts ...Option) CommandProcessor',
    exported: true,
    file_path: 'pkg/handlers/command.go',
    start_line: 67,
    start_col: 1,
    end_line: 85,
    end_col: 2,
    run_id: 44,
  },
  {
    symbol_hash: 'c9d5e1f4',
    name: 'commandProcessorImpl',
    kind: 'type',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'type commandProcessorImpl struct { ... }',
    exported: false,
    file_path: 'pkg/handlers/command.go',
    start_line: 89,
    start_col: 1,
    end_line: 95,
    end_col: 2,
    run_id: 44,
  },
  {
    symbol_hash: 'd0e6f2g5',
    name: 'Process',
    kind: 'method',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    signature: 'func (p *commandProcessorImpl) Process(ctx context.Context, cmd Command) (Result, error)',
    exported: true,
    file_path: 'pkg/handlers/command.go',
    start_line: 127,
    start_col: 1,
    end_line: 165,
    end_col: 2,
    run_id: 44,
  },
]

export const mockSymbolRefs: SymbolRef[] = [
  { file_path: 'pkg/handlers/command.go', start_line: 45, start_col: 6, end_line: 45, end_col: 22, is_declaration: true },
  { file_path: 'pkg/handlers/factory.go', start_line: 12, start_col: 10, end_line: 12, end_col: 26, is_declaration: false },
  { file_path: 'pkg/api/server.go', start_line: 87, start_col: 15, end_line: 87, end_col: 31, is_declaration: false },
]

export const mockCodeUnits: CodeUnit[] = [
  {
    code_unit_hash: 'cu_a1b2c3',
    kind: 'func',
    name: 'NewCommandProcessor',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    file_path: 'pkg/handlers/command.go',
    start_line: 67,
    start_col: 1,
    end_line: 85,
    end_col: 2,
    body_hash: 'body_hash_1',
    run_id: 44,
  },
  {
    code_unit_hash: 'cu_d4e5f6',
    kind: 'method',
    name: 'Process',
    receiver: '*commandProcessorImpl',
    package_path: 'github.com/go-go-golems/glazed/pkg/handlers',
    file_path: 'pkg/handlers/command.go',
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
    commit_hash: 'abc1234',
    subject: 'Rename CommandProcessor to Handler',
    body: 'This change renames the CommandProcessor interface to Handler for consistency with the rest of the codebase.',
    author_name: 'Alice',
    author_email: 'alice@example.com',
    commit_date: '2026-02-05T06:00:00Z',
    run_id: 42,
  },
  {
    commit_hash: 'def5678',
    subject: 'Fix middleware registration order',
    author_name: 'Bob',
    author_email: 'bob@example.com',
    commit_date: '2026-02-05T04:00:00Z',
    run_id: 42,
  },
  {
    commit_hash: '789abcd',
    subject: 'Add context propagation to grpc layer',
    author_name: 'Alice',
    author_email: 'alice@example.com',
    commit_date: '2026-02-04T10:00:00Z',
    run_id: 42,
  },
]

export const mockCommitFiles: CommitFile[] = [
  { file_path: 'pkg/handlers/command.go', status: 'M', additions: 15, deletions: 8 },
  { file_path: 'pkg/handlers/types.go', status: 'A', additions: 45, deletions: 0 },
  { file_path: 'pkg/handlers/old_handler.go', status: 'D', additions: 0, deletions: 32 },
]

export const mockDiffFiles: DiffFile[] = [
  {
    file_path: 'pkg/handlers/command.go',
    status: 'M',
    hunks_count: 3,
    additions: 45,
    deletions: 23,
  },
  {
    file_path: 'pkg/handlers/middleware.go',
    status: 'M',
    hunks_count: 1,
    additions: 12,
    deletions: 5,
  },
  {
    file_path: 'pkg/handlers/types.go',
    status: 'A',
    hunks_count: 1,
    additions: 89,
    deletions: 0,
  },
]

export const mockDiffHunks: DiffHunk[] = [
  {
    hunk_id: 1,
    old_start: 45,
    old_count: 8,
    new_start: 45,
    new_count: 10,
    lines: [
      { kind: ' ', old_line: 45, new_line: 45, content: '// CommandProcessor handles command execution' },
      { kind: '-', old_line: 46, content: 'type Processor interface {' },
      { kind: '+', new_line: 46, content: 'type CommandProcessor interface {' },
      { kind: ' ', old_line: 47, new_line: 47, content: '\tProcess(ctx context.Context, cmd Command) (Result, error)' },
      { kind: '+', new_line: 48, content: '\tValidate(cmd Command) error' },
      { kind: ' ', old_line: 48, new_line: 49, content: '}' },
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
  { file_path: 'pkg/handlers/command.go', term: 'Processor', line: 45, col: 6, match_text: 'type CommandProcessor interface', run_id: 44 },
  { file_path: 'pkg/api/server.go', term: 'Processor', line: 87, col: 15, match_text: 'var proc CommandProcessor', run_id: 44 },
  { file_path: 'pkg/handlers/factory.go', term: 'Processor', line: 23, col: 12, match_text: 'func NewProcessor() CommandProcessor', run_id: 44 },
]

export const mockFileTree: FileEntry[] = [
  { path: 'cmd', is_dir: true, children_count: 3 },
  { path: 'pkg', is_dir: true, children_count: 8 },
  { path: 'internal', is_dir: true, children_count: 4 },
  { path: 'go.mod', is_dir: false },
  { path: 'go.sum', is_dir: false },
  { path: 'README.md', is_dir: false },
  { path: 'Makefile', is_dir: false },
]

export const mockSearchResults: SearchResult[] = [
  {
    type: 'symbol',
    id: 'a7b3c9f2',
    label: 'CommandProcessor',
    snippet: 'type CommandProcessor interface { ... }',
    location: 'pkg/handlers/command.go:45',
    payload: mockSymbols[0],
  },
  {
    type: 'code_unit',
    id: 'cu_a1b2c3',
    label: 'NewCommandProcessor',
    snippet: 'func NewCommandProcessor(opts ...Option) CommandProcessor',
    location: 'pkg/handlers/command.go:67',
    payload: mockCodeUnits[0],
  },
  {
    type: 'diff',
    id: 'diff_1',
    label: '+type CommandProcessor interface',
    snippet: '+type CommandProcessor interface {',
    location: 'pkg/handlers/command.go:46',
    payload: { kind: '+', content: 'type CommandProcessor interface {' },
  },
  {
    type: 'commit',
    id: 'abc1234',
    label: 'Rename CommandProcessor to Handler',
    snippet: 'This change renames the CommandProcessor...',
    location: 'abc1234 by Alice',
    payload: mockCommits[0],
  },
  {
    type: 'doc',
    id: 'doc_1',
    label: 'CommandProcessor',
    snippet: '...the CommandProcessor interface handles...',
    location: 'docs/api.md:23',
    payload: { term: 'CommandProcessor', file: 'docs/api.md', line: 23 },
  },
]
