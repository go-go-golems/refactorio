import { useMemo } from 'react'
import { useGetSessionsQuery } from '../api/client'
import { useAppSelector, selectActiveWorkspaceId, selectActiveSessionId } from '../store'

export interface SessionRunIds {
  symbols?: number
  codeUnits?: number
  commits?: number
  diff?: number
  docs?: number
  goplsRefs?: number
  treeSitter?: number
}

export function useSessionContext() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const sessionId = useAppSelector(selectActiveSessionId)

  const { data: sessions, isLoading: sessionsLoading } = useGetSessionsQuery(workspaceId!, {
    skip: !workspaceId,
  })

  const activeSession = useMemo(() => {
    if (!sessions || sessions.length === 0) return undefined
    if (sessionId) {
      const match = sessions.find((session) => session.id === sessionId)
      if (match) return match
    }
    return sessions[0]
  }, [sessions, sessionId])

  const runIds = useMemo<SessionRunIds>(() => {
    const runs = activeSession?.runs
    return {
      symbols: runs?.symbols,
      codeUnits: runs?.code_units,
      commits: runs?.commits,
      diff: runs?.diff,
      docs: runs?.doc_hits,
      goplsRefs: runs?.gopls_refs,
      treeSitter: runs?.tree_sitter,
    }
  }, [activeSession])

  const searchRunIds = useMemo<Record<string, number>>(() => {
    const map: Record<string, number> = {}
    if (runIds.symbols) map.symbols = runIds.symbols
    if (runIds.codeUnits) map.code_units = runIds.codeUnits
    if (runIds.diff) map.diffs = runIds.diff
    if (runIds.commits) map.commits = runIds.commits
    if (runIds.docs) map.docs = runIds.docs
    return map
  }, [runIds])

  return {
    workspaceId,
    sessionId,
    sessions,
    sessionsLoading,
    activeSession,
    runIds,
    searchRunIds,
  }
}
