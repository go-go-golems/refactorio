import type { DiffHunk, DiffLine } from '../../types/api'

export interface DiffViewerProps {
  /** Diff hunks to display */
  hunks: DiffHunk[]
  /** Display mode */
  mode?: 'unified' | 'split'
  /** Show line numbers */
  showLineNumbers?: boolean
  /** Number of context lines to show around changes */
  contextLines?: number
  /** Called when a line is clicked */
  onLineClick?: (line: DiffLine) => void
  /** Search query to highlight */
  highlightQuery?: string
  /** Custom class name */
  className?: string
}

function highlightMatch(text: string, query?: string): React.ReactNode {
  if (!query || !text) return text
  const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  const parts = text.split(regex)
  return parts.map((part, i) =>
    regex.test(part) ? (
      <mark key={i} className="bg-warning px-0">
        {part}
      </mark>
    ) : (
      part
    )
  )
}

function DiffLineRow({
  line,
  showLineNumbers,
  onClick,
  highlightQuery,
}: {
  line: DiffLine
  showLineNumbers: boolean
  onClick?: () => void
  highlightQuery?: string
}) {
  const kindClass = {
    '+': 'diff-add',
    '-': 'diff-remove',
    ' ': 'diff-context',
  }[line.kind]

  const kindSymbol = {
    '+': '+',
    '-': '-',
    ' ': ' ',
  }[line.kind]

  return (
    <div
      className={`diff-line ${kindClass} ${onClick ? 'clickable' : ''}`}
      onClick={onClick}
    >
      {showLineNumbers && (
        <>
          <span className="line-number old">
            {line.old_line ?? ''}
          </span>
          <span className="line-number new">
            {line.new_line ?? ''}
          </span>
        </>
      )}
      <span className="diff-symbol">{kindSymbol}</span>
      <span className="line-content">
        {highlightQuery ? highlightMatch(line.content, highlightQuery) : line.content}
      </span>
    </div>
  )
}

function HunkHeader({ hunk }: { hunk: DiffHunk }) {
  return (
    <div className="hunk-header">
      @@ -{hunk.old_start},{hunk.old_count} +{hunk.new_start},{hunk.new_count} @@
    </div>
  )
}

export function DiffViewer({
  hunks,
  mode = 'unified',
  showLineNumbers = true,
  onLineClick,
  highlightQuery,
  className = '',
}: DiffViewerProps) {
  if (hunks.length === 0) {
    return (
      <div className={`diff-viewer empty text-center text-muted py-4 ${className}`}>
        No diff content
      </div>
    )
  }

  // For now, only unified mode is implemented
  return (
    <div className={`diff-viewer ${className}`}>
      <pre className="m-0">
        <code>
          {hunks.map((hunk, hunkIndex) => (
            <div key={hunk.hunk_id || hunkIndex} className="diff-hunk">
              <HunkHeader hunk={hunk} />
              {hunk.lines.map((line, lineIndex) => (
                <DiffLineRow
                  key={lineIndex}
                  line={line}
                  showLineNumbers={showLineNumbers}
                  onClick={onLineClick ? () => onLineClick(line) : undefined}
                  highlightQuery={highlightQuery}
                />
              ))}
            </div>
          ))}
        </code>
      </pre>

      <style>{`
        .diff-viewer {
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
          font-size: 13px;
          line-height: 1.5;
          background: var(--bs-body-bg);
          border: 1px solid var(--bs-border-color);
          border-radius: 0.375rem;
          overflow-x: auto;
        }
        .diff-viewer pre {
          margin: 0;
          padding: 0;
        }
        .diff-viewer code {
          display: block;
        }
        .diff-hunk {
          margin-bottom: 0.5rem;
        }
        .diff-hunk:last-child {
          margin-bottom: 0;
        }
        .hunk-header {
          background: var(--bs-tertiary-bg);
          color: var(--bs-secondary);
          padding: 0.25rem 0.5rem;
          font-size: 0.85em;
          border-bottom: 1px solid var(--bs-border-color);
        }
        .diff-line {
          display: flex;
          padding: 0 0.5rem;
        }
        .diff-line.clickable {
          cursor: pointer;
        }
        .diff-line.clickable:hover {
          filter: brightness(0.95);
        }
        .diff-line.diff-add {
          background: #e6ffec;
        }
        .diff-line.diff-remove {
          background: #ffebe9;
        }
        .diff-line.diff-context {
          background: transparent;
        }
        .line-number {
          display: inline-block;
          min-width: 2.5rem;
          padding-right: 0.5rem;
          text-align: right;
          color: var(--bs-secondary);
          user-select: none;
          flex-shrink: 0;
        }
        .line-number.old {
          border-right: 1px solid var(--bs-border-color);
        }
        .line-number.new {
          margin-right: 0.25rem;
        }
        .diff-symbol {
          display: inline-block;
          width: 1.5rem;
          text-align: center;
          flex-shrink: 0;
          font-weight: bold;
        }
        .diff-add .diff-symbol {
          color: #1a7f37;
        }
        .diff-remove .diff-symbol {
          color: #cf222e;
        }
        .line-content {
          flex-grow: 1;
          white-space: pre;
        }

        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
          .diff-line.diff-add {
            background: rgba(46, 160, 67, 0.15);
          }
          .diff-line.diff-remove {
            background: rgba(248, 81, 73, 0.15);
          }
        }
      `}</style>
    </div>
  )
}
