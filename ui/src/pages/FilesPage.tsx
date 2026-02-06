import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetFilesQuery, useLazyGetFilesQuery, useGetFileContentQuery } from '../api/client'
import { FileTree } from '../components/navigation/FileTree'
import { CodeViewer } from '../components/code-display/CodeViewer'
import type { FileEntry } from '../types/api'
import { parseFileDrillInParams } from '../features/search-drill-in'

function parentPrefixes(path: string): string[] {
  const segments = path.split('/').filter(Boolean)
  if (segments.length <= 1) return []
  const dirs = segments.slice(0, -1)
  return dirs.map((_, index) => dirs.slice(0, index + 1).join('/'))
}

export function FilesPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [searchParams] = useSearchParams()
  const drillIn = parseFileDrillInParams(searchParams)
  const [selectedPath, setSelectedPath] = useState<string | undefined>(undefined)
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set())
  const [childrenMap, setChildrenMap] = useState<Record<string, FileEntry[]>>({})
  const [loadError, setLoadError] = useState<string | null>(null)
  const loadedPrefixesRef = useRef<Set<string>>(new Set())

  const { data: rootEntries, isLoading: rootLoading } = useGetFilesQuery(
    { workspace_id: workspaceId!, prefix: '' },
    { skip: !workspaceId },
  )
  const [fetchFiles, { isFetching: loadingPrefix }] = useLazyGetFilesQuery()

  useEffect(() => {
    if (!rootEntries) return
    setChildrenMap((prev) => ({ ...prev, '': rootEntries }))
    loadedPrefixesRef.current.add('')
  }, [rootEntries])

  const loadPrefix = useCallback(
    async (prefix: string) => {
      if (!workspaceId) return
      if (loadedPrefixesRef.current.has(prefix)) return

      try {
        const entries = await fetchFiles({ workspace_id: workspaceId, prefix }).unwrap()
        setChildrenMap((prev) => ({ ...prev, [prefix]: entries }))
        loadedPrefixesRef.current.add(prefix)
        setLoadError(null)
      } catch {
        setLoadError(`Unable to load directory prefix: ${prefix}`)
      }
    },
    [workspaceId, fetchFiles],
  )

  useEffect(() => {
    if (!workspaceId || !drillIn.path) return
    let cancelled = false

    const hydratePath = async () => {
      const prefixes = parentPrefixes(drillIn.path!)
      for (const prefix of prefixes) {
        if (cancelled) return
        await loadPrefix(prefix)
        if (cancelled) return
        setExpandedPaths((prev) => {
          if (prev.has(prefix)) return prev
          const next = new Set(prev)
          next.add(prefix)
          return next
        })
      }
      if (!cancelled) {
        setSelectedPath(drillIn.path)
      }
    }

    void hydratePath()
    return () => {
      cancelled = true
    }
  }, [workspaceId, drillIn.path, loadPrefix])

  const entriesByPath = useMemo(() => {
    const map: Record<string, FileEntry> = {}
    for (const entries of Object.values(childrenMap)) {
      for (const entry of entries) {
        map[entry.path] = entry
      }
    }
    return map
  }, [childrenMap])

  const selectedEntry = selectedPath ? entriesByPath[selectedPath] : undefined
  const selectedIsDirectory = selectedEntry?.kind === 'dir'

  const { data: fileContent, isFetching: contentLoading, isError: contentMissing } = useGetFileContentQuery(
    { workspace_id: workspaceId!, path: selectedPath ?? '' },
    { skip: !workspaceId || !selectedPath || selectedIsDirectory },
  )

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div style={{ width: 280, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
        <h6 className="mb-2">File Explorer</h6>
        <FileTree
          entries={childrenMap[''] ?? []}
          selectedPath={selectedPath}
          expandedPaths={expandedPaths}
          childrenMap={childrenMap}
          onSelect={(entry: FileEntry) => {
            if (entry.kind === 'dir') {
              setExpandedPaths((prev) => {
                const next = new Set(prev)
                if (next.has(entry.path)) {
                  next.delete(entry.path)
                } else {
                  next.add(entry.path)
                  void loadPrefix(entry.path)
                }
                return next
              })
            } else {
              setSelectedPath(entry.path)
            }
          }}
          loading={rootLoading && !childrenMap['']}
        />
      </div>
      <div className="flex-grow-1 p-3 overflow-auto">
        {loadError && (
          <div className="alert alert-warning py-2">{loadError}</div>
        )}
        {drillIn.path && contentMissing && (
          <div className="alert alert-warning py-2">
            Target file <code>{drillIn.path}</code> could not be loaded in the current scope.
          </div>
        )}
        {selectedPath ? (
          <>
            <div className="bg-body-tertiary p-2 rounded mb-2 d-flex justify-content-between align-items-center">
              <span className="font-monospace small">{selectedPath}</span>
              {drillIn.line && selectedPath === drillIn.path && (
                <span className="badge bg-secondary-subtle text-secondary">line {drillIn.line}</span>
              )}
            </div>
            {contentLoading || loadingPrefix ? (
              <div className="placeholder-glow p-3">
                {Array.from({ length: 10 }).map((_, i) => (
                  <div key={i} className="placeholder col-12 mb-1" style={{ height: 16 }} />
                ))}
              </div>
            ) : fileContent ? (
              <CodeViewer content={fileContent.content} />
            ) : (
              <div className="text-muted p-4">Unable to load file content</div>
            )}
          </>
        ) : (
          <div className="text-muted p-4 text-center">Select a file to view its content</div>
        )}
      </div>
    </div>
  )
}
