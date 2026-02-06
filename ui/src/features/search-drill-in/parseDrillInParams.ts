import type {
  CodeUnitDrillInParams,
  CommitDrillInParams,
  DiffDrillInParams,
  DocDrillInParams,
  DrillInCommonParams,
  FileDrillInParams,
  SymbolDrillInParams,
} from './types'

function readString(params: URLSearchParams, key: string): string | undefined {
  const value = params.get(key)
  if (!value) return undefined
  const trimmed = value.trim()
  return trimmed === '' ? undefined : trimmed
}

function readNumber(params: URLSearchParams, key: string): number | undefined {
  const raw = readString(params, key)
  if (!raw) return undefined
  const parsed = Number(raw)
  return Number.isFinite(parsed) ? parsed : undefined
}

export function parseDrillInCommonParams(params: URLSearchParams): DrillInCommonParams {
  return {
    from: readString(params, 'from'),
    q: readString(params, 'q'),
    sessionId: readString(params, 'session_id'),
    runId: readNumber(params, 'run_id'),
  }
}

export function parseSymbolDrillInParams(params: URLSearchParams): SymbolDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    symbolHash: readString(params, 'symbol_hash'),
    path: readString(params, 'path'),
    line: readNumber(params, 'line'),
  }
}

export function parseCodeUnitDrillInParams(params: URLSearchParams): CodeUnitDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    unitHash: readString(params, 'unit_hash'),
    path: readString(params, 'path'),
    line: readNumber(params, 'line'),
  }
}

export function parseCommitDrillInParams(params: URLSearchParams): CommitDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    commitHash: readString(params, 'commit_hash'),
  }
}

export function parseDiffDrillInParams(params: URLSearchParams): DiffDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    path: readString(params, 'path'),
    lineNew: readNumber(params, 'line_new'),
    lineOld: readNumber(params, 'line_old'),
    hunkId: readNumber(params, 'hunk_id'),
  }
}

export function parseDocDrillInParams(params: URLSearchParams): DocDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    term: readString(params, 'term'),
    path: readString(params, 'path'),
    line: readNumber(params, 'line'),
    col: readNumber(params, 'col'),
  }
}

export function parseFileDrillInParams(params: URLSearchParams): FileDrillInParams {
  const common = parseDrillInCommonParams(params)
  return {
    ...common,
    path: readString(params, 'path'),
    line: readNumber(params, 'line'),
  }
}
