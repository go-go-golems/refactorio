export interface BreadcrumbItem {
  label: string
  path?: string
}

export interface BreadcrumbProps {
  /** Path segments */
  items: BreadcrumbItem[]
  /** Called when a segment is clicked */
  onNavigate?: (path: string) => void
  /** Maximum visible items before truncation */
  maxItems?: number
}

export function Breadcrumb({ items, onNavigate, maxItems }: BreadcrumbProps) {
  let displayItems = items
  let truncated = false

  if (maxItems && items.length > maxItems) {
    // Show first item, ellipsis, then last (maxItems - 2) items
    const tail = items.slice(-(maxItems - 1))
    displayItems = [items[0], { label: '\u2026' }, ...tail]
    truncated = true
  }

  return (
    <nav aria-label="breadcrumb">
      <ol className="breadcrumb mb-0" style={{ fontSize: '0.85rem' }}>
        {displayItems.map((item, i) => {
          const isLast = i === displayItems.length - 1
          const isEllipsis = truncated && i === 1

          if (isEllipsis) {
            return (
              <li key="ellipsis" className="breadcrumb-item text-muted">
                {'\u2026'}
              </li>
            )
          }

          if (isLast || !item.path) {
            return (
              <li key={item.path ?? i} className={`breadcrumb-item ${isLast ? 'active' : ''}`} aria-current={isLast ? 'page' : undefined}>
                {item.label}
              </li>
            )
          }

          return (
            <li key={item.path} className="breadcrumb-item">
              <button
                type="button"
                className="btn btn-link p-0 text-decoration-none"
                style={{ fontSize: 'inherit' }}
                onClick={() => item.path && onNavigate?.(item.path)}
              >
                {item.label}
              </button>
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
