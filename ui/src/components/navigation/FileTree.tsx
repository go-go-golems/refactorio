import { useState, useCallback } from 'react'
import type { FileEntry } from '../../types/api'

export interface FileTreeProps {
  /** File/directory entries at current level */
  entries: FileEntry[]
  /** Currently selected file path */
  selectedPath?: string
  /** Set of expanded directory paths */
  expandedPaths?: Set<string>
  /** Called when a file or directory is selected */
  onSelect?: (entry: FileEntry) => void
  /** Called when a directory is expanded (for lazy loading) */
  onExpand?: (path: string) => void
  /** Called when a directory is collapsed */
  onCollapse?: (path: string) => void
  /** Whether the tree is loading */
  loading?: boolean
  /** Optional badges (e.g. diff count) keyed by path */
  badges?: Record<string, string | number>
  /** Nested children keyed by parent path */
  childrenMap?: Record<string, FileEntry[]>
}

const FolderIcon = ({ open }: { open: boolean }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" style={{ flexShrink: 0 }}>
    {open ? (
      <path d="M.54 3.87L.5 3a2 2 0 012-2h3.672a2 2 0 011.414.586l.828.828A2 2 0 009.828 3H13.5a2 2 0 011.95 1.555L.54 3.87zM1.059 5.56L2 14h12l1-8.5H1.059z" />
    ) : (
      <path d="M1 3.5A1.5 1.5 0 012.5 2h3.879a1.5 1.5 0 011.06.44l1.122 1.12A1.5 1.5 0 009.62 4H13.5A1.5 1.5 0 0115 5.5v7a1.5 1.5 0 01-1.5 1.5h-11A1.5 1.5 0 011 12.5v-9z" />
    )}
  </svg>
)

const FileIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" style={{ flexShrink: 0 }}>
    <path d="M3.75 1.5a.25.25 0 00-.25.25v12.5c0 .138.112.25.25.25h8.5a.25.25 0 00.25-.25V6H9.75A1.75 1.75 0 018 4.25V1.5H3.75zm5.75.56v2.19c0 .138.112.25.25.25h2.19L9.5 2.06zM2 1.75C2 .784 2.784 0 3.75 0h5.086c.464 0 .909.184 1.237.513l3.414 3.414c.329.328.513.773.513 1.237v8.086A1.75 1.75 0 0112.25 15h-8.5A1.75 1.75 0 012 13.25V1.75z" />
  </svg>
)

