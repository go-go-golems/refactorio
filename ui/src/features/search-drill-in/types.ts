import type { SearchResult } from '../../types/api'

export interface SearchSymbolPayload {
  run_id: number
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
}

export interface SearchCodeUnitPayload {
  run_id: number
  unit_hash: string
  name: string
  kind: string
  pkg: string
  recv?: string
  signature?: string
  file: string
  start_line: number
  start_col: number
  end_line: number
  end_col: number
  body_text?: string
}

export interface SearchDiffPayload {
  run_id: number
  path: string
  kind: string
  line_no_old?: number
  line_no_new?: number
  text: string
  diff_file_id: number
  hunk_id: number
}

export interface SearchCommitPayload {
  run_id: number
  hash: string
  subject?: string
  body?: string
  author_name?: string
  author_email?: string
  author_date?: string
  committer_date?: string
}

export interface SearchDocPayload {
  run_id: number
  term: string
  path: string
  line: number
  col: number
  match_text: string
}

export interface SearchFilePayload {
  path: string
  ext?: string
  exists?: boolean
  is_binary?: boolean
}

export interface BuildSearchDrillInHrefArgs {
  result: SearchResult
  query?: string
  sessionId?: string
  source?: string
}

export interface DrillInCommonParams {
  from?: string
  q?: string
  sessionId?: string
  runId?: number
}

export interface SymbolDrillInParams extends DrillInCommonParams {
  symbolHash?: string
  path?: string
  line?: number
}

export interface CodeUnitDrillInParams extends DrillInCommonParams {
  unitHash?: string
  path?: string
  line?: number
}

export interface CommitDrillInParams extends DrillInCommonParams {
  commitHash?: string
}

export interface DiffDrillInParams extends DrillInCommonParams {
  path?: string
  lineNew?: number
  lineOld?: number
  hunkId?: number
}

export interface DocDrillInParams extends DrillInCommonParams {
  term?: string
  path?: string
  line?: number
  col?: number
}

export interface FileDrillInParams extends DrillInCommonParams {
  path?: string
  line?: number
}
