export interface HighlightRange {
  startLine: number
  endLine: number
  startCol?: number
  endCol?: number
  className?: string
}

export interface CodeViewerProps {
  /** Code content */
  content: string
  /** Programming language (for future syntax highlighting) */
  language?: string
  /** Starting line number */
  startLine?: number
  /** Lines to highlight */
  highlightLines?: number[]
  /** Ranges to highlight */
  highlightRanges?: HighlightRange[]
  /** Called when a line is clicked */
  onLineClick?: (line: number) => void
  /** Show line numbers */
  showLineNumbers?: boolean
  /** Wrap long lines */
  wrapLines?: boolean
  /** Maximum height (CSS value) */
  maxHeight?: string
  /** Custom class name */
  className?: string
}

export function CodeViewer({
  content,
  language,
  startLine = 1,
  highlightLines = [],
  highlightRanges = [],
  onLineClick,
  showLineNumbers = true,
  wrapLines = false,
  maxHeight,
  className = '',
}: CodeViewerProps) {
  const lines = content.split('\n')
  const highlightSet = new Set(highlightLines)

  const isLineHighlighted = (lineNum: number) => {
    if (highlightSet.has(lineNum)) return true
    return highlightRanges.some((r) => lineNum >= r.startLine && lineNum <= r.endLine)
  }

  const getLineHighlightClass = (lineNum: number) => {
    const range = highlightRanges.find((r) => lineNum >= r.startLine && lineNum <= r.endLine)
    return range?.className || 'highlight-default'
  }

  return (
    <div
      className={`code-viewer ${className}`}
      style={{ maxHeight, overflowY: maxHeight ? 'auto' : undefined }}
      data-language={language}
    >
      <pre className="m-0">
        <code>
          {lines.map((line, index) => {
            const lineNum = startLine + index
            const highlighted = isLineHighlighted(lineNum)
            const highlightClass = highlighted ? getLineHighlightClass(lineNum) : ''

            return (
              <div
                key={lineNum}
                className={`code-line ${highlighted ? highlightClass : ''} ${onLineClick ? 'clickable' : ''}`}
                onClick={() => onLineClick?.(lineNum)}
              >
                {showLineNumbers && (
                  <span className="line-number">{lineNum}</span>
                )}
                <span
                  className="line-content"
                  style={{ whiteSpace: wrapLines ? 'pre-wrap' : 'pre' }}
                >
                  {line || '\n'}
                </span>
              </div>
            )
          })}
        </code>
      </pre>

      <style>{`
        .code-viewer {
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
          font-size: 13px;
          line-height: 1.5;
          background: var(--bs-body-bg);
          border: 1px solid var(--bs-border-color);
          border-radius: 0.375rem;
          overflow-x: auto;
        }
        .code-viewer pre {
          margin: 0;
          padding: 0;
        }
        .code-viewer code {
          display: block;
        }
        .code-line {
          display: flex;
          padding: 0 0.5rem;
        }
        .code-line.clickable {
          cursor: pointer;
        }
        .code-line.clickable:hover {
          background: var(--bs-tertiary-bg);
        }
        .code-line.highlight-default {
          background: rgba(255, 255, 0, 0.15);
        }
        .code-line.highlight-add {
          background: rgba(0, 255, 0, 0.1);
        }
        .code-line.highlight-remove {
          background: rgba(255, 0, 0, 0.1);
        }
        .code-line.highlight-focus {
          background: rgba(0, 100, 255, 0.15);
        }
        .line-number {
          display: inline-block;
          min-width: 3rem;
          padding-right: 1rem;
          text-align: right;
          color: var(--bs-secondary);
          user-select: none;
          flex-shrink: 0;
        }
        .line-content {
          flex-grow: 1;
        }
      `}</style>
    </div>
  )
}