const ChevronIcon = ({ expanded }: { expanded: boolean }) => (
  <svg
    width="12"
    height="12"
    viewBox="0 0 12 12"
    fill="currentColor"
    style={{
      flexShrink: 0,
      transition: 'transform 0.15s',
      transform: expanded ? 'rotate(90deg)' : 'rotate(0deg)',
    }}
  >
    <path d="M4.5 2l4 4-4 4" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

function TreeNode({
  entry,
  depth,
  selectedPath,
  expandedPaths,
  onSelect,
  onExpand,
  onCollapse,
  badges,
  childrenMap,
}: {
  entry: FileEntry
  depth: number
  selectedPath?: string
  expandedPaths: Set<string>
  onSelect?: (entry: FileEntry) => void
  onExpand?: (path: string) => void
  onCollapse?: (path: string) => void
  badges?: Record<string, string | number>
  childrenMap?: Record<string, FileEntry[]>
}) {
  const isExpanded = expandedPaths.has(entry.path)
  const isSelected = selectedPath === entry.path
  const badge = badges?.[entry.path]
  const children = childrenMap?.[entry.path]

  const handleClick = useCallback(() => {
    onSelect?.(entry)
    if (entry.is_dir) {
      if (isExpanded) {
        onCollapse?.(entry.path)
      } else {
        onExpand?.(entry.path)
      }
    }
  }, [entry, isExpanded, onSelect, onExpand, onCollapse])

  const name = entry.path.includes('/') ? entry.path.split('/').pop()! : entry.path

  return (
    <>
      <button
        type="button"
        className={`file-tree-node btn btn-sm w-100 text-start d-flex align-items-center gap-1 rounded-0 border-0 ${
          isSelected ? 'bg-primary bg-opacity-10 text-primary' : ''
        }`}
        style={{
          paddingLeft: `${depth * 16 + 4}px`,
          paddingTop: '2px',
          paddingBottom: '2px',
          minHeight: '28px',
        }}
        onClick={handleClick}
        title={entry.path}
      >
        {entry.is_dir ? (
          <>
            <ChevronIcon expanded={isExpanded} />
            <FolderIcon open={isExpanded} />
          </>
        ) : (
          <>
            <span style={{ width: 12 }} />
            <FileIcon />
          </>
        )}
        <span className="text-truncate flex-grow-1" style={{ fontSize: '0.85rem' }}>
          {name}
        </span>
        {entry.is_dir && entry.children_count != null && !isExpanded && (
          <span className="text-muted" style={{ fontSize: '0.7rem' }}>
            {entry.children_count}
          </span>
        )}
        {badge != null && (
          <span className="badge bg-secondary rounded-pill" style={{ fontSize: '0.65rem' }}>
            {badge}
          </span>
        )}
      </button>
      {entry.is_dir && isExpanded && children && (
        <div role="group">
          {children.map((child) => (
            <TreeNode
              key={child.path}
              entry={child}
              depth={depth + 1}
              selectedPath={selectedPath}
              expandedPaths={expandedPaths}
              onSelect={onSelect}
              onExpand={onExpand}
              onCollapse={onCollapse}
              badges={badges}
              childrenMap={childrenMap}
            />
          ))}
        </div>
      )}
    </>
  )
}

function SkeletonTree() {
  return (
    <div className="placeholder-glow p-2">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="d-flex align-items-center gap-2 mb-1" style={{ paddingLeft: `${(i % 3) * 16}px` }}>
          <span className="placeholder col-1" style={{ height: 16, width: 16 }} />
          <span className={`placeholder col-${3 + (i % 4)}`} style={{ height: 14 }} />
        </div>
      ))}
    </div>
  )
}

export function FileTree({
  entries,
  selectedPath,
  expandedPaths: controlledExpanded,
  onSelect,
  onExpand,
  onCollapse,
  loading = false,
  badges,
  childrenMap,
}: FileTreeProps) {
  const [internalExpanded, setInternalExpanded] = useState<Set<string>>(new Set())
  const expandedPaths = controlledExpanded ?? internalExpanded

  const handleExpand = useCallback(
    (path: string) => {
      if (controlledExpanded) {
        onExpand?.(path)
      } else {
        setInternalExpanded((prev) => new Set([...prev, path]))
        onExpand?.(path)
      }
    },
    [controlledExpanded, onExpand],
  )

  const handleCollapse = useCallback(
    (path: string) => {
      if (controlledExpanded) {
        onCollapse?.(path)
      } else {
        setInternalExpanded((prev) => {
          const next = new Set(prev)
          next.delete(path)
          return next
        })
        onCollapse?.(path)
      }
    },
    [controlledExpanded, onCollapse],
  )

  if (loading) {
    return <SkeletonTree />
  }

  if (entries.length === 0) {
    return (
      <div className="text-muted text-center p-3" style={{ fontSize: '0.85rem' }}>
        No files found
      </div>
    )
  }

  return (
    <div className="file-tree" role="tree">
      {entries.map((entry) => (
        <TreeNode
          key={entry.path}
          entry={entry}
          depth={0}
          selectedPath={selectedPath}
          expandedPaths={expandedPaths}
          onSelect={onSelect}
          onExpand={handleExpand}
          onCollapse={handleCollapse}
          badges={badges}
          childrenMap={childrenMap}
        />
      ))}
      <style>{`
        .file-tree-node:hover {
          background-color: var(--bs-tertiary-bg) !important;
        }
      `}</style>
    </div>
  )
}
