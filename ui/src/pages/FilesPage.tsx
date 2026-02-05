import { useState } from 'react'
import { useAppSelector, selectActiveWorkspaceId } from '../store'
import { useGetFilesQuery, useGetFileContentQuery } from '../api/client'
import { FileTree } from '../components/navigation/FileTree'
import { CodeViewer } from '../components/code-display/CodeViewer'
import type { FileEntry } from '../types/api'

export function FilesPage() {
  const workspaceId = useAppSelector(selectActiveWorkspaceId)
  const [selectedPath, setSelectedPath] = useState<string | undefined>(undefined)
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set())

  const { data: files, isLoading } = useGetFilesQuery(
    { workspace_id: workspaceId!, prefix: '' },
    { skip: !workspaceId },
  )

  const selectedEntry = files?.find((f) => f.path === selectedPath && f.kind === 'file')

  const { data: fileContent, isFetching: contentLoading } = useGetFileContentQuery(
    { workspace_id: workspaceId!, path: selectedPath ?? '' },
    { skip: !workspaceId || !selectedEntry },
  )

  if (!workspaceId) return <div className="p-4 text-muted">Select a workspace first.</div>

  return (
    <div className="d-flex h-100">
      <div style={{ width: 280, flexShrink: 0, borderRight: '1px solid var(--bs-border-color)' }} className="p-3 overflow-auto">
        <h6 className="mb-2">File Explorer</h6>
        <FileTree
          entries={files ?? []}
          selectedPath={selectedPath}
          expandedPaths={expandedPaths}
          onSelect={(entry: FileEntry) => {
            if (entry.kind === 'dir') {
              setExpandedPaths((prev) => {
                const next = new Set(prev)
                if (next.has(entry.path)) next.delete(entry.path)
                else next.add(entry.path)
                return next
              })
            } else {
              setSelectedPath(entry.path)
            }
          }}
          loading={isLoading}
        />
      </div>
      <div className="flex-grow-1 p-3 overflow-auto">
        {selectedEntry ? (
          <>
            <div className="bg-body-tertiary p-2 rounded mb-2">
              <span className="font-monospace small">{selectedEntry.path}</span>
            </div>
            {contentLoading ? (
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
