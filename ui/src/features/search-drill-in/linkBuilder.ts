import type { SearchResult } from '../../types/api'
import type {
  BuildSearchDrillInHrefArgs,
  SearchCodeUnitPayload,
  SearchCommitPayload,
  SearchDiffPayload,
  SearchDocPayload,
  SearchFilePayload,
  SearchSymbolPayload,
} from './types'

function asString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value.trim() : undefined
}

function asNumber(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const n = Number(value)
    if (Number.isFinite(n)) return n
  }
  return undefined
}

function setIfString(params: URLSearchParams, key: string, value: unknown) {
  const s = asString(value)
  if (s) params.set(key, s)
}

function setIfNumber(params: URLSearchParams, key: string, value: unknown) {
  const n = asNumber(value)
  if (n !== undefined) params.set(key, String(n))
}

function buildHref(path: string, params: URLSearchParams): string {
  const qs = params.toString()
  return qs ? `${path}?${qs}` : path
}

function applyCommonParams(params: URLSearchParams, query?: string, sessionId?: string, source = 'search') {
  setIfString(params, 'from', source)
  setIfString(params, 'q', query)
  setIfString(params, 'session_id', sessionId)
}

function buildSymbolHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchSymbolPayload> | undefined
  const symbolHash = asString(payload?.symbol_hash)
  if (!symbolHash) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('symbol_hash', symbolHash)
  setIfNumber(params, 'run_id', result.run_id ?? payload?.run_id)
  setIfString(params, 'path', result.path ?? payload?.file)
  setIfNumber(params, 'line', result.line ?? payload?.line)

  return buildHref('/symbols', params)
}

function buildCodeUnitHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchCodeUnitPayload> | undefined
  const unitHash = asString(payload?.unit_hash)
  if (!unitHash) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('unit_hash', unitHash)
  setIfNumber(params, 'run_id', result.run_id ?? payload?.run_id)
  setIfString(params, 'path', result.path ?? payload?.file)
  setIfNumber(params, 'line', result.line ?? payload?.start_line)

  return buildHref('/code-units', params)
}

function buildCommitHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchCommitPayload> | undefined
  const commitHash = asString(result.commit_hash ?? payload?.hash)
  if (!commitHash) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('commit_hash', commitHash)
  setIfNumber(params, 'run_id', result.run_id ?? payload?.run_id)

  return buildHref('/commits', params)
}

function buildDiffHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchDiffPayload> | undefined
  const runId = asNumber(result.run_id ?? payload?.run_id)
  const path = asString(result.path ?? payload?.path)
  if (runId === undefined || !path) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('run_id', String(runId))
  params.set('path', path)
  setIfNumber(params, 'line_new', result.line ?? payload?.line_no_new)
  setIfNumber(params, 'line_old', payload?.line_no_old)
  setIfNumber(params, 'hunk_id', payload?.hunk_id)

  return buildHref('/diffs', params)
}

function buildDocHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchDocPayload> | undefined
  const term = asString(payload?.term ?? result.primary)
  if (!term) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('term', term)
  setIfNumber(params, 'run_id', result.run_id ?? payload?.run_id)
  setIfString(params, 'path', result.path ?? payload?.path)
  setIfNumber(params, 'line', result.line ?? payload?.line)
  setIfNumber(params, 'col', result.col ?? payload?.col)

  return buildHref('/docs', params)
}

function buildFileHref(result: SearchResult, query?: string, sessionId?: string, source?: string): string | null {
  const payload = result.payload as Partial<SearchFilePayload> | undefined
  const path = asString(result.path ?? payload?.path ?? result.primary)
  if (!path) return null

  const params = new URLSearchParams()
  applyCommonParams(params, query, sessionId, source)
  params.set('path', path)
  setIfNumber(params, 'line', result.line)

  return buildHref('/files', params)
}

export function buildSearchDrillInHref({ result, query, sessionId, source = 'search' }: BuildSearchDrillInHrefArgs): string | null {
  switch (result.type) {
    case 'symbol':
      return buildSymbolHref(result, query, sessionId, source)
    case 'code_unit':
      return buildCodeUnitHref(result, query, sessionId, source)
    case 'commit':
      return buildCommitHref(result, query, sessionId, source)
    case 'diff':
      return buildDiffHref(result, query, sessionId, source)
    case 'doc':
      return buildDocHref(result, query, sessionId, source)
    case 'file':
      return buildFileHref(result, query, sessionId, source)
    default:
      return null
  }
}
