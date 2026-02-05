export interface FilterOption {
  value: string
  label: string
}

export interface FilterConfig {
  key: string
  label: string
  type: 'checkbox-group' | 'select' | 'text' | 'date-range'
  options?: FilterOption[]
  placeholder?: string
}

export interface FilterPanelProps {
  /** Filter definitions */
  filters: FilterConfig[]
  /** Current filter values */
  values: Record<string, unknown>
  /** Called when a filter value changes */
  onChange: (key: string, value: unknown) => void
  /** Called to reset all filters */
  onReset?: () => void
}

function CheckboxGroupFilter({
  config,
  value,
  onChange,
}: {
  config: FilterConfig
  value: string[]
  onChange: (value: string[]) => void
}) {
  const handleToggle = (optValue: string) => {
    if (value.includes(optValue)) {
      onChange(value.filter((v) => v !== optValue))
    } else {
      onChange([...value, optValue])
    }
  }

  return (
    <div>
      {config.options?.map((opt) => (
        <div className="form-check" key={opt.value}>
          <input
            className="form-check-input"
            type="checkbox"
            id={`filter-${config.key}-${opt.value}`}
            checked={value.includes(opt.value)}
            onChange={() => handleToggle(opt.value)}
          />
          <label className="form-check-label small" htmlFor={`filter-${config.key}-${opt.value}`}>
            {opt.label}
          </label>
        </div>
      ))}
    </div>
  )
}

function SelectFilter({
  config,
  value,
  onChange,
}: {
  config: FilterConfig
  value: string
  onChange: (value: string) => void
}) {
  return (
    <select
      className="form-select form-select-sm"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      <option value="">All</option>
      {config.options?.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  )
}

function TextFilter({
  config,
  value,
  onChange,
}: {
  config: FilterConfig
  value: string
  onChange: (value: string) => void
}) {
  return (
    <input
      type="text"
      className="form-control form-control-sm"
      placeholder={config.placeholder ?? `Filter by ${config.label.toLowerCase()}\u2026`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  )
}

function DateRangeFilter({
  config,
  value,
  onChange,
}: {
  config: FilterConfig
  value: { from?: string; to?: string }
  onChange: (value: { from?: string; to?: string }) => void
}) {
  return (
    <div className="d-flex gap-1 align-items-center">
      <input
        type="date"
        className="form-control form-control-sm"
        value={value.from ?? ''}
        onChange={(e) => onChange({ ...value, from: e.target.value || undefined })}
        aria-label={`${config.label} from date`}
      />
      <span className="text-muted small">&ndash;</span>
      <input
        type="date"
        className="form-control form-control-sm"
        value={value.to ?? ''}
        onChange={(e) => onChange({ ...value, to: e.target.value || undefined })}
        aria-label={`${config.label} to date`}
      />
    </div>
  )
}

function hasActiveFilters(values: Record<string, unknown>): boolean {
  return Object.values(values).some((v) => {
    if (Array.isArray(v)) return v.length > 0
    if (typeof v === 'object' && v !== null) {
      return Object.values(v as Record<string, unknown>).some(Boolean)
    }
    return v !== '' && v !== undefined && v !== null
  })
}

export function FilterPanel({ filters, values, onChange, onReset }: FilterPanelProps) {
  return (
    <div className="filter-panel">
      <div className="d-flex justify-content-between align-items-center mb-2">
        <span className="fw-semibold small">Filters</span>
        {onReset && hasActiveFilters(values) && (
          <button
            type="button"
            className="btn btn-link btn-sm p-0 text-decoration-none"
            onClick={onReset}
          >
            Reset
          </button>
        )}
      </div>

      {filters.map((config) => (
        <div key={config.key} className="mb-3">
          <label className="form-label small text-muted mb-1">{config.label}</label>
          {config.type === 'checkbox-group' && (
            <CheckboxGroupFilter
              config={config}
              value={(values[config.key] as string[]) ?? []}
              onChange={(v) => onChange(config.key, v)}
            />
          )}
          {config.type === 'select' && (
            <SelectFilter
              config={config}
              value={(values[config.key] as string) ?? ''}
              onChange={(v) => onChange(config.key, v)}
            />
          )}
          {config.type === 'text' && (
            <TextFilter
              config={config}
              value={(values[config.key] as string) ?? ''}
              onChange={(v) => onChange(config.key, v)}
            />
          )}
          {config.type === 'date-range' && (
            <DateRangeFilter
              config={config}
              value={(values[config.key] as { from?: string; to?: string }) ?? {}}
              onChange={(v) => onChange(config.key, v)}
            />
          )}
        </div>
      ))}
    </div>
  )
}
