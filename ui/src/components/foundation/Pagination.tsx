export interface PaginationProps {
  /** Total number of items (optional for endless pagination) */
  total?: number
  /** Items per page */
  limit: number
  /** Current offset */
  offset: number
  /** Called when page changes */
  onChange: (offset: number) => void
  /** Show total count display */
  showTotal?: boolean
  /** Compact mode (prev/next only) */
  compact?: boolean
  /** Custom class name */
  className?: string
}

export function Pagination({
  total,
  limit,
  offset,
  onChange,
  showTotal = true,
  compact = false,
  className = '',
}: PaginationProps) {
  const currentPage = Math.floor(offset / limit) + 1
  const totalPages = total ? Math.ceil(total / limit) : undefined
  const hasNextPage = total ? offset + limit < total : true
  const hasPrevPage = offset > 0

  const startItem = offset + 1
  const endItem = total ? Math.min(offset + limit, total) : offset + limit

  const handlePrev = () => {
    if (hasPrevPage) {
      onChange(Math.max(0, offset - limit))
    }
  }

  const handleNext = () => {
    if (hasNextPage) {
      onChange(offset + limit)
    }
  }

  const handlePage = (page: number) => {
    onChange((page - 1) * limit)
  }

  // Generate page numbers to show
  const getPageNumbers = (): (number | 'ellipsis')[] => {
    if (!totalPages || totalPages <= 7) {
      return Array.from({ length: totalPages || 1 }, (_, i) => i + 1)
    }

    const pages: (number | 'ellipsis')[] = []
    if (currentPage <= 4) {
      pages.push(1, 2, 3, 4, 5, 'ellipsis', totalPages)
    } else if (currentPage >= totalPages - 3) {
      pages.push(1, 'ellipsis', totalPages - 4, totalPages - 3, totalPages - 2, totalPages - 1, totalPages)
    } else {
      pages.push(1, 'ellipsis', currentPage - 1, currentPage, currentPage + 1, 'ellipsis', totalPages)
    }
    return pages
  }

  if (compact) {
    return (
      <nav className={`d-flex align-items-center gap-2 ${className}`} aria-label="Pagination">
        {showTotal && total !== undefined && (
          <span className="text-muted small">
            {startItem}-{endItem} of {total.toLocaleString()}
          </span>
        )}
        <div className="btn-group btn-group-sm">
          <button
            type="button"
            className="btn btn-outline-secondary"
            onClick={handlePrev}
            disabled={!hasPrevPage}
            aria-label="Previous page"
          >
            ‹
          </button>
          <button
            type="button"
            className="btn btn-outline-secondary"
            onClick={handleNext}
            disabled={!hasNextPage}
            aria-label="Next page"
          >
            ›
          </button>
        </div>
      </nav>
    )
  }

  const pageNumbers = getPageNumbers()

  return (
    <nav className={`d-flex align-items-center gap-3 ${className}`} aria-label="Pagination">
      {showTotal && total !== undefined && (
        <span className="text-muted small">
          {startItem.toLocaleString()}-{endItem.toLocaleString()} of {total.toLocaleString()}
        </span>
      )}
      <ul className="pagination pagination-sm mb-0">
        <li className={`page-item ${!hasPrevPage ? 'disabled' : ''}`}>
          <button
            type="button"
            className="page-link"
            onClick={handlePrev}
            disabled={!hasPrevPage}
            aria-label="Previous page"
          >
            ‹
          </button>
        </li>
        {pageNumbers.map((page, index) =>
          page === 'ellipsis' ? (
            <li key={`ellipsis-${index}`} className="page-item disabled">
              <span className="page-link">…</span>
            </li>
          ) : (
            <li key={page} className={`page-item ${page === currentPage ? 'active' : ''}`}>
              <button
                type="button"
                className="page-link"
                onClick={() => handlePage(page)}
                aria-label={`Page ${page}`}
                aria-current={page === currentPage ? 'page' : undefined}
              >
                {page}
              </button>
            </li>
          )
        )}
        <li className={`page-item ${!hasNextPage ? 'disabled' : ''}`}>
          <button
            type="button"
            className="page-link"
            onClick={handleNext}
            disabled={!hasNextPage}
            aria-label="Next page"
          >
            ›
          </button>
        </li>
      </ul>
    </nav>
  )
}
