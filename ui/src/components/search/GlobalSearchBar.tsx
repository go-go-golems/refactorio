import { useState, useCallback, useRef, useEffect } from 'react'

export interface GlobalSearchBarProps {
  /** Current search value */
  value?: string
  /** Called on input change */
  onChange?: (value: string) => void
  /** Called on Enter or submit */
  onSubmit?: (value: string) => void
  /** Placeholder text */
  placeholder?: string
  /** Suggestion items */
  suggestions?: string[]
  /** Whether a search is in progress */
  loading?: boolean
  /** Auto-focus the input */
  autoFocus?: boolean
}

const SearchIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
    <path d="M11.742 10.344a6.5 6.5 0 10-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 001.415-1.414l-3.85-3.85a1.007 1.007 0 00-.115-.1zM12 6.5a5.5 5.5 0 11-11 0 5.5 5.5 0 0111 0z" />
  </svg>
)

const ClearIcon = () => (
  <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
    <path d="M4.646 4.646a.5.5 0 01.708 0L8 7.293l2.646-2.647a.5.5 0 01.708.708L8.707 8l2.647 2.646a.5.5 0 01-.708.708L8 8.707l-2.646 2.647a.5.5 0 01-.708-.708L7.293 8 4.646 5.354a.5.5 0 010-.708z" />
  </svg>
)

export function GlobalSearchBar({
  value: controlledValue,
  onChange,
  onSubmit,
  placeholder = 'Search symbols, code, commits, files\u2026',
  suggestions,
  loading = false,
  autoFocus = false,
}: GlobalSearchBarProps) {
  const [internalValue, setInternalValue] = useState('')
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const value = controlledValue ?? internalValue

  const handleChange = useCallback(
    (v: string) => {
      if (controlledValue === undefined) setInternalValue(v)
      onChange?.(v)
      setShowSuggestions(v.length > 0 && (suggestions?.length ?? 0) > 0)
      setSelectedIndex(-1)
    },
    [controlledValue, onChange, suggestions],
  )

  const handleSubmit = useCallback(
    (v: string) => {
      onSubmit?.(v)
      setShowSuggestions(false)
    },
    [onSubmit],
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!suggestions || suggestions.length === 0) {
        if (e.key === 'Enter') handleSubmit(value)
        return
      }

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex((prev) => Math.min(prev + 1, suggestions.length - 1))
          break
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex((prev) => Math.max(prev - 1, -1))
          break
        case 'Enter':
          e.preventDefault()
          if (selectedIndex >= 0) {
            handleChange(suggestions[selectedIndex])
            handleSubmit(suggestions[selectedIndex])
          } else {
            handleSubmit(value)
          }
          break
        case 'Escape':
          setShowSuggestions(false)
          setSelectedIndex(-1)
          break
      }
    },
    [suggestions, selectedIndex, value, handleChange, handleSubmit],
  )

  const handleClear = useCallback(() => {
    handleChange('')
    inputRef.current?.focus()
  }, [handleChange])

  // Close suggestions on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowSuggestions(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const filteredSuggestions = suggestions?.filter((s) =>
    s.toLowerCase().includes(value.toLowerCase()),
  )

  return (
    <div ref={containerRef} className="position-relative" style={{ maxWidth: 480 }}>
      <div className="input-group input-group-sm">
        <span className="input-group-text bg-transparent border-end-0">
          {loading ? (
            <span className="spinner-border spinner-border-sm" role="status" aria-label="Searching" />
          ) : (
            <SearchIcon />
          )}
        </span>
        <input
          ref={inputRef}
          type="search"
          className="form-control border-start-0"
          placeholder={placeholder}
          value={value}
          onChange={(e) => handleChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => value.length > 0 && (filteredSuggestions?.length ?? 0) > 0 && setShowSuggestions(true)}
          autoFocus={autoFocus}
          aria-label="Search"
          aria-autocomplete={suggestions ? 'list' : undefined}
          aria-expanded={showSuggestions}
        />
        {value.length > 0 && (
          <button
            type="button"
            className="btn btn-outline-secondary border-start-0"
            onClick={handleClear}
            aria-label="Clear search"
          >
            <ClearIcon />
          </button>
        )}
      </div>

      {showSuggestions && filteredSuggestions && filteredSuggestions.length > 0 && (
        <ul
          className="list-group position-absolute w-100 shadow-sm"
          style={{ zIndex: 1050, maxHeight: 240, overflowY: 'auto', top: '100%' }}
          role="listbox"
        >
          {filteredSuggestions.map((suggestion, i) => (
            <button
              key={suggestion}
              type="button"
              className={`list-group-item list-group-item-action py-1 ${i === selectedIndex ? 'active' : ''}`}
              style={{ fontSize: '0.85rem' }}
              role="option"
              aria-selected={i === selectedIndex}
              onClick={() => {
                handleChange(suggestion)
                handleSubmit(suggestion)
              }}
              onMouseEnter={() => setSelectedIndex(i)}
            >
              {suggestion}
            </button>
          ))}
        </ul>
      )}
    </div>
  )
}
