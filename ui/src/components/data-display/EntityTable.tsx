import { Pagination, type PaginationProps } from '../foundation'

export interface Column<T> {
  /** Unique key for the column */
  key: string
  /** Header text */
  header: string
  /** Column width (CSS value) */
  width?: string
  /** Custom render function */
  render?: (item: T) => React.ReactNode
  /** Whether column is sortable */
  sortable?: boolean
}

export interface EntityTableProps<T> {
  /** Column definitions */
  columns: Column<T>[]
  /** Data items */
  data: T[]
  /** Loading state */
  loading?: boolean
  /** Pagination props */
  pagination?: PaginationProps
  /** Currently selected item ID */
  selectedId?: string
  /** Called when a row is clicked */
  onSelect?: (item: T) => void
  /** Called when sort changes */
  onSort?: (column: string, direction: 'asc' | 'desc') => void
  /** Current sort column */
  sortColumn?: string
  /** Current sort direction */
  sortDirection?: 'asc' | 'desc'
  /** Function to get item ID */
  getItemId: (item: T) => string
  /** Message shown when data is empty */
  emptyMessage?: string
  /** Custom class name */
  className?: string
}

function SkeletonRow({ columns }: { columns: number }) {
  return (
    <tr>
      {Array.from({ length: columns }).map((_, i) => (
        <td key={i}>
          <div className="placeholder-glow">
            <span className="placeholder col-8"></span>
          </div>
        </td>
      ))}
    </tr>
  )
}

export function EntityTable<T>({
  columns,
  data,
  loading = false,
  pagination,
  selectedId,
  onSelect,
  onSort,
  sortColumn,
  sortDirection = 'asc',
  getItemId,
  emptyMessage = 'No items found',
  className = '',
}: EntityTableProps<T>) {
  const handleHeaderClick = (column: Column<T>) => {
    if (!column.sortable || !onSort) return

    const newDirection =
      sortColumn === column.key && sortDirection === 'asc' ? 'desc' : 'asc'
    onSort(column.key, newDirection)
  }

  const renderSortIndicator = (column: Column<T>) => {
    if (!column.sortable) return null
    if (sortColumn !== column.key) {
      return <span className="text-muted opacity-25 ms-1">↕</span>
    }
    return (
      <span className="ms-1">{sortDirection === 'asc' ? '↑' : '↓'}</span>
    )
  }

  return (
    <div className={`entity-table ${className}`}>
      <div className="table-responsive">
        <table className="table table-hover table-sm mb-0">
          <thead className="table-light sticky-top">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.key}
                  style={{ width: column.width }}
                  className={column.sortable ? 'cursor-pointer user-select-none' : ''}
                  onClick={() => handleHeaderClick(column)}
                >
                  {column.header}
                  {renderSortIndicator(column)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              // Skeleton loading rows
              Array.from({ length: 5 }).map((_, i) => (
                <SkeletonRow key={i} columns={columns.length} />
              ))
            ) : data.length === 0 ? (
              // Empty state
              <tr>
                <td colSpan={columns.length} className="text-center text-muted py-4">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              // Data rows
              data.map((item) => {
                const itemId = getItemId(item)
                const isSelected = itemId === selectedId
                return (
                  <tr
                    key={itemId}
                    className={`${isSelected ? 'table-primary' : ''} ${onSelect ? 'cursor-pointer' : ''}`}
                    onClick={() => onSelect?.(item)}
                  >
                    {columns.map((column) => (
                      <td key={column.key}>
                        {column.render
                          ? column.render(item)
                          : String((item as Record<string, unknown>)[column.key] ?? '')}
                      </td>
                    ))}
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      {pagination && (
        <div className="p-2 border-top bg-body-tertiary">
          <Pagination {...pagination} />
        </div>
      )}

      <style>{`
        .cursor-pointer {
          cursor: pointer;
        }
        .entity-table .table {
          margin-bottom: 0;
        }
        .entity-table thead {
          position: sticky;
          top: 0;
          z-index: 1;
        }
      `}</style>
    </div>
  )
}
